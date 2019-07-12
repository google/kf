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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/spaces"

	"github.com/spf13/cobra"
)

// NewDeleteQuotaCommand allows users to delete quotas.
func NewDeleteQuotaCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-quota SPACE_NAME",
		Short: "Delete a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			spaceName := args[0]

			err := client.Transform(spaceName, func(space *v1alpha1.Space) error {
				kfspace := spaces.NewFromSpace(space)
				return kfspace.DeleteQuota()
			})

			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Quota in space %q successfully deleted\n", spaceName)

			return nil

		},
	}

	return cmd
}
