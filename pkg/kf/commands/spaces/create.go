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
	"context"
	"fmt"
	"time"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/spaces"

	"github.com/spf13/cobra"
)

// NewCreateSpaceCommand allows users to create spaces.
func NewCreateSpaceCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	var (
		containerRegistry string
		domains           []string
	)

	cmd := &cobra.Command{
		Use:   "create-space SPACE",
		Short: "Create a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			toCreate := spaces.NewKfSpace()
			toCreate.SetName(name)
			toCreate.SetContainerRegistry(containerRegistry)

			for i, domain := range domains {
				toCreate.AppendDomains(v1alpha1.SpaceDomain{Domain: domain, Default: i == 0})
			}

			if _, err := client.Create(toCreate.ToSpace()); err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			fmt.Fprintln(w, "Space requested, waiting for subcomponents to be created")
			space, err := client.WaitFor(context.Background(), name, 1*time.Second, spaces.IsStatusFinal)
			if err != nil {
				return err
			}
			fmt.Fprintln(w, "Space created")
			describe.DuckStatus(w, space.Status.Status)
			fmt.Fprintln(w)

			printAdditionalCommands(cmd.OutOrStdout(), name)
			return nil
		},
	}

	cmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry apps and sources will be stored in.",
	)

	cmd.Flags().StringArrayVar(
		&domains,
		"domain",
		nil,
		"Sets the valid domains for the space. The first provided domain will be the default.",
	)

	return cmd
}
