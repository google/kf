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

// AppSpecSourceMask is a _shallow_ copy of the SourceSpec object to a new
// SourceSpec object bringing over only the fields allowed to be set in
// the App by developers. This does not validate the contents or the bounds of
// the provided fields.
//
// This function should be used with
// godoc.org/knative.dev/pkg/apis#CheckDisallowedFields to validate that the
// user hasn't set any fields they're not allowed to in the source of AppSpec.
func AppSpecSourceMask(in SourceSpec) SourceSpec {
	out := SourceSpec{}

	// Allowed fields. This is exhaustive to prevent new fields added to
	// SourceSpec from being accidentally exposed.
	out.BuildpackBuild.Buildpack = in.BuildpackBuild.Buildpack
	out.BuildpackBuild.Env = in.BuildpackBuild.Env
	out.BuildpackBuild.Source = in.BuildpackBuild.Source
	out.BuildpackBuild.Stack = in.BuildpackBuild.Stack
	out.UpdateRequests = in.UpdateRequests
	out.ContainerImage.Image = in.ContainerImage.Image

	// Disallowed fields
	// This list is unnecessary, but added here for clarity
	out.BuildpackBuild.Image = ""
	out.BuildpackBuild.BuildpackBuilder = ""
	out.ServiceAccount = ""

	return out
}
