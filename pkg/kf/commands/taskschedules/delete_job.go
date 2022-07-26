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

package taskschedules

import (
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/spf13/cobra"
)

// NewDeleteJobCommand allows users to delete TaskSchedules.
func NewDeleteJobCommand(p *config.KfParams) *cobra.Command {
	cmd := genericcli.NewDeleteByNameCommand(
		resourceInfo,
		p,
		genericcli.WithDeleteByNameCommandName("delete-job"),
		genericcli.WithDeleteByNameAdditionalLongText(`
		Note: Deleting the Job will delete all associated executions,
		even executions which are still running.
		`),
	)

	return cmd
}
