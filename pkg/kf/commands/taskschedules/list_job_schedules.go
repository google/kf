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
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// NewListJobSchedulesCommand lists all TaskSchedules that are not suspended.
func NewListJobSchedulesCommand(p *config.KfParams) *cobra.Command {
	req, err := labels.NewRequirement(v1alpha1.TaskScheduleSuspendLabel, selection.Equals, []string{"false"})
	if err != nil {
		panic(errors.Wrap(err, "Failed to create suspend label requirement"))
	}
	return genericcli.NewListCommand(
		resourceInfo,
		p,
		genericcli.WithListCommandName("job-schedules"),
		genericcli.WithListPluralFriendlyName("job schedules"),
		genericcli.WithListLabelRequirements([]labels.Requirement{*req}),
	)
}
