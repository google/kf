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

package tasks

import (
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	genericcli "github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/tasks"
	"github.com/spf13/cobra"
)

// NewTasksCommand allows users to list Tasks on a given App.
func NewTasksCommand(p *config.KfParams) *cobra.Command {
	return genericcli.NewListCommand(tasks.NewResourceInfo(), p,
		genericcli.WithListArgumentFilters([]genericcli.ListArgumentFilter{
			{
				Name:    "APP",
				Handler: genericcli.NewAddLabelFilter(v1alpha1.NameLabel),
			},
		}))
}
