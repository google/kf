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
	"text/tabwriter"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/spaces"
	"k8s.io/apimachinery/pkg/api/meta/table"

	"github.com/spf13/cobra"
)

// NewListSpacesCommand allows users to list spaces.
func NewListSpacesCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "List all kf spaces",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			list, err := client.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(p.Output, 8, 4, 1, ' ', tabwriter.StripEscape)
			defer w.Flush()

			// Status is important here as spaces may be in a deleting status.
			fmt.Fprintln(w, "Name\tStatus\tAge")
			for _, space := range list {
				fmt.Fprintf(w, "%s\t%s\t%s",
					space.Name,
					space.Status.Phase,
					table.ConvertToHumanReadableDateType(space.CreationTimestamp))
				fmt.Fprintln(w)
			}

			return nil
		},
	}

	return cmd
}
