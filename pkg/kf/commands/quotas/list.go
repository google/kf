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
	"text/tabwriter"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/quotas"

	"github.com/spf13/cobra"
)

// NewListQuotasCommand allows users to list quotas.
func NewListQuotasCommand(p *config.KfParams, client quotas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotas",
		Short: "List all kf quotas",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Getting quotas in namespace: %s\n", p.Namespace)

			allQuotas, err := client.List(p.Namespace)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d quotas in namespace %s\n", len(allQuotas), p.Namespace)
			fmt.Fprintln(cmd.OutOrStdout())

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 4, 1, ' ', tabwriter.StripEscape)
			defer w.Flush()

			fmt.Fprintln(w, "NAME\tMEMORY\tCPU\tROUTES")
			for _, quota := range allQuotas {
				kfquota := quotas.NewFromResourceQuota(&quota)
				mem, _ := kfquota.GetMemory()
				cpu, _ := kfquota.GetCPU()
				routes, _ := kfquota.GetServices()
				fmt.Fprintf(w, "%s\t%v\t%v\t%v\n",
					kfquota.GetName(),
					mem.String(),
					cpu.String(),
					routes.String())
			}

			return nil
		},
	}

	return cmd
}
