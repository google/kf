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
	"errors"
	"fmt"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCreateUserProvidedServiceCommand allows users to create user-provided service instances.
func NewCreateUserProvidedServiceCommand(p *config.KfParams, client serviceinstances.Client, secretsClient secrets.Client) *cobra.Command {
	var (
		configAsJSON   string
		syslogDrainURL string
		routeURL       string
		tags           string
		async          utils.AsyncFlags
		mockClassName  string
		mockPlanName   string
	)

	cmd := &cobra.Command{
		Use:     "create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-t TAGS]",
		Aliases: []string{"cups"},
		Short:   "Create a standalone service instance from existing credentials.",
		Long: `
		Creates a standalone service instance from existing credentials.
		User-provided services can be used to inject credentials for services managed
		outside of Kf into Apps.

		Credentials are stored in a Kubernetes Secret in the Space the service is
		created in. On GKE these Secrets are encrypted at rest and can optionally
		be encrypted using KMS.
		`,
		Example: `
		# Bring an existing database service
		kf create-user-provided-service db-service -p '{"url":"mysql://..."}'

		# Create a service with tags for autowiring
		kf create-user-provided-service db-service -t "mysql,database,sql"
		`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			instanceName := args[0]

			// Determine whether the route service feature flag is enabled.
			routeServicesEnabled := p.FeatureFlags(ctx).RouteServices().IsEnabled()

			// Explicitly call out these parameters which are valid in CF.
			if syslogDrainURL != "" {
				return errors.New("Kf doesn't currently support syslog drains")
			}

			if !routeServicesEnabled && routeURL != "" {
				return errors.New(`Route services feature is toggled off. Set "enable_route_services" to true in "config-defaults" to enable route services`)
			}

			parsedURL, err := getParsedURL(routeServicesEnabled, routeURL)
			if err != nil {
				return err
			}

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			paramBytes, err := utils.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			userTags := utils.SplitTags(tags)

			paramsSecretName := v1alpha1.GenerateName("serviceinstance", instanceName, "params")
			desiredInstance := &v1alpha1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceName,
					Namespace: p.Space,
				},
				Spec: v1alpha1.ServiceInstanceSpec{
					ServiceType: v1alpha1.ServiceType{
						UPS: &v1alpha1.UPSInstance{
							MockPlanName:    mockPlanName,
							MockClassName:   mockClassName,
							RouteServiceURL: parsedURL,
						},
					},
					Tags: userTags,
					ParametersFrom: corev1.LocalObjectReference{
						Name: paramsSecretName,
					},
				},
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Creating service instance %q in Space %q\n", instanceName, p.Space)

			actualInstance, err := client.Create(ctx, p.Space, desiredInstance)
			if err != nil {
				return err
			}

			if _, err := secretsClient.CreateParamsSecret(ctx, actualInstance, paramsSecretName, paramBytes); err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for service instance to become ready", func() (err error) {
				_, err = client.WaitForConditionReadyTrue(context.Background(), p.Space, instanceName, 1*time.Second)
				return
			})
		},
	}

	async.Add(cmd)

	cmd.Flags().StringVar(
		&configAsJSON,
		"params",
		"{}",
		"JSON object or path to a JSON file containing configuration parameters. DEPRECATED: use --parameters instead.")

	cmd.Flags().StringVarP(
		&configAsJSON,
		"parameters",
		"p",
		"{}",
		"JSON object or path to a JSON file containing configuration parameters.")

	cmd.Flags().StringVarP(
		&tags,
		"tags",
		"t",
		"",
		"User-defined tags to differentiate services during injection.")

	cmd.Flags().StringVar(
		&mockClassName,
		"mock-class",
		"",
		"Mock class name to use in VCAP_SERVICES rather than 'user-provided'.")

	cmd.Flags().StringVar(
		&mockPlanName,
		"mock-plan",
		"",
		"Mock plan name to use in VCAP_SERVICES rather than blank.")

	cmd.Flags().StringVarP(
		&routeURL,
		"route",
		"r",
		"",
		"URL to which requests for bound routes will be forwarded. Scheme must be https. NOTE: This is a preivew feature.")

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
	}

	return cmd
}

// getParsedURL returns nil if the UPS is not a route service (or if route services are disabled).
// Otherwise, it returns the net/url parsing of the route service URL string.
func getParsedURL(routeServicesEnabled bool, routeURL string) (*v1alpha1.RouteServiceURL, error) {
	if !routeServicesEnabled || routeURL == "" {
		return nil, nil
	}
	parsedURL, err := v1alpha1.ParseURL(routeURL)
	if err != nil {
		return nil, err
	}
	return parsedURL, err
}
