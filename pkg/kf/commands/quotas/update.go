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
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/quotas"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

// NewUpdateQuotaCommand allows users to create quotas.
func NewUpdateQuotaCommand(p *config.KfParams, client quotas.Client) *cobra.Command {
	var (
		memory string
		cpu    string
		routes string
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
				return setQuotaValues(memory, cpu, routes, kfquota)
			})
		},
	}

	cmd.Flags().StringVarP(
		&memory,
		"memory",
		"m",
		defaultQuota,
		"The total available memory across all builds and applications in a space (e.g. 10Gi, 500Mi). Default: unlimited",
	)

	cmd.Flags().StringVarP(
		&cpu,
		"cpu",
		"c",
		defaultQuota,
		"The total available CPU across all builds and applications in a space (e.g. 400m). Default: unlimited",
	)

	cmd.Flags().StringVarP(
		&routes,
		"routes",
		"r",
		defaultQuota,
		"The total number of routes that can exist in a space. Default: unlimited",
	)

	return cmd
}
