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
	"fmt"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

// NewTargetCommand creates a command that can set the default space.
func NewTargetCommand(p *config.KfParams) *cobra.Command {
	var space string

	command := &cobra.Command{
		Use:     "target",
		Short:   "Set or view the targeted space",
		Example: `  kf target`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if space != "" {
				p.Namespace = space
				if err := config.Write(p.Config, p); err != nil {
					return err
				}
			}

			fmt.Println("Current space is:", p.Namespace)

			return nil
		},
	}

	command.Flags().StringVarP(&space, "space", "s", "", "Target the given space.")

	return command
}
