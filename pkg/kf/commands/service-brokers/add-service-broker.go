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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	servicebrokers "github.com/google/kf/pkg/kf/service-brokers"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

// NewAddServiceBrokerCommand adds a namespaced service broker to the service catalog.
func NewAddServiceBrokerCommand(p *config.KfParams, client servicebrokers.Client) *cobra.Command {
	var (
		serviceBrokerName string
		url               string
	)

	createCmd := &cobra.Command{
		Use:     "add-service-broker BROKER_NAME URL",
		Aliases: []string{"asb"},
		Short:   "Add a namespaced service broker to service catalog",
		Example: `  kf add-service-broker mybroker http://mybroker.broker.svc.cluster.local`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceBrokerName = args[0]
			url = args[1]

			cmd.SilenceUsage = true

			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			// TODO (juliaguo): validate URL

			desiredBroker := &servicecatalogv1beta1.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceBrokerName,
					Namespace: p.Namespace,
				},
				Spec: servicecatalogv1beta1.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalogv1beta1.CommonServiceBrokerSpec{
						URL: url,
					},
				},
			}

			_, err := client.Create(p.Namespace, desiredBroker)

			if err != nil {
				return err
			}

			return nil
		},
	}

	return createCmd
}
