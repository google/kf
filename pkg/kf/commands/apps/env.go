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
	"fmt"
	"io"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/reconciler/app/resources"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// NewEnvCommand creates a Env command.
func NewEnvCommand(p *config.KfParams, appClient apps.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "env APP_NAME",
		Short:   "Print information about an App's environment variables.",
		Example: `kf env myapp`,
		Args:    cobra.ExactArgs(1),
		Long: heredoc.Doc(`
		The env command gets the names and values of developer managed
		environment variables for an App.

		Environment variables are evaluated in the following order with later values
		overriding earlier ones with the same name:

		1. Space (set by administrators)
		1. App (set by developers)
		1. System (set by Kf)

		Environment variables containing variable substitution "$(...)" are
		replaced at runtime by Kubernetes.
		`) + heredoc.Doc(resources.RuntimeEnvVarDocs(resources.CFRunning)),
		ValidArgsFunction: completion.AppCompletionFn(p),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			appName := args[0]

			app, err := appClient.Get(cmd.Context(), p.Space, appName)
			if err != nil {
				return err
			}

			space, err := p.GetTargetSpace(cmd.Context())
			if err != nil {
				return err
			}

			envs := []struct {
				name string
				env  []corev1.EnvVar
			}{
				{name: "Space-Provided", env: space.Status.RuntimeConfig.Env},
				{name: "User-Provided", env: (*apps.KfApp)(app).GetEnvVars()},
				{name: "System-Provided", env: resources.BuildRuntimeEnvVars(resources.CFRunning, app)},
			}

			for _, env := range envs {
				describe.SectionWriter(cmd.OutOrStdout(), env.name, func(w io.Writer) {
					for _, e := range env.env {
						value := e.Value

						if e.ValueFrom != nil {
							value = "[Resolved at runtime]"
						}

						fmt.Fprintf(w, "%s:\t%s\n", e.Name, value)
					}
				})
			}

			return nil
		},
	}

	return cmd
}
