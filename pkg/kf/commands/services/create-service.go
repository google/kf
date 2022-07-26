// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/marketplace"
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	logging "knative.dev/pkg/logging"
)

// NewCreateServiceCommand allows users to create service instances.
func NewCreateServiceCommand(p *config.KfParams, client serviceinstances.Client, secretsClient secrets.Client, marketplaceClient marketplace.ClientInterface) *cobra.Command {
	var (
		configAsJSON string
		broker       string
		tags         string
		async        utils.AsyncFlags
		timeout      time.Duration
	)

	createCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-b service-broker] [-t TAGS]",
		Aliases: []string{"cs"},
		Short:   "Create a service instance from a marketplace template.",
		Long: `
		Create service creates a new ServiceInstance using a template from the
		marketplace.
		`,
		Example: `
		# Creates a new instance of a db-service with the name mydb, plan silver, and provisioning configuration
		kf create-service db-service silver mydb -c '{"ram_gb":4}'

		# Creates a new instance of a db-service from the broker named local-broker
		kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json -b local-broker

		# Creates a new instance of a db-service with the name mydb and override tags
		kf create-service db-service silver mydb -t "list, of, tags"`,
		Args:         cobra.ExactArgs(3),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			serviceName := args[0]
			planName := args[1]
			instanceName := args[2]

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			paramBytes, err := utils.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			catalog, err := marketplaceClient.Marketplace(cmd.Context(), p.Space)
			if err != nil {
				return err
			}

			planFilters := marketplace.ListPlanOptions{
				PlanName:    planName,
				ServiceName: serviceName,
				BrokerName:  broker,
			}

			matchingClusterPlans := catalog.ListClusterPlans(planFilters)
			matchingNamespacedPlans := catalog.ListNamespacedPlans(p.Space, planFilters)

			var namespaceScoped bool
			var lineage marketplace.PlanLineage

			switch {
			case len(matchingClusterPlans)+len(matchingNamespacedPlans) > 1:
				return errors.New("plans matched from multiple brokers, specify a broker with --broker")
			case len(matchingClusterPlans) == 1:
				namespaceScoped = false
				lineage = matchingClusterPlans[0]
			case len(matchingNamespacedPlans) == 1:
				namespaceScoped = true
				lineage = matchingNamespacedPlans[0]
			case broker != "":
				return fmt.Errorf("no plan %s found for class %s for the service-broker %s", planName, serviceName, broker)
			default:
				return fmt.Errorf("no plan %s found for class %s for all service-brokers", planName, serviceName)
			}

			tagSet := sets.NewString(lineage.ServiceOffering.Tags...)
			tagSet.Insert(utils.SplitTags(tags)...)
			mergedTags := tagSet.List()

			osbInstance := &v1alpha1.OSBInstance{
				BrokerName:              lineage.Broker.GetName(),
				ClassName:               lineage.ServiceOffering.DisplayName,
				ClassUID:                lineage.ServiceOffering.UID,
				PlanName:                lineage.ServicePlan.DisplayName,
				PlanUID:                 lineage.ServicePlan.UID,
				Namespaced:              namespaceScoped,
				ProgressDeadlineSeconds: int64(timeout / time.Second),
			}

			var serviceType v1alpha1.ServiceType
			if lineage.Broker.GetKind() == v1alpha1.VolumeBrokerKind {
				serviceType = v1alpha1.ServiceType{
					Volume: osbInstance,
				}
			} else {
				serviceType = v1alpha1.ServiceType{
					OSB: osbInstance,
				}
			}

			paramsSecretName := v1alpha1.GenerateName("serviceinstance", instanceName, "params")
			desiredInstance := &v1alpha1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceName,
					Namespace: p.Space,
				},
				Spec: v1alpha1.ServiceInstanceSpec{
					ServiceType: serviceType,
					Tags:        mergedTags,
					ParametersFrom: corev1.LocalObjectReference{
						Name: paramsSecretName,
					},
				},
			}

			logger := logging.FromContext(ctx)
			logger.Infof("Creating ServiceInstance %q in Space %q\n", instanceName, p.Space)

			describe.SectionWriter(cmd.ErrOrStderr(), "ServiceInstance Parameters", func(w io.Writer) {
				if err := describe.UnstructuredStruct(w, desiredInstance.Spec); err != nil {
					fmt.Fprintln(w, err.Error())
				}
			})

			actualInstance, err := client.Create(ctx, p.Space, desiredInstance)
			if err != nil {
				return err
			}

			logger.Infof("Creating parameters Secret %q in Space %q\n", paramsSecretName, p.Space)
			if _, err := secretsClient.CreateParamsSecret(ctx, actualInstance, paramsSecretName, paramBytes); err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for ServiceInstance to become ready", func() (err error) {
				_, err = client.WaitForConditionReadyTrue(context.Background(), p.Space, instanceName, 1*time.Second)
				return
			})
		},
	}

	async.Add(createCmd)

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"parameters",
		"c",
		"{}",
		"JSON object or path to a JSON file containing configuration parameters.")

	createCmd.Flags().StringVarP(
		&broker,
		"broker",
		"b",
		"",
		"Name of the service broker that will create the instance.")

	createCmd.Flags().StringVarP(
		&tags,
		"tags",
		"t",
		"",
		"User-defined tags to differentiate services during injection.")

	createCmd.Flags().DurationVar(
		&timeout,
		"timeout",
		time.Duration(v1alpha1.DefaultServiceInstanceProgressDeadlineSeconds)*time.Second,
		`Amount of time to wait for the operation to complete. Valid units are "s", "m", "h".`,
	)

	return createCmd
}
