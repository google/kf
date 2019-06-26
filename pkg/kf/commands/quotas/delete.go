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

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/quotas"

	"github.com/spf13/cobra"
)

// NewDeleteQuotaCommand allows users to delete quotas.
func NewDeleteQuotaCommand(p *config.KfParams, client quotas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-quota QUOTA",
		Short: "Delete a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := client.Delete(p.Namespace, name); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Quota %q successfully deleted\n", name)

			return nil

		},
	}

	return cmd
}
