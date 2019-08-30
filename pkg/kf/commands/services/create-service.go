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
	"encoding/json"
	"fmt"

	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/services"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewCreateServiceCommand allows users to create service instances.
func NewCreateServiceCommand(p *config.KfParams, client servicecatalogclient.Interface) *cobra.Command {
	var (
		configAsJSON string
		broker       string
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

			params, err := services.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}
			paramBytes, err := json.Marshal(params)
			if err != nil {
				return err
			}
			rawParams := &runtime.RawExtension{
				Raw: paramBytes,
			}

			matchingClusterPlans, err := findMatchingClusterPlans(client, planName, serviceName, broker)
			if err != nil {
				return err
			}

			if len(matchingClusterPlans) != 0 {

				// plan found
				created, err := client.ServicecatalogV1beta1().
					ServiceInstances(p.Namespace).
					Create(&servicecatalogv1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name:      instanceName,
							Namespace: p.Namespace,
						},
						Spec: servicecatalogv1beta1.ServiceInstanceSpec{
							PlanReference: servicecatalogv1beta1.PlanReference{
								ClusterServicePlanExternalName:  planName,
								ClusterServiceClassExternalName: serviceName,
							},
							Parameters: rawParams,
						},
					})
				if err != nil {
					return err
				}

				describe.ServiceInstance(cmd.OutOrStdout(), created)
				return nil
			}

			matchingNamespacedPlans, err := findMatchingNamespacedPlans(client, p.Namespace, planName, serviceName, broker)
			if err != nil {
				return err
			}

			if len(matchingNamespacedPlans) != 0 {
				// plan found
				created, err := client.ServicecatalogV1beta1().
					ServiceInstances(p.Namespace).
					Create(&servicecatalogv1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name:      instanceName,
							Namespace: p.Namespace,
						},
						Spec: servicecatalogv1beta1.ServiceInstanceSpec{
							PlanReference: servicecatalogv1beta1.PlanReference{
								ServicePlanExternalName:  planName,
								ServiceClassExternalName: serviceName,
							},
							Parameters: rawParams,
						},
					})
				if err != nil {
					return err
				}

				describe.ServiceInstance(cmd.OutOrStdout(), created)
				return nil
			}

			if broker != "" {
				return fmt.Errorf("no plan %s found for class %s for the service-broker %s", planName, serviceName, broker)
			} else {
				return fmt.Errorf("no plan %s found for class %s for all service-brokers", planName, serviceName)
			}
		},
	}

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

func findMatchingClusterPlans(client servicecatalogclient.Interface, planName, serviceName, broker string) ([]servicecatalogv1beta1.ClusterServicePlan, error) {

	var matchingPlans []servicecatalogv1beta1.ClusterServicePlan

	plans, err := client.ServicecatalogV1beta1().
		ClusterServicePlans().
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, plan := range plans.Items {
		if planName != plan.Spec.ExternalName {
			continue
		}

		class, err := client.ServicecatalogV1beta1().
			ClusterServiceClasses().
			Get(plan.Spec.ClusterServiceClassRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if serviceName != class.Spec.ExternalName {
			continue
		}

		if broker != "" && broker != plan.Spec.ClusterServiceBrokerName {
			continue
		}

		matchingPlans = append(matchingPlans, plan)
	}

	return matchingPlans, nil
}

func findMatchingNamespacedPlans(client servicecatalogclient.Interface, namespace, planName, serviceName, broker string) ([]servicecatalogv1beta1.ServicePlan, error) {

	var matchingPlans []servicecatalogv1beta1.ServicePlan

	plans, err := client.ServicecatalogV1beta1().
		ServicePlans(namespace).
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, plan := range plans.Items {
		if planName != plan.Spec.ExternalName {
			continue
		}

		class, err := client.ServicecatalogV1beta1().
			ServiceClasses(namespace).
			Get(plan.Spec.ServiceClassRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if serviceName != class.Spec.ExternalName {
			continue
		}

		if broker != "" && broker != plan.Spec.ServiceBrokerName {
			continue
		}

		matchingPlans = append(matchingPlans, plan)
	}

	return matchingPlans, nil
}
