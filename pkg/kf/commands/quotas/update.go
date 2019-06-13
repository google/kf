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
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/quotas"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// NewUpdateQuotaCommand allows users to create quotas.
func NewUpdateQuotaCommand(p *config.KfParams, client quotas.Client) *cobra.Command {
	var (
		memory   string
		cpu      string
		services string
	)

	cmd := &cobra.Command{
		Use:   "update-quota QUOTA",
		Short: "Update a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]

			return client.Transform(p.Namespace, name, func(quota *v1.ResourceQuota) error {
				kfquota := quotas.NewFromResourceQuota(quota)

				var quotaInputs = []struct {
					Value  string
					Setter func(r resource.Quantity)
				}{
					{memory, kfquota.SetMemory},
					{cpu, kfquota.SetCPU},
					{services, kfquota.SetServices},
				}

				// Only update resource quotas for inputted flags
				for _, quota := range quotaInputs {
					if quota.Value != "undefined" {
						quantity, err := resource.ParseQuantity(quota.Value)
						if err != nil {
							return err
						}
						quota.Setter(quantity)
					}

				}
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(
		&memory,
		"memory",
		"m",
		"undefined",
		"The quota for total memory in a space",
	)

	cmd.Flags().StringVarP(
		&cpu,
		"cpu",
		"c",
		"undefined",
		"The quota for total CPU in a space",
	)

	cmd.Flags().StringVarP(
		&services,
		"services",
		"s",
		"undefined",
		"The quota for number of services in a space",
	)

	return cmd
}
