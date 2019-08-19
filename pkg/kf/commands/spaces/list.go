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

package spaces

import (
	"fmt"
	"io"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/spaces"
	"k8s.io/apimachinery/pkg/api/meta/table"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"

	"github.com/spf13/cobra"
)

// NewListSpacesCommand allows users to list spaces.
func NewListSpacesCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "List all kf spaces",
		Long: `List spaces and their statuses for the currently targeted cluster.

		The output of this command is similar to what you'd get by running:

		    kubectl get spaces.kf.dev

		`,
		Example: `kf spaces`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			list, err := client.List()
			if err != nil {
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Name\tAge\tReady\tReason")
				for _, space := range list {
					// Status is important here as spaces may be in a deleting status.
					ready := ""
					reason := ""
					if cond := space.Status.GetCondition(v1alpha1.SpaceConditionReady); cond != nil {
						ready = fmt.Sprintf("%v", cond.Status)
						reason = cond.Reason
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s",
						space.Name,
						table.ConvertToHumanReadableDateType(space.CreationTimestamp),
						ready,
						reason,
					)
					fmt.Fprintln(w)
				}
			})

			return nil
		},
	}

	return cmd
}
