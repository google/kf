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
	"fmt"
	"io"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

// NewGetQuotaCommand allows users to get quota info.
func NewGetQuotaCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quota SPACE_NAME",
		Short:   "Show quota info for a space",
		Example: `kf quota my-space`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceName := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "Getting info for quota in space: %s\n\n", spaceName)

			space, err := client.Get(spaceName)
			if err != nil {
				return err
			}

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Memory\tCPU\tRoutes")

				kfspace := spaces.NewFromSpace(space)
				mem, _ := kfspace.GetMemory()
				cpu, _ := kfspace.GetCPU()
				routes, _ := kfspace.GetServices()
				fmt.Fprintf(w, "%v\t%v\t%v\n",
					mem.String(),
					cpu.String(),
					routes.String())
			})
			return nil
		},
	}

	return cmd
}
