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
	build "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *Build) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Build")
}

const (

	// TaskRunParamSourceImage is the Tekton Param name for the container image
	// which contains the source code for a build.
	TaskRunParamSourceImage = "SOURCE_IMAGE"

	// TaskRunResourceURL is the Tekton param name for the desired destination image.
	TaskRunParamDestinationImage = "DESTINATION_IMAGE"
)

// PropagateBuildStatus copies fields from the Build status to Source and
// updates the readiness based on the current phase.
func (status *BuildStatus) PropagateBuildStatus(build *build.TaskRun) {

	if build == nil {
		return
	}

	status.BuildName = build.Name
	status.StartTime = build.Status.StartTime
	status.CompletionTime = build.Status.CompletionTime

	// Generate a build duration for easy printing.
	if status.StartTime != nil && status.CompletionTime != nil {
		status.Duration = &metav1.Duration{
			Duration: status.CompletionTime.Time.Sub(status.StartTime.Time),
		}
	}

	status.manage().MarkUnknown(BuildConditionTaskRunReady, "initializing", "Build in progress")

	cond := build.Status.GetCondition(apis.ConditionSucceeded)
	if PropagateCondition(status.manage(), BuildConditionTaskRunReady, cond) {
		image := GetTaskRunResults(build, TaskRunParamDestinationImage)

		// try fetch from task resources. remove this after Kf v2.11.20.
		if image == "" {
			image = GetTaskRunOutputResource(build, "IMAGE", "url")
		}

		status.Image = image
	}
}

// PropagateSourcePackageStatus copies the condition from the SourcePackage to
// the BuildStatus.
func (status *BuildStatus) PropagateSourcePackageStatus(sourcePackage *SourcePackage) {
	if sourcePackage == nil {
		return
	}

	status.SourcePackageCondition().MarkUnknown("initializing", "Upload in progress")

	cond := sourcePackage.Status.GetCondition(apis.ConditionSucceeded)
	PropagateCondition(status.manage(), BuildConditionSourcePackageReady, cond)
}

// PropagateTerminatingStatus updates the ready status of the build to False if the build received a delete request.
func (status *BuildStatus) PropagateTerminatingStatus() {
	status.manage().MarkFalse(BuildConditionSucceeded, "Terminating", "Build is terminating")
}

func GetTaskRunResults(b *build.TaskRun, paramName string) string {
	for _, result := range b.Status.TaskRunResults {
		if result.Name != paramName {
			continue
		}

		return result.Value.StringVal
	}

	return ""
}

// b/291822470 remove after kf v2.11.20
func GetTaskRunOutputResource(b *build.TaskRun, resourceName, paramName string) string {
	for _, resource := range b.Spec.Resources.Outputs {
		if resource.PipelineResourceBinding.Name != resourceName {
			continue
		}

		for _, param := range resource.PipelineResourceBinding.ResourceSpec.Params {
			if param.Name == paramName {
				return param.Value
			}
		}
	}

	return ""
}

// MarkSpaceHealthy notes that the Space was able to be retrieved and
// defaults can be applied from it.
func (status *BuildStatus) MarkSpaceHealthy() {
	status.SpaceCondition().MarkSuccess()
}

// MarkSpaceUnhealthy notes that the Space was could not be retrieved.
func (status *BuildStatus) MarkSpaceUnhealthy(reason, message string) {
	status.SpaceCondition().MarkFalse(reason, message)
}
