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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/google/kf/pkg/kf/algorithms"
	routecfg "github.com/google/kf/third_party/knative-serving/pkg/reconciler/route/config"
)

// TODO(#396): We should pull these from a ConfigMap
const (
	// DefaultBuilderImage contains the default buildpack builder image.
	DefaultBuilderImage = "gcr.io/kf-releases/buildpack-builder:latest"

	// DefaultDomainTemplate contains the default domain template. It should
	// be used with `fmt.Sprintf(DefaultDomainTemplate, namespace)`
	DefaultDomainTemplate = "%s.%s"
)

// SetDefaults implements apis.Defaultable
func (k *Space) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx, k.Name)
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpec) SetDefaults(ctx context.Context, name string) {
	k.Security.SetDefaults(ctx)
	k.BuildpackBuild.SetDefaults(ctx)
	k.Execution.SetDefaults(ctx, name)
	k.ResourceLimits.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecSecurity) SetDefaults(ctx context.Context) {
	// TODO(#458): We eventually want this to be configurable.
	k.EnableDeveloperLogsAccess = true
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecBuildpackBuild) SetDefaults(ctx context.Context) {
	if k.BuilderImage == "" {
		k.BuilderImage = DefaultBuilderImage
	}
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecExecution) SetDefaults(ctx context.Context, name string) {
	if len(k.Domains) == 0 {
		k.Domains = append(
			k.Domains,
			SpaceDomain{
				Domain:  fmt.Sprintf(DefaultDomainTemplate, name, DefaultDomain(ctx)),
				Default: true,
			},
		)
	}

	k.Domains = []SpaceDomain(algorithms.Dedupe(
		SpaceDomains(k.Domains),
	).(SpaceDomains))
}

// DefaultDomain gets the default domain to use for spaces from the context.
func DefaultDomain(ctx context.Context) (domain string) {
	// routecfg.FromContext can panic if the resource isn't on the context rather
	// than returning nil. In this case, we still just want the DefaultDomain.
	defer func() {
		if r := recover(); r != nil {
			domain = routecfg.DefaultDomain
		}
	}()

	if ctx == nil {
		return routecfg.DefaultDomain
	}

	rc := routecfg.FromContext(ctx)
	if rc == nil {
		return routecfg.DefaultDomain
	}

	if domainCfg := rc.Domain; domainCfg != nil {
		return domainCfg.LookupDomainForLabels(map[string]string{})
	}

	return routecfg.DefaultDomain
}

// SetDefaults implements apis.Defaultable
func (k *SpaceSpecResourceLimits) SetDefaults(ctx context.Context) {
	// XXX: currently no defaults to set
}
