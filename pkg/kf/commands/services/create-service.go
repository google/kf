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
	"time"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/marketplace"
	"github.com/google/kf/pkg/kf/services"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewCreateServiceCommand allows users to create service instances.
func NewCreateServiceCommand(p *config.KfParams, client services.Client, marketplaceClient marketplace.ClientInterface) *cobra.Command {
	var (
		configAsJSON string
		broker       string
		async        utils.AsyncFlags
	)

	createCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-b service-broker]",
		Aliases: []string{"cs"},
		Short:   "Create a service instance",
		Example: `
  # Creates a new instance of a db-service with the name mydb, plan silver, and provisioning configuration
  kf create-service db-service silver mydb -c '{"ram_gb":4}'

  # Creates a new instance of a db-service from the broker named local-broker
  kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json -b local-broker`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			planName := args[1]
			instanceName := args[2]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			paramBytes, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			planFilters := marketplace.ListPlanOptions{
				PlanName:    planName,
				ServiceName: serviceName,
				BrokerName:  broker,
			}

			matchingClusterPlans, err := marketplaceClient.ListClusterPlans(planFilters)
			if err != nil {
				return err
			}
			hasClusterPlans := len(matchingClusterPlans) > 0

			matchingNamespacedPlans, err := marketplaceClient.ListNamespacedPlans(p.Namespace, planFilters)
			if err != nil {
				return err
			}
			hasNamespacedPlans := len(matchingNamespacedPlans) > 0

			var planRef servicecatalogv1beta1.PlanReference

			switch {
			case hasClusterPlans && hasNamespacedPlans:
				return errors.New("plans matched from multiple brokers, specify a broker with --broker")

			case hasClusterPlans:
				planRef = servicecatalogv1beta1.PlanReference{
					ClusterServicePlanExternalName:  planName,
					ClusterServiceClassExternalName: serviceName,
				}

			case hasNamespacedPlans:
				planRef = servicecatalogv1beta1.PlanReference{
					ServicePlanExternalName:  planName,
					ServiceClassExternalName: serviceName,
				}

			// No plans match
			case broker != "":
				return fmt.Errorf("no plan %s found for class %s for the service-broker %s", planName, serviceName, broker)
			default:
				return fmt.Errorf("no plan %s found for class %s for all service-brokers", planName, serviceName)
			}

			created, err := client.Create(p.Namespace, &servicecatalogv1beta1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceName,
					Namespace: p.Namespace,
				},
				Spec: servicecatalogv1beta1.ServiceInstanceSpec{
					PlanReference: planRef,
					Parameters: &runtime.RawExtension{
						Raw: paramBytes,
					},
				},
			})
			if err != nil {
				return err
			}

			action := fmt.Sprintf("Creating service instance %q in space %q", instanceName, p.Namespace)
			if err := async.AwaitAndLog(cmd.OutOrStdout(), action, func() (err error) {
				created, err = client.WaitForProvisionSuccess(context.Background(), p.Namespace, instanceName, 1*time.Second)
				return
			}); err != nil {
				return err
			}

			describe.ServiceInstance(cmd.OutOrStdout(), created)
			return nil
		},
	}

	async.Add(createCmd)

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"config",
		"c",
		"{}",
		"Valid JSON object containing service-specific configuration parameters, provided in-line or in a file.")

	createCmd.Flags().StringVarP(
		&broker,
		"broker",
		"b",
		"",
		"Service broker to use.")

	return createCmd
}
