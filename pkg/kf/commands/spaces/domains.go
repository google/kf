// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/describe"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewDomainsCommand allows developers to list domains for a Space.
func NewDomainsCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "domains",
		Short:        "List domains that can be used in the targeted Space.",
		Example:      `kf domains`,
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			space, err := client.Get(ctx, p.Space)
			if err != nil {
				return fmt.Errorf("failed to get Space: %v", err)
			}

			logging.FromContext(ctx).Infof("Listing domains in Space: %s", p.Space)

			describe.TabbedWriter(cmd.OutOrStdout(), func(w io.Writer) {
				fmt.Fprintln(w, "Domain\tGateway")

				// Space status has domains in a deterministic order.
				for _, domain := range space.Status.NetworkConfig.Domains {
					fmt.Fprintf(w, "%s\t%s\n", domain.Domain, domain.GatewayName)
				}
			})

			utils.SuggestNextAction(utils.NextAction{
				Description: "Add a domain",
				Commands: []string{
					"kf configure-space append-domain",
				},
			})

			return nil
		},
	}

	return cmd
}
