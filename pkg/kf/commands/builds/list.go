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

package builds

import (
	"fmt"
	"io"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/sources"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta/table"
)

// NewListBuildsCommand allows users to list spaces.
func NewListBuildsCommand(p *config.KfParams, client sources.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "builds",
		Short:   "List the builds in the current space",
		Example: `kf builds`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			list, err := client.List(p.Namespace)
			if err != nil {
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Name\tAge\tReady\tReason\tImage")

				for _, source := range list {
					ready := ""
					reason := ""
					if cond := source.Status.GetCondition(v1alpha1.SourceConditionSucceeded); cond != nil {
						ready = fmt.Sprintf("%v", cond.Status)
						reason = cond.Reason
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s",
						source.Name,
						table.ConvertToHumanReadableDateType(source.CreationTimestamp),
						ready,
						reason,
						source.Status.Image,
					)
					fmt.Fprintln(w)
				}
			})

			return nil
		},
	}

	return cmd
}
