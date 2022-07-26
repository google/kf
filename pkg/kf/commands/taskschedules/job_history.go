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
	"github.com/google/kf/v2/pkg/kf/tasks"
	"github.com/google/kf/v2/pkg/reconciler/taskschedule/resources"
	"github.com/spf13/cobra"
)

// NewJobHistoryCommand lists all Tasks run by the given TaskSchedule.
func NewJobHistoryCommand(p *config.KfParams) *cobra.Command {
	return genericcli.NewListCommand(
		tasks.NewResourceInfo(),
		p,
		genericcli.WithListCommandName("job-history"),
		genericcli.WithListExample("kf job-history my-job"),
		genericcli.WithListShort("List the execution history of a Job."),
		genericcli.WithListLong("The job-history sub-command lets operators view the execution history of a Job."),
		genericcli.WithListPluralFriendlyName("job history"),
		genericcli.WithListArgumentFilters([]genericcli.ListArgumentFilter{
			{
				Name:     "JOB_NAME",
				Handler:  genericcli.NewAddLabelFilter(resources.OwningTaskSchedule),
				Required: true,
			},
		}),
	)
}
