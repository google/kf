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

package builds

import (
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/internal/genericcli"
	"github.com/google/kf/pkg/kf/sources"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

// NewGetBuildCommand allows users to get builds.
func NewGetBuildCommand(p *config.KfParams, client dynamic.Interface) *cobra.Command {
	return genericcli.NewDescribeCommand(sources.NewResourceInfo(), p, client)
}
