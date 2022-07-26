// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// BuildpackV3BuildTaskName is the name of the buildpack Tekton Task.
	BuildpackV3BuildTaskName = "buildpackv3"

	// DockerfileBuildTaskName is the name of the Dockerfile Tekton Task.
	DockerfileBuildTaskName = "kaniko"

	// BuildpackV2BuildTaskName is the name of the Cloud Foundry Buildpacks
	// Tekton Task.
	BuildpackV2BuildTaskName = "buildpackv2"

	// SourceImageParamName is the key for the source image param.
	SourceImageParamName = "SOURCE_IMAGE"

	// BuildNameParamName is the key for the Build name param.
	BuildNameParamName = "BUILD_NAME"

	// SourcePackageNameParamName is the key for the SourcePackage name param.
	SourcePackageNameParamName = "SOURCE_PACKAGE_NAME"

	// SourcePackageNamespaceParamName is the key for the SourcePackage namespace param.
	SourcePackageNamespaceParamName = "SOURCE_PACKAGE_NAMESPACE"

	// BuildpackV2ParamName is the key for the Buildpack list param on V2 Builds.
	BuildpackV2ParamName = "BUILDPACKS"

	// BuildpackV3ParamName is the key for the Buildpack list param on V3 Builds.
	BuildpackV3ParamName = "BUILDPACK"

	// RunImageParamName is the key for the run image param.
	RunImageParamName = "RUN_IMAGE"

	// StackV2EnvVarName is the key for the Stack on V2 Builds.
	StackV2EnvVarName = "CF_STACK"

	// BuiltinTaskKind indicates that a Task has an impelemtnation
	// defined by Kf.
	BuiltinTaskKind = "KfBuiltinTask"

	// BuiltinTaskAPIVersion is the version used with Builtin tasks.
	BuiltinTaskAPIVersion = "builtin.kf.dev/v1alpha1"

	// DefaultBuildTimeout is the default timeout value for build to complete.
	DefaultBuildTimeout time.Duration = time.Duration(1 * time.Hour)

	// DefaultBuildRetentionCount is the defualt retention count of number of builds for garbage collection.
	DefaultBuildRetentionCount = 5
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Build represents the source code and build configuration for an App.
type Build struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec BuildSpec `json:"spec,omitempty"`

	// +optional
	Status BuildStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*Build)(nil)
var _ apis.Defaultable = (*Build)(nil)

