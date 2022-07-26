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
	"context"
	"fmt"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewUnbindServiceCommand allows users to unbind apps from service instances.
func NewUnbindServiceCommand(p *config.KfParams, client serviceinstancebindings.Client) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:     "unbind-service APP_NAME SERVICE_INSTANCE",
		Aliases: []string{"us"},
		Short:   "Revoke an App's access to a service instance.",
		Long: `Unbind removes an App's access to a service instance.

		This will delete the credential from the service broker that created the
		instance and update the VCAP_SERVICES environment variable for the
		App to remove the reference to the instance.
		`,
		Example:           `kf unbind-service myapp my-instance`,
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

			mutator := func(b *v1alpha1.ServiceInstanceBinding) error {
				b.Spec.UnbindRequests++

				return nil
			}

			if _, err := client.Transform(cmd.Context(), p.Space, bindingName, mutator); err != nil {
				return fmt.Errorf("Failed to update unbinding requests: %s", err)
			}

			if err := client.Delete(ctx, p.Space, bindingName); err != nil {
				return err
			}

			action := fmt.Sprintf("Deleting service instance binding in Space %q", p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForDeletion(context.Background(), p.Space, bindingName, 1*time.Second)
				if err != nil {
					return fmt.Errorf("unbind failed: %s", err)
				}
				logging.FromContext(ctx).Infof("Use 'kf restart %s' to ensure your changes take effect", appName)
				return nil
			})
		},
	}

	async.Add(cmd)

	return cmd
}
