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
	"errors"
	"fmt"
	"path"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/selectorutil"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektonresource "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const (
	managedByLabel = "app.kubernetes.io/managed-by"
)

// TaskRunName gets the name of a TaskRun for a Build.
func TaskRunName(build *v1alpha1.Build) string {
	return build.Name
}

func makeObjectMeta(
	name string,
	build *v1alpha1.Build,
	defaultsConfig *kfconfig.DefaultsConfig,
) metav1.ObjectMeta {
	sidecarAnnotation := "true"
	if defaultsConfig != nil && defaultsConfig.BuildDisableIstioSidecar {
		sidecarAnnotation = "false"
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: build.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(build),
		},
		// Copy labels from the parent
		Labels: v1alpha1.UnionMaps(
			build.GetLabels(), map[string]string{
				managedByLabel:              "kf",
				v1alpha1.NetworkPolicyLabel: v1alpha1.NetworkPolicyBuild,
			}),
		Annotations: map[string]string{
			// Allow Istio injection on Tekton tasks.
			"sidecar.istio.io/inject": sidecarAnnotation,
		},
	}
}

// DestinationImageName gets the image name for an application build.
func DestinationImageName(source *v1alpha1.Build, space *v1alpha1.Space) string {
	registry := space.Status.BuildConfig.ContainerRegistry

	// Use underscores because those aren't permitted in k8s names so you can't
	// cause accidental conflicts.

	return path.Join(registry, fmt.Sprintf("app_%s_%s:%s", source.Namespace, source.Name, source.UID))
}

// MakeTaskRun creates a Tekton TaskRun for a Build.
func MakeTaskRun(
	build *v1alpha1.Build,
	taskSpec *tektonv1beta1.TaskSpec,
	space *v1alpha1.Space,
	sourcePackage *v1alpha1.SourcePackage,
	secretConfig *config.SecretsConfig,
	defaultsConfig *kfconfig.DefaultsConfig,
) (*tektonv1beta1.TaskRun, error) {
	if taskSpec == nil {
		return nil, errors.New("taskSpec can't be nil")
	}

	// Copy the whole TaskSpec into the TaskRun. This allows us to do two things:
	// 1) Give a historical view about what _exactly_ went into a particular task.
	// 2) Allow us to set a container step template.
	spec := taskSpec.DeepCopy()
	spec.StepTemplate = &tektonv1beta1.StepTemplate{
		Env: build.Spec.Env,
	}

	// WI is used in only in V3 build pipeline. GKE metadata server needs a few seconds to start up.
	// https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#troubleshoot-timeout
	buildPipelineType := build.Spec.BuildTaskRef.Name
	if secretConfig != nil && secretConfig.GoogleServiceAccount != "" && buildPipelineType == v1alpha1.BuildpackV3BuildTaskName {
		initStep := tektonv1beta1.Step{
			Name:    "wait-for-wi",
			Image:   "gcr.io/google.com/cloudsdktool/cloud-sdk:326.0.0-alpine",
			Command: []string{"bash"},
			Args: []string{
				"-c",
				"curl -s -H 'Metadata-Flavor: Google' 'http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token' --retry 30 --retry-connrefused --retry-max-time 30 > /dev/null || exit 1",
			},
		}
		spec.Steps = append([]tektonv1beta1.Step{initStep}, spec.Steps...)
	}

	// Convert to Tekton params.
	var taskParams []tektonv1beta1.Param
	for _, p := range build.Spec.Params {
		taskParams = append(taskParams, p.ToTektonParam())
	}

	// Set the source package if it is present.
	if sourcePackage != nil {
		taskParams = append(
			taskParams,
			tektonv1beta1.Param{
				Name: v1alpha1.SourcePackageNamespaceParamName,
				Value: tektonv1beta1.ArrayOrString{
					Type:      tektonv1beta1.ParamTypeString,
					StringVal: sourcePackage.Namespace,
				},
			},
			tektonv1beta1.Param{
				Name: v1alpha1.SourcePackageNameParamName,
				Value: tektonv1beta1.ArrayOrString{
					Type:      tektonv1beta1.ParamTypeString,
					StringVal: sourcePackage.Name,
				},
			})
	}

	// Add Build-Name task param
	// TODO: apply it only for V2 and Dockerfile for now until V3 are updated.
	if build.Spec.BuildTaskRef.Name == v1alpha1.BuildpackV2BuildTaskName || build.Spec.BuildTaskRef.Name == v1alpha1.DockerfileBuildTaskName {
		taskParams = append(taskParams, tektonv1beta1.Param{
			Name: v1alpha1.BuildNameParamName,
			Value: tektonv1beta1.ArrayOrString{
				Type:      tektonv1beta1.ParamTypeString,
				StringVal: build.Name,
			},
		})
	}

	// The default nodeSelectors is the one specified on the space level.
	nodeSelector := selectorutil.GetNodeSelector(&build.Spec, space)

	// When build node selectors are specified, override node selectors with build node selectors.
	// The buildNodeSelectors is from the config-defaults ConfigMap.
	if defaultsConfig != nil && len(defaultsConfig.BuildNodeSelectors) > 0 {
		nodeSelector = defaultsConfig.BuildNodeSelectors
	}
	podSpec := &tektonv1beta1.PodTemplate{
		NodeSelector: nodeSelector,
	}

	var timeout time.Duration
	// BuildTimeout is set by user at kfsystem, and updated to config-defaults ConfigMap at kf namespace.
	if defaultsConfig != nil && defaultsConfig.BuildTimeout != "" {
		configuredTimeout, err := time.ParseDuration(defaultsConfig.BuildTimeout)
		if err != nil {
			return nil, err
		}
		timeout = configuredTimeout
	} else {
		timeout = v1alpha1.DefaultBuildTimeout
	}

	return &tektonv1beta1.TaskRun{
		ObjectMeta: makeObjectMeta(TaskRunName(build), build, defaultsConfig),
		Spec: tektonv1beta1.TaskRunSpec{
			Timeout:            &metav1.Duration{Duration: timeout},
			ServiceAccountName: space.Status.BuildConfig.ServiceAccount,
			TaskSpec:           spec,
			Params:             taskParams,
			PodTemplate:        podSpec,
			Resources: &tektonv1beta1.TaskRunResources{
				Outputs: []tektonv1beta1.TaskResourceBinding{
					{
						PipelineResourceBinding: tektonv1beta1.PipelineResourceBinding{
							Name: v1alpha1.TaskRunResourceNameImage,
							ResourceSpec: &tektonresource.PipelineResourceSpec{
								Type: "image",
								Params: []tektonv1beta1.ResourceParam{
									{
										Name:  v1alpha1.TaskRunResourceURL,
										Value: DestinationImageName(build, space),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
