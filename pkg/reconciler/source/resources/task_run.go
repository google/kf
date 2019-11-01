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

package resources

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const (
	managedByLabel            = "app.kubernetes.io/managed-by"
	buildpackClusterTask      = "buildpack"
	containerImageClusterTask = "container"
	dockerImageClusterTask    = "kaniko"
)

// TaskRunName gets the name of a TaskRun for a Source.
func TaskRunName(source *v1alpha1.Source) string {
	return source.Name
}

// TaskRunSecretName gets the name of a Secret for a Source.
func TaskRunSecretName(source *v1alpha1.Source) string {
	return TaskRunName(source)
}

func makeContainerImageTaskRun(source *v1alpha1.Source) (*tekton.TaskRun, *corev1.Secret, error) {
	return &tekton.TaskRun{
		ObjectMeta: makeObjectMeta(TaskRunName(source), source),
		Spec: tekton.TaskRunSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			TaskRef: &tekton.TaskRef{
				Name: containerImageClusterTask,
				Kind: "ClusterTask",
			},
			Outputs: makeOutputs(source.Spec.ContainerImage.Image),
		},
	}, nil, nil
}

func makeDockerImageTaskRun(source *v1alpha1.Source) (*tekton.TaskRun, *corev1.Secret, error) {
	return &tekton.TaskRun{
		ObjectMeta: makeObjectMeta(TaskRunName(source), source),
		Spec: tekton.TaskRunSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			TaskRef: &tekton.TaskRef{
				Name: dockerImageClusterTask,
				Kind: "ClusterTask",
			},
			Inputs: tekton.TaskRunInputs{
				Params: []tekton.Param{
					{
						Name: v1alpha1.TaskRunParamSourceContainer,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.Dockerfile.Source,
						},
					},
					{
						Name: v1alpha1.TaskRunParamDockerfile,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.Dockerfile.Path,
						},
					},
				},
			},
			Outputs: makeOutputs(source.Spec.Dockerfile.Image),
		},
	}, nil, nil
}

func makeBuildpackTaskRun(source *v1alpha1.Source) (*tekton.TaskRun, *corev1.Secret, error) {
	secret := makeSecret(source)

	return &tekton.TaskRun{
		ObjectMeta: makeObjectMeta(TaskRunName(source), source),
		Spec: tekton.TaskRunSpec{
			ServiceAccountName: source.Spec.ServiceAccount,
			TaskRef: &tekton.TaskRef{
				Name: buildpackClusterTask,
				Kind: "ClusterTask",
			},
			Inputs: tekton.TaskRunInputs{
				Params: []tekton.Param{
					{
						Name: v1alpha1.TaskRunParamSourceContainer,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.BuildpackBuild.Source,
						},
					},
					{
						Name: v1alpha1.TaskRunParamBuildpackBuilder,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.BuildpackBuild.BuildpackBuilder,
						},
					},
					{
						Name: v1alpha1.TaskRunParamBuildpack,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.BuildpackBuild.Buildpack,
						},
					},
					{
						Name: v1alpha1.TaskRunParamBuildpackRunImage,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: source.Spec.BuildpackBuild.Stack,
						},
					},
					{
						Name: v1alpha1.TaskRunParamEnvSecret,
						Value: tekton.ArrayOrString{
							Type:      tekton.ParamTypeString,
							StringVal: secret.Name,
						},
					},
				},
			},
			Outputs: makeOutputs(source.Spec.BuildpackBuild.Image),
		},
	}, secret, nil
}

func makeSecret(source *v1alpha1.Source) *corev1.Secret {
	m := map[string][]byte{}

	// TODO(#821): Support source.Env.ValueFrom
	for _, e := range source.Spec.BuildpackBuild.Env {
		m[e.Name] = []byte(e.Value)
	}

	return &corev1.Secret{
		ObjectMeta: makeObjectMeta(TaskRunSecretName(source), source),
		Data:       m,
	}
}

func makeObjectMeta(name string, source *v1alpha1.Source) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: source.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(source),
		},
		// Copy labels from the parent
		Labels: v1alpha1.UnionMaps(
			source.GetLabels(), map[string]string{
				managedByLabel: "kf",
			}),
	}
}

func makeOutputs(image string) tekton.TaskRunOutputs {
	return tekton.TaskRunOutputs{
		Resources: []tekton.TaskResourceBinding{
			{
				PipelineResourceBinding: tekton.PipelineResourceBinding{
					Name: v1alpha1.TaskRunResourceNameImage,
					ResourceSpec: &tekton.PipelineResourceSpec{
						Type: "image",
						Params: []tekton.ResourceParam{
							{
								Name:  v1alpha1.TaskRunResourceURL,
								Value: image,
							},
						},
					},
				},
			},
		},
	}
}

// MakeTaskRun creates a TaskRun and Secret for a Source. The Secret CAN be nil if
// the Source does not require it.
func MakeTaskRun(source *v1alpha1.Source) (*tekton.TaskRun, *corev1.Secret, error) {
	switch {
	case source.Spec.IsContainerBuild():
		return makeContainerImageTaskRun(source)
	case source.Spec.IsDockerfileBuild():
		return makeDockerImageTaskRun(source)
	default:
		return makeBuildpackTaskRun(source)
	}
}
