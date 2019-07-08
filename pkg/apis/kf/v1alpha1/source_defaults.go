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

import "context"

// SetDefaults implements apis.Defaultable
func (k *Source) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *SourceSpec) SetDefaults(ctx context.Context) {
	// XXX: currently no defaults to set
}

// SetSpaceDefaults sets the default values for the source based on the space's
// settings.
func (k *Source) SetSpaceDefaults(space *Space) {
	k.Spec.SetSpaceDefaults(space)
}

// SetSpaceDefaults sets the default values for the SourceSpec based on the
// space's settings.
func (k *SourceSpec) SetSpaceDefaults(space *Space) {
	if k.IsBuildpackBuild() {
		if k.BuildpackBuild.BuildpackBuilder == "" {
			k.BuildpackBuild.BuildpackBuilder = space.Spec.BuildpackBuild.BuilderImage
		}

		if k.BuildpackBuild.Registry == "" {
			k.BuildpackBuild.Registry = space.Spec.BuildpackBuild.ContainerRegistry
		}

		k.BuildpackBuild.Env = space.Spec.BuildpackBuild.Env
	}
}
