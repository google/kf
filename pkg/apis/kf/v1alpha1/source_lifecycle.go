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
	build "github.com/google/kf/third_party/knative-build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Source) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Source")
}

const (
	// SourceConditionSucceeded is set when the source is configured
	// and is usable by developers.
	SourceConditionSucceeded = apis.ConditionSucceeded
	// SourceConditionBuildSucceeded is set when Build has succeeded.
	SourceConditionBuildSucceeded apis.ConditionType = "BuildSucceeded"
	// SourceConditionBuildSecretReady is set when the build Secret is ready.
	SourceConditionBuildSecretReady apis.ConditionType = "BuildSecretReady"

	BuildArgImage             = "IMAGE"
	BuildArgBuildpack         = "BUILDPACK"
	BuildArgBuildpackBuilder  = "BUILDER_IMAGE"
	BuildArgBuildpackRunImage = "RUN_IMAGE"
	BuildArgDockerfile        = "DOCKERFILE"
)

func (status *SourceStatus) manage() apis.ConditionManager {
	return apis.NewBatchConditionSet(
		SourceConditionBuildSucceeded,
		SourceConditionBuildSecretReady,
	).Manage(status)
}

// Succeeded returns if the space is ready to be used.
func (status *SourceStatus) Succeeded() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *SourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *SourceStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// PropagateBuildStatus copies fields from the Build status to Source and
// updates the readiness based on the current phase.
func (status *SourceStatus) PropagateBuildStatus(build *build.Build) {

	if build == nil {
		return
	}

	status.BuildName = build.Name
	status.manage().MarkUnknown(SourceConditionBuildSucceeded, "initializing", "Build in progress")

	cond := build.Status.GetCondition(apis.ConditionSucceeded)
	if PropagateCondition(status.manage(), SourceConditionBuildSucceeded, cond) {
		status.Image = GetBuildArg(build, BuildArgImage)
	}
}

// PropagateBuildSecretStatus copies fields from the Secret status to Source
// and updates the readiness based on the current phase.
func (status *SourceStatus) PropagateBuildSecretStatus(secret *corev1.Secret) {
	status.manage().MarkTrue(SourceConditionBuildSecretReady)
}

// BuildCondition gets a manager for the state of the build.
func (status *SourceStatus) BuildCondition() SingleConditionManager {
	return NewSingleConditionManager(
		status.manage(),
		SourceConditionBuildSucceeded,
		"Build",
	)
}

// BuildSecretCondition gets a manager for the state of the env var secret.
func (status *SourceStatus) BuildSecretCondition() SingleConditionManager {
	return NewSingleConditionManager(
		status.manage(),
		SourceConditionBuildSecretReady,
		"Build Secret",
	)
}

func GetBuildArg(b *build.Build, key string) string {
	for _, arg := range b.Spec.Template.Arguments {
		if arg.Name == key {
			return arg.Value
		}
	}
	return ""
}

func (status *SourceStatus) duck() *duckv1beta1.Status {
	return &status.Status
}
