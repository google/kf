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

package commands

import (
	"errors"
	"fmt"

	"github.com/google/kf/v2/pkg/kf/commands/completion"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/logging"
)

func suggestMissingSpaceNextAction() {
	utils.SuggestNextAction(utils.NextAction{
		Description: "List known spaces",
		Commands: []string{
			"kf spaces",
			"kubectl get spaces",
		},
	})
}

// NewTargetCommand creates a command that can set the default space.
func NewTargetCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	command := &cobra.Command{
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		Use:   "target",
		Short: "Set the default Space to run commands against.",
		Example: `
		# See the current Space
		kf target
		# Target a Space
		kf target -s my-space
		`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flagsChanged := cmd.Flags().Lookup("target-space").Changed
			flagsChanged = flagsChanged || cmd.Parent().PersistentFlags().Lookup("space").Changed

			// We'll support the use case where the user hands us an argument
			// (instead of using a flag). But we're not going to pick between
			// the argument and the flag, so if they gave us both, fail.

			updateSpace := false

			switch {
			case flagsChanged && len(args) > 0:
				// Both flags and an arg was used. Fail.
				return errors.New("--space (or --target-space) can't be used when the Space is provided via arguments.")
			case flagsChanged:
				// Only flags were used, update.
				updateSpace = true
			case len(args) > 0:
				// Only an arg was used, update.
				p.Space = args[0]
				updateSpace = true
			default:
				updateSpace = false
			}

			if updateSpace {
				if _, err := client.Get(cmd.Context(), p.Space); apierrors.IsNotFound(err) {
					suggestMissingSpaceNextAction()
					return fmt.Errorf("Space %q doesn't exist", p.Space)
				} else if err != nil {
					return err
				}

				if err := config.Write(p.Config, p); err != nil {
					return err
				}
				logging.FromContext(ctx).Info("Updated target Space:")
			}

			fmt.Fprintln(cmd.OutOrStdout(), p.Space)

			return nil
		},
	}

	command.Flags().StringVarP(&p.Space, "target-space", "s", "", "Target the given space.")
	command.RegisterFlagCompletionFunc("target-space", completion.SpaceCompletionFn(p))

	return command
}
