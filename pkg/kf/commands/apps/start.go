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

package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/spf13/cobra"
)

// NewStartCommand creates a command capable of starting an app.
func NewStartCommand(
	p *config.KfParams,
	client apps.Client,
) *cobra.Command {
	var async utils.AsyncFlags

	cmd := &cobra.Command{
		Use:   "start APP_NAME",
		Short: "Deploy a stopped App and route traffic to it once healthy.",
		Long: `
		Starting an App will create a Kubernetes Deployment and Service.
		The Deployment will use the supplied health checks to validate liveness.
		While the Deployment is sclaing up, Routes will be modified to send traffic
		to live instances of the App.
		`,
		Example:           `kf start myapp`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			mutator := func(app *v1alpha1.App) error {
				app.Spec.Instances.Stopped = false
				return nil
			}

			if _, err := client.Transform(cmd.Context(), p.Space, appName, mutator); err != nil {
				return fmt.Errorf("failed to start app: %s", err)
			}

			action := fmt.Sprintf("Starting app %q in space %q", appName, p.Space)
			return async.AwaitAndLog(cmd.OutOrStdout(), action, func() error {
				_, err := client.WaitForConditionKnativeServiceReadyTrue(context.Background(), p.Space, appName, 1*time.Second)
				return err
			})
		},
	}

	async.Add(cmd)

	return cmd
}
