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

package servicebrokers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/service-brokers/cluster"
	"github.com/google/kf/pkg/kf/service-brokers/namespaced"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

// NewCreateServiceBrokerCommand adds a service broker (either cluster or namespaced) to the service catalog.
// TODO (juliaguo): Add user/pw args to match cf
func NewCreateServiceBrokerCommand(p *config.KfParams, clusterClient cluster.Client, namespacedClient namespaced.Client) *cobra.Command {
	var (
		spaceScoped bool
		async       utils.AsyncFlags
	)

	createCmd := &cobra.Command{
		Use:     "create-service-broker BROKER_NAME URL",
		Aliases: []string{"csb"},
		Short:   "Add a service broker to service catalog",
		Example: `  kf create-service-broker mybroker http://mybroker.broker.svc.cluster.local`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceBrokerName := args[0]
			url := args[1]

			cmd.SilenceUsage = true

			// TODO (juliaguo): validate URL

			if spaceScoped {
				if err := utils.ValidateNamespace(p); err != nil {
					return err
				}

				desiredBroker := PopulateSpaceBrokerTemplate(p.Namespace, serviceBrokerName, url)
				if _, err := namespacedClient.Create(p.Namespace, desiredBroker); err != nil {
					return err
				}

				action := fmt.Sprintf("Creating service broker %q in space %q", serviceBrokerName, p.Namespace)
				return async.AwaitAndLog(cmd.OutOrStdout(), action, func() (err error) {
					_, err = namespacedClient.WaitForConditionReadyTrue(context.Background(), p.Namespace, serviceBrokerName, 1*time.Second)
					return
				})
			}

			desiredBroker := PopulateClusterBrokerTemplate(serviceBrokerName, url)
			if _, err := clusterClient.Create(desiredBroker); err != nil {
				return err
			}

			action := fmt.Sprintf("Creating cluster service broker %q", serviceBrokerName)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() (err error) {
				_, err = clusterClient.WaitForConditionReadyTrue(context.Background(), serviceBrokerName, 1*time.Second)
				return
			})
		},
	}

	async.Add(createCmd)

	createCmd.Flags().BoolVar(
		&spaceScoped,
		"space-scoped",
		false,
		"Set to create a space scoped service broker.")

	return createCmd
}

// PopulateSpaceBrokerTemplate fills in a broker template that can be used
// to generate space scoped service-brokers.
func PopulateSpaceBrokerTemplate(namespace, name, url string) *servicecatalogv1beta1.ServiceBroker {
	return &servicecatalogv1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: servicecatalogv1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: servicecatalogv1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}

// PopulateClusterBrokerTemplate fills in a broker template that can be used
// to generate global service-brokers.
func PopulateClusterBrokerTemplate(name, url string) *servicecatalogv1beta1.ClusterServiceBroker {
	return &servicecatalogv1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: servicecatalogv1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: servicecatalogv1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}
