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

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/osbutil"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/secrets"
	cluster "github.com/google/kf/v2/pkg/kf/service-brokers/cluster"
	namespaced "github.com/google/kf/v2/pkg/kf/service-brokers/namespaced"
	"github.com/spf13/cobra"
	"knative.dev/pkg/kmeta"
)

var (
	provisionTimeout = 1 * time.Minute
)

// NewCreateServiceBrokerCommand adds a service broker (either cluster or namespaced) to the service catalog.
func NewCreateServiceBrokerCommand(
	p *config.KfParams,
	clusterClient cluster.Client,
	namespacedClient namespaced.Client,
	secretsClient secrets.Client,
) *cobra.Command {

	var (
		spaceScoped bool
		async       utils.AsyncFlags
	)

	createCmd := &cobra.Command{
		Use:          "create-service-broker NAME USERNAME PASSWORD URL",
		Aliases:      []string{"csb"},
		Short:        "Add a service broker to the marketplace.",
		Example:      `  kf create-service-broker mybroker user pass http://mybroker.broker.svc.cluster.local`,
		Args:         cobra.ExactArgs(4),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceBrokerName := args[0]
			username := args[1]
			password := args[2]
			url := args[3]

			// TODO: validate URL

			if spaceScoped {
				if err := p.ValidateSpaceTargeted(); err != nil {
					return err
				}
			}

			var callback func(context.Context) error
			var provisionedBroker kmeta.OwnerRefable

			brokerSecretName := v1alpha1.GenerateName(serviceBrokerName, "auth")

			switch {
			case spaceScoped:
				desiredBroker := populateV1alpha1SpaceBrokerTemplate(p.Space, serviceBrokerName, brokerSecretName)
				actualBroker, err := namespacedClient.Create(cmd.Context(), p.Space, desiredBroker)
				if err != nil {
					return err
				}
				provisionedBroker = actualBroker
				callback = func(ctx context.Context) error {
					_, err := namespacedClient.WaitForConditionReadyTrue(ctx, p.Space, serviceBrokerName, 1*time.Second)
					return err
				}

			default:
				desiredBroker := populateV1alpha1ClusterBrokerTemplate(serviceBrokerName, brokerSecretName)
				actualBroker, err := clusterClient.Create(cmd.Context(), desiredBroker)
				if err != nil {
					return err
				}
				provisionedBroker = actualBroker
				callback = func(ctx context.Context) error {
					_, err := clusterClient.WaitForConditionReadyTrue(ctx, serviceBrokerName, 1*time.Second)
					return err
				}
			}

			// Create secret
			secret := osbutil.NewBasicAuthSecret(brokerSecretName, username, password, url, provisionedBroker)
			if _, err := secretsClient.Create(cmd.Context(), secret.Namespace, secret); err != nil {
				return err
			}

			// Set up messages for the user
			var action, deleteAction string
			if spaceScoped {
				action = fmt.Sprintf("Creating service broker %q in Space %q", serviceBrokerName, p.Space)
				deleteAction = fmt.Sprintf("kf delete-service-broker --space %s --space-scoped %s", p.Space, serviceBrokerName)
			} else {
				action = fmt.Sprintf("Creating cluster service broker %q", serviceBrokerName)
				deleteAction = fmt.Sprintf("kf delete-service-broker %s", serviceBrokerName)
			}

			// Wait for provision
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				ctx, cancel := context.WithTimeout(context.Background(), provisionTimeout)
				defer cancel()

				err := callback(ctx)

				if err != nil {
					w := cmd.OutOrStdout()
					fmt.Fprintln(w, "Waiting failed, check your URL, credentials and the broker status.")
					fmt.Fprintln(w, utils.Warnf("NOTE: The broker HAS NOT been deleted."))
					fmt.Fprintln(w, "You can delete the broker with:")
					fmt.Fprintf(w, "  %s\n", deleteAction)
				}

				return err
			})
		},
	}

	async.Add(createCmd)

	createCmd.Flags().BoolVar(
		&spaceScoped,
		"space-scoped",
		false,
		"Only create the broker in the targeted space.")

	return createCmd
}

// populateV1alpha1SpaceBrokerTemplate fills in a broker template that can be used
// to generate Kf space scoped service-brokers.
func populateV1alpha1SpaceBrokerTemplate(namespace, name, secretName string) *v1alpha1.ServiceBroker {
	broker := &v1alpha1.ServiceBroker{}

	broker.Name = name
	broker.Namespace = namespace
	broker.Spec.Credentials.Name = secretName

	return broker
}

// populateV1alpha1ClusterBrokerTemplate fills in a broker template that can be used
// to generate Kf sluster scoped service-brokers.
func populateV1alpha1ClusterBrokerTemplate(name, secretName string) *v1alpha1.ClusterServiceBroker {
	broker := &v1alpha1.ClusterServiceBroker{}

	broker.Name = name
	broker.Spec.Credentials.Name = secretName
	broker.Spec.Credentials.Namespace = v1alpha1.KfNamespace

	return broker
}
