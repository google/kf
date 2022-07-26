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

package servicebindings

import (
	"fmt"
	"io"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
)

// NewBindServiceCommand allows users to bind apps to service instances.
func NewBindServiceCommand(p *config.KfParams, client serviceinstancebindings.Client, secretsClient secrets.Client, appClient apps.Client) *cobra.Command {
	var (
		bindingOverride string
		configAsJSON    string
		async           utils.AsyncFlags
		timeout         time.Duration
	)

	createCmd := &cobra.Command{
		Use:     "bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME]",
		Aliases: []string{"bs"},
		Short:   "Grant an App access to a service instance.",
		Long: `
		Binding a service injects information about the service into the App via the
		VCAP_SERVICES environment variable.
		`,
		Example:           `  kf bind-service myapp mydb -c '{"permissions":"read-only"}'`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appName := args[0]
			instanceName := args[1]
			bindingName := v1alpha1.MakeServiceBindingName(appName, instanceName)

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			app, err := appClient.Get(ctx, p.Space, appName)
			if err != nil {
				return fmt.Errorf("failed to get App for binding: %s", err)
			}

			paramBytes, err := utils.ParseJSONOrFile(configAsJSON)
			if err != nil {
				return err
			}

			paramsSecretName := v1alpha1.MakeServiceBindingParamsSecretName(appName, instanceName)

			desiredBinding := &v1alpha1.ServiceInstanceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      bindingName,
					Namespace: p.Space,
					OwnerReferences: []metav1.OwnerReference{
						*kmeta.NewControllerRef(app),
					},
				},
				Spec: v1alpha1.ServiceInstanceBindingSpec{
					BindingType: v1alpha1.BindingType{
						App: &v1alpha1.AppRef{
							Name: appName,
						},
					},
					InstanceRef: v1.LocalObjectReference{
						Name: instanceName,
					},
					ParametersFrom: v1.LocalObjectReference{
						Name: paramsSecretName,
					},
					BindingNameOverride:     bindingOverride,
					ProgressDeadlineSeconds: int64(timeout / time.Second),
				},
			}

			logger := logging.FromContext(ctx)
			logger.Infof("Creating ServiceInstanceBinding %q in Space %q", bindingName, p.Space)
			describe.SectionWriter(cmd.ErrOrStderr(), "ServiceInstanceBinding Parameters", func(w io.Writer) {
				if err := describe.UnstructuredStruct(w, desiredBinding.Spec); err != nil {
					fmt.Fprintln(w, err.Error())
				}
			})

			actualBinding, err := client.Create(ctx, p.Space, desiredBinding)
			if err != nil {
				return err
			}

			logger.Infof("Creating parameters Secret %q in Space %q", paramsSecretName, p.Space)
			if _, err := secretsClient.CreateParamsSecret(ctx, actualBinding, paramsSecretName, paramBytes); err != nil {
				return err
			}

			return async.AwaitAndLog(cmd.ErrOrStderr(), "Waiting for ServiceInstanceBinding to become ready", func() (err error) {
				_, err = client.WaitForConditionReadyTrue(ctx, p.Space, bindingName, 1*time.Second)
				if err != nil {
					return fmt.Errorf("bind failed: %s", err)
				}
				logger.Infof("Use 'kf restart %s' to ensure your changes take effect", appName)
				utils.SuggestNextAction(utils.NextAction{
					Description: "List bindings",
					Commands: []string{
						fmt.Sprint("kf bindings"),
					},
				})
				return nil
			})
		},
	}

	createCmd.Flags().StringVarP(
		&configAsJSON,
		"parameters",
		"c",
		"{}",
		"JSON object or path to a JSON file containing configuration parameters.")

	createCmd.Flags().StringVarP(
		&bindingOverride,
		"binding-name",
		"b",
		"",
		"Name of the binding injected into the app, defaults to the service instance name.")

	createCmd.Flags().DurationVar(
		&timeout,
		"timeout",
		time.Duration(v1alpha1.DefaultServiceInstanceBindingProgressDeadlineSeconds)*time.Second,
		`Amount of time to wait for the operation to complete. Valid units are "s", "m", "h".`,
	)

	async.Add(createCmd)

	return createCmd
}
