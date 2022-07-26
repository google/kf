// Copyright 2020 Google LLC
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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/spf13/cobra"
)

// NewUpdateUserProvidedServiceCommand allows users to update user-provided service instances.
// New credentials overwrite old credentials, and parameters not provided are deleted.
// The updated service instance credentials are reflected on the application after unbinding and binding the service.
// See https://docs.cloudfoundry.org/devguide/services/user-provided.html for more information.
func NewUpdateUserProvidedServiceCommand(p *config.KfParams, client serviceinstances.Client, secretsClient secrets.Client) *cobra.Command {
	var (
		configAsJSON   string
		syslogDrainURL string
		routeURL       string
		tags           string
		async          utils.AsyncFlags
	)

	cmd := &cobra.Command{
		Use:     "update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-t TAGS]",
		Aliases: []string{"uups"},
		Short:   "Update a standalone service instance with new credentials.",
		Long: `
		Updates the credentials stored in the Kubernetes Secret for a user-provided service.
		These credentials will be propagated to Apps.

		Apps may need to be restarted to receive the updated credentials.
		`,
		Example: `
		# Update an existing database service
		kf update-user-provided-service db-service -p '{"url":"mysql://..."}'

		# Update a service with tags for autowiring
		kf update-user-provided-service db-service -t "mysql,database,sql"
		`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]

			// Explicitly call out these parameters which are valid in CF.
			if syslogDrainURL != "" {
				return errors.New("Kf doesn't currently support syslog drains")
			}

			if routeURL != "" {
				return errors.New("Kf doesn't currently support user-provided route services")
			}

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			paramBytes, err := utils.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			userTags := utils.SplitTags(tags)

			existingInstance, err := client.Get(cmd.Context(), p.Space, instanceName)
			if err != nil {
				return err
			}
			if !existingInstance.IsUserProvided() {
				return errors.New("Service instance is not user-provided")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updating service instance %q in space %q\n", instanceName, p.Space)

			// Overwrite tags if they are specified, otherwise keep existing tags
			if len(userTags) > 0 {
				_, err = client.Transform(cmd.Context(), p.Space, instanceName, func(serviceinstance *v1alpha1.ServiceInstance) error {
					serviceinstance.Spec.Tags = userTags
					return nil
				})
				if err != nil {
					return err
				}
			}

			// Overwrite credentials if they are specified, otherwise keep existing credentials
			if !reflect.DeepEqual(paramBytes, json.RawMessage("{}")) {
				if _, err := secretsClient.UpdateParamsSecret(cmd.Context(), existingInstance.Namespace, existingInstance.Status.SecretName, paramBytes); err != nil {
					return err
				}
			}

			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for service instance to become ready", func() (err error) {
				_, err = client.WaitForConditionReadyTrue(context.Background(), p.Space, instanceName, 1*time.Second)
				return
			})
		},
	}

	async.Add(cmd)

	cmd.Flags().StringVarP(
		&configAsJSON,
		"params",
		"p",
		"{}",
		"Valid JSON object containing service-specific configuration parameters, provided in-line or in a file.")

	cmd.Flags().StringVarP(
		&tags,
		"tags",
		"t",
		"",
		"Comma-separated tags for the service instance.")

	{
		// The flags in this block are unsupported, but could be called if a
		// customer replaces a cf call with a Kf one. We provide special errors
		// for them instead.
		cmd.Flags().StringVarP(
			&syslogDrainURL,
			"syslog-drain",
			"l",
			"",
			"URL to which logs for bound applications will be streamed.")
		cmd.Flags().MarkHidden("syslog-drain")

		cmd.Flags().StringVarP(
			&routeURL,
			"route",
			"r",
			"",
			"URL to which requests for bound routes will be forwarded. Scheme must be https. NOTE: This is a preivew feature.")
		cmd.Flags().MarkHidden("route")
	}

	return cmd
}
