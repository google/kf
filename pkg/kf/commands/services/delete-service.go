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

package services

import (
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/spf13/cobra"
)

// NewDeleteServiceCommand allows users to delete service instances.
func NewDeleteServiceCommand(p *config.KfParams, client serviceinstances.Client) *cobra.Command {
	cmd := genericcli.NewDeleteByNameCommand(
		serviceinstances.NewResourceInfo(),
		p,
		genericcli.WithDeleteByNameCommandName("delete-service"),
		genericcli.WithDeleteByNameAliases([]string{"ds"}),
		genericcli.WithDeleteByNameAdditionalLongText(`
		You should delete all bindings before deleting a service. If you don't, the
		service will wait for that to occur before deleting.
		`),
	)

	originalRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		ctx := c.Context()
		resourceName := args[0]

		mutator := func(instance *v1alpha1.ServiceInstance) error {
			instance.Spec.DeleteRequests++
			return nil
		}

		if _, err := client.Transform(ctx, p.Space, resourceName, mutator); err != nil {
			return fmt.Errorf("Failed to update unbinding requests: %s", err)
		}

		return originalRunE(c, args)
	}

	return cmd
}