// BuildSpec defines the source code for an App.
type BuildSpec struct {

	// SourcePackage references the SourcePackage for the source code of the
	// App. If left empty, the SOURCE_IMAGE environment variable will not be
	// added to the resulting TaskRun.
	// +optional
	SourcePackage corev1.LocalObjectReference `json:"sourcePackage,omitempty"`

	// TaskRef is inlined to reference the Tekton Task to use for the Build.
	BuildTaskRef `json:",inline"`

	// Params is the map of keys and values used for the custom task
	Params []BuildParam `json:"params,omitempty"`

	// Env represents the environment variables to apply when building the App.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// NodeSelector represents the selectors to apply when building and deploying the App.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// BuildParam holds custom parameters for the build being run.
// Unlike Tekton Params, Kf only supports string values, but the type
// is otherwise over-the-wire compatible.
type BuildParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ToTektonParam converts the BuildParam the equivalent value for Tekton
// TaskRuns to consume.
func (bp *BuildParam) ToTektonParam() tektonv1beta1.Param {
	return tektonv1beta1.Param{
		Name: bp.Name,
		Value: tektonv1beta1.ArrayOrString{
			// Kf only supports string types, but could expand to Array types in the
			// future if necessary.
			Type:      tektonv1beta1.ParamTypeString,
			StringVal: bp.Value,
		},
	}
}

// BuildTaskRef can be used to refer to a specific instance of a Tekton Task.
type BuildTaskRef struct {
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name,omitempty"`
	// Kind indicates the kind of the task, namespaced or cluster scoped.
	Kind string `json:"kind,omitempty"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
}

// BuildStatus is the current configuration and running state for an App's Build.
type BuildStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	BuildStatusFields `json:",inline"`
}

// BuildStatusFields holds the fields of Build's status that
// are shared. This is defined separately and inlined so that
// other types can readily consume these fields via duck typing.
type BuildStatusFields struct {
	// Image is the latest successfully built image.
	// +optional
	Image string `json:"image,omitempty"`

	// BuildName is the name of the build that produced the image.
	// +optional
	BuildName string `json:"buildName,omitempty"`

	// StartTime contains the time the build started.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime contains the time the build completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Duration contains the duration of the build.
	Duration *metav1.Duration `json:"duration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BuildList is a list of Build resources.
type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Build `json:"items"`
}

func builtinTaskRef(name string) BuildTaskRef {
	return BuildTaskRef{
		Name:       name,
		Kind:       BuiltinTaskKind,
		APIVersion: BuiltinTaskAPIVersion,
	}
}

// StringParam returns a BuildParam for the given key.
func StringParam(key, value string) BuildParam {
	return BuildParam{
		Name:  key,
		Value: value,
	}
}

// buildpackV3BuildTaskRef returns a BuildTaskRef for the buildpack ClusterTask.
func buildpackV3BuildTaskRef() BuildTaskRef {
	return builtinTaskRef(BuildpackV3BuildTaskName)
}

// dockerfileBuildTaskRef returns a BuildTaskRef for the Dockerfile ClusterTask.
func dockerfileBuildTaskRef() BuildTaskRef {
	return builtinTaskRef(DockerfileBuildTaskName)
}

// buildpackV2BuildTaskRef returns a TaskRef for the Cloud Foundry Buildpacks
// ClusterTask.
func buildpackV2BuildTaskRef() BuildTaskRef {
	return builtinTaskRef(BuildpackV2BuildTaskName)
}

// DockerfileBuild is a BuildSpec for building with a Dockerfile.
func DockerfileBuild(sourceImage, dockerfile string) BuildSpec {
	// The params set here should match the params in
	// pkg/reconciler/build/resources/builtin_tasks.go
	return BuildSpec{
		BuildTaskRef: dockerfileBuildTaskRef(),
		Params: []BuildParam{
			StringParam(SourceImageParamName, sourceImage),
			StringParam("DOCKERFILE", dockerfile),
		},
	}
}

// BuildpackV3Build is a BuildSpec for building with Cloud Native Buildpacks.
func BuildpackV3Build(sourceImage string, stack config.StackV3Definition, buildpacks []string) BuildSpec {
	// The params set here should match the params in
	// pkg/reconciler/build/resources/builtin_tasks.go
	return BuildSpec{
		BuildTaskRef: buildpackV3BuildTaskRef(),
		Params: []BuildParam{
			StringParam(SourceImageParamName, sourceImage),
			StringParam(BuildpackV3ParamName, strings.Join(buildpacks, ",")),
			StringParam(RunImageParamName, stack.RunImage),
			StringParam("BUILDER_IMAGE", stack.BuildImage),
		},
		NodeSelector: stack.NodeSelector,
	}
}

// BuildpackV2Build is a BuildSpec for building with Cloud Foundry Buildpacks.
func BuildpackV2Build(sourceImage string, stack config.StackV2Definition, buildpacks []string, skipDetect bool) BuildSpec {
	// The params set here should match the params in
	// pkg/reconciler/build/resources/builtin_tasks.go
	return BuildSpec{
		BuildTaskRef: buildpackV2BuildTaskRef(),
		Params: []BuildParam{
			StringParam(SourceImageParamName, sourceImage),
			StringParam(BuildpackV2ParamName, strings.Join(buildpacks, ",")),
			StringParam(RunImageParamName, stack.Image),
			StringParam("BUILDER_IMAGE", stack.Image),
			StringParam("SKIP_DETECT", fmt.Sprintf("%v", skipDetect)),
		},
		Env: []corev1.EnvVar{
			{
				Name:  StackV2EnvVarName,
				Value: stack.Name,
			},
		},
		NodeSelector: stack.NodeSelector,
	}
}
