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
	"hash/crc64"
	"path"
	"strconv"
)

const (
	// RouteHostname is the hostname of a route.
	RouteHostname = "route.kf.dev/hostname"
	// RouteDomain is the domain of a route.
	RouteDomain = "route.kf.dev/domain"
	// RoutePath is the URL path of a route.
	RoutePath = "route.kf.dev/path"
	// RouteAppName is the App's name that owns the Route.
	RouteAppName = "route.kf.dev/appname"
)

// GenerateRouteClaimName creates the deterministic name for a Route claim.
func GenerateRouteClaimName(hostname, domain, urlPath string) string {
	return GenerateRouteName(hostname, domain, urlPath, "")
}

// GenerateRouteName creates the deterministic name for a Route.
func GenerateRouteName(hostname, domain, urlPath, appName string) string {
	return GenerateName(hostname, domain, path.Join("/", urlPath), appName)
}

// GenerateRouteNameFromSpec creates the deterministic name for a Route.
func GenerateRouteNameFromSpec(spec RouteSpecFields, appName string) string {
	return GenerateName(spec.Hostname, spec.Domain, spec.Path, appName)
}

// SetDefaults implements apis.Defaultable
func (k *Route) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
	k.Labels = UnionMaps(k.Labels, k.Spec.RouteSpecFields.labels())
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpec) SetDefaults(ctx context.Context) {
	k.RouteSpecFields.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *RouteSpecFields) SetDefaults(ctx context.Context) {
	k.Path = path.Join("/", k.Path)
}

func (k *RouteSpecFields) labels() map[string]string {
	return map[string]string{
		ManagedByLabel: "kf",
		ComponentLabel: "route",
		RouteHostname:  k.Hostname,
		RouteDomain:    k.Domain,
		RoutePath:      ToBase36(k.Path),
	}
}

// SetSpaceDefaults sets the default values for the Route based on the space's
// settings.
func (k *Route) SetSpaceDefaults(space *Space) {
	k.Spec.SetSpaceDefaults(space)
}

// SetSpaceDefaults sets the default values for the RouteSpec based on the
// space's settings.
func (k *RouteSpec) SetSpaceDefaults(space *Space) {
	k.RouteSpecFields.SetSpaceDefaults(space)
}

// SetSpaceDefaults sets the default values for the RouteSpec based on the
// space's settings.
func (k *RouteSpecFields) SetSpaceDefaults(space *Space) {
	if k.Domain == "" {
		// Use space's default domain
		for _, domain := range space.Spec.Execution.Domains {
			if !domain.Default {
				continue
			}
			k.Domain = domain.Domain
			break
		}
	}
}

// SetDefaults sets the defaults for a RouteClaim.
func (k *RouteClaim) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
	k.Labels = UnionMaps(k.Labels, k.Spec.RouteSpecFields.labels())
}

// SetDefaults implements apis.Defaultable
func (k *RouteClaimSpec) SetDefaults(ctx context.Context) {
	k.RouteSpecFields.SetDefaults(ctx)
}

// ToBase36 is a helpful function that converts a string into something that
// is encoded and safe for URLs, names etc... Base 36 uses 0-9a-z
func ToBase36(s string) string {
	return strconv.FormatUint(
		crc64.Checksum(
			[]byte(s),
			crc64.MakeTable(crc64.ECMA),
		),
		36)
}
