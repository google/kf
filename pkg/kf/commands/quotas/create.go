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

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/quotas"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// Default value when the user does not pass in a quota for a particular resource.
	// This value is never set in the actual ResourceQuota definition.
	DefaultQuota = "undefined"
)

// NewCreateQuotaCommand allows users to create quotas.
func NewCreateQuotaCommand(p *config.KfParams, client quotas.Client) *cobra.Command {
	var (
		memory string
		cpu    string
		routes string
	)
	cmd := &cobra.Command{
		Use:   "create-quota QUOTA",
		Short: "Create a quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			kfquota := quotas.NewKfQuota()
			kfquota.SetName(name)

			var quotaInputs = []struct {
				Value  string
				Setter func(r resource.Quantity)
			}{
				{memory, kfquota.SetMemory},
				{cpu, kfquota.SetCPU},
				{routes, kfquota.SetServices},
			}

			// Only set resource quotas for inputted flags
			for _, quota := range quotaInputs {
				if quota.Value != DefaultQuota {
					quantity, err := resource.ParseQuantity(quota.Value)
					if err != nil {
						return err
					}
					quota.Setter(quantity)
				}

			}

			if _, creationErr := client.Create(p.Namespace, kfquota.ToResourceQuota()); creationErr != nil {
				return creationErr
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Quota %q successfully created\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(
		&memory,
		"memory",
		"m",
		DefaultQuota,
		"The total available memory across all builds and applications in a space (e.g. 10Gi, 500Mi). Default: unlimited",
	)

	cmd.Flags().StringVarP(
		&cpu,
		"cpu",
		"c",
		DefaultQuota,
		"The total available CPU across all builds and applications in a space (e.g. 400m). Default: unlimited",
	)

	cmd.Flags().StringVarP(
		&routes,
		"routes",
		"r",
		DefaultQuota,
		"The total number of routes that can exist in a space. Default: unlimited",
	)

	return cmd
}
