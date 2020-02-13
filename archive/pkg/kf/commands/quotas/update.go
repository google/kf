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

package quotas

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/completion"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

const (
	// Default value when the user does not pass in a quota for a particular resource.
	// This value is never set in the actual ResourceQuota definition.
	defaultQuota = "undefined"
)

// NewUpdateQuotaCommand allows users to create a quota for a space.
func NewUpdateQuotaCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	var (
		memory string
		cpu    string
		routes string
	)

	cmd := &cobra.Command{
		Use:        "update-quota SPACE_NAME [-m MEMORY] [-r ROUTES] [-c CPU]",
		Short:      "Update the quota for a space",
		Example:    "kf update-quota my-space --memory 100Gi --routes 50",
		Args:       cobra.ExactArgs(1),
		Aliases:    []string{"create-quota"},
		SuggestFor: []string{"create-space-quota", "update-space-quota"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			spaceName := args[0]

			_, err := client.Transform(spaceName, spaces.DiffWrapper(cmd.OutOrStdout(), func(space *v1alpha1.Space) error {
				kfspace := spaces.NewFromSpace(space)
				return setQuotaValues(memory, cpu, routes, kfspace)
			}))

			return err
		},
	}

	cmd.Flags().StringVarP(
		&memory,
		"memory",
		"m",
		defaultQuota,
		"Total amount of memory the space can have (e.g. 10Gi, 500Mi) (default: unlimited)",
	)

	cmd.Flags().StringVarP(
		&cpu,
		"cpu",
		"c",
		defaultQuota,
		"Total amount of CPU the space can have (e.g. 400m) (default: unlimited)",
	)

	cmd.Flags().StringVarP(
		&routes,
		"routes",
		"r",
		defaultQuota,
		"Maximum number of routes the space can have (default: unlimited)",
	)

	completion.MarkArgCompletionSupported(cmd, completion.SpaceCompletion)

	return cmd
}
