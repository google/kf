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
	"context"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/builds"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewBuildsCommand allows users to get builds.
func NewBuildsCommand(p *config.KfParams) *cobra.Command {
	return genericcli.NewListCommand(&adxBuildResourceInfo{
		p:   p,
		old: builds.NewResourceInfo(),
	}, p, genericcli.WithListLabelFilters(map[string]string{
		"app": v1alpha1.NameLabel,
	}))
}

type adxBuildResourceInfo struct {
	p   *config.KfParams
	old genericcli.Type
}

func (a *adxBuildResourceInfo) Namespaced() bool {
	return true
}

func (a *adxBuildResourceInfo) GroupVersionResource(ctx context.Context) schema.GroupVersionResource {
	if a.p.FeatureFlags(ctx).AppDevExperienceBuilds().IsDisabled() {
		return a.old.GroupVersionResource(ctx)
	}

	return schema.GroupVersionResource{
		Group:    "builds.appdevexperience.dev",
		Version:  "v1alpha1",
		Resource: "builds",
	}
}

func (a *adxBuildResourceInfo) GroupVersionKind(ctx context.Context) schema.GroupVersionKind {
	if a.p.FeatureFlags(ctx).AppDevExperienceBuilds().IsDisabled() {
		return a.old.GroupVersionKind(ctx)
	}

	return schema.GroupVersionKind{
		Group:   "builds.appdevexperience.dev",
		Version: "v1alpha1",
		Kind:    "Build",
	}
}

func (a *adxBuildResourceInfo) FriendlyName() string {
	return "Build"
}
