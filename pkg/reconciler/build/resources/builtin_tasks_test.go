// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestFindBuiltinTask(t *testing.T) {
	t.Parallel()
	cfg := config.BuiltinDefaultsConfig()

	cases := map[string]struct {
		instantiation   v1alpha1.BuildSpec
		desiredTaskSpec *tektonv1beta1.TaskSpec
	}{
		"buildpackv2": {
			instantiation:   v1alpha1.BuildpackV2Build("source", config.StackV2Definition{}, []string{}, true),
			desiredTaskSpec: buildpackV2Task(cfg),
		},
		"dockerfile": {
			instantiation:   v1alpha1.DockerfileBuild("source", "path/to/Dockerfile"),
			desiredTaskSpec: dockerfileBuildTask(cfg),
		},
		"buildpackv3": {
			instantiation:   v1alpha1.BuildpackV3Build("source", config.StackV3Definition{}, []string{}),
			desiredTaskSpec: buildpackV3Build(cfg, v1alpha1.BuildSpec{}, ""),
		},
		"buildpackv3 with env": {
			instantiation: func() (spec v1alpha1.BuildSpec) {
				spec = v1alpha1.BuildpackV3Build("source", config.StackV3Definition{}, []string{})
				spec.Env = []corev1.EnvVar{
					{Name: "VCAP_SERVICES", Value: "values-ignored"},
					{Name: "VCAP_APPLICATION", Value: "values-ignored"},
				}
				return
			}(),
			desiredTaskSpec: buildpackV3Build(
				cfg,
				v1alpha1.BuildSpec{
					Env: []corev1.EnvVar{
						// order independent
						{Name: "VCAP_APPLICATION"},
						{Name: "VCAP_SERVICES"},
					},
				},
				"",
			),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotTaskSpec := FindBuiltinTask(cfg, tc.instantiation, "")

			// Make sure we found the correct spec
			testutil.AssertEqual(t, "finds-spec", tc.desiredTaskSpec, gotTaskSpec)

			// Check that params line up
			gotParams := sets.NewString()
			for _, p := range tc.instantiation.Params {
				gotParams.Insert(p.Name)
			}

			requiredParams := sets.NewString()
			possibleParams := sets.NewString()
			for _, p := range tc.desiredTaskSpec.Params {
				if p.Default == nil {
					requiredParams.Insert(p.Name)
				}
				possibleParams.Insert(p.Name)
			}

			if missing := requiredParams.Difference(gotParams); missing.Len() > 0 {
				t.Fatalf("Missing required param(s): %v", missing.List())
			}

			if extra := gotParams.Difference(possibleParams); extra.Len() > 0 {
				t.Fatalf("extra param(s): %v", extra.List())
			}
		})
	}
}

func TestBuiltinTaskCoherence(t *testing.T) {
	t.Parallel()
	cfg := config.BuiltinDefaultsConfig()

	expectedQuantity := resource.MustParse("1Gi")
	cfgWithResources := config.BuiltinDefaultsConfig()
	cfgWithResources.BuildPodResources = &corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU: expectedQuantity,
		},
	}

	cases := map[string]struct {
		task                    *tektonv1beta1.TaskSpec
		containersWithResources []string
	}{
		"buildpackv2": {
			task: buildpackV2Task(cfg),
		},
		"buildpackv2 with resources": {
			task:                    buildpackV2Task(cfgWithResources),
			containersWithResources: []string{"run-lifecycle", "build"},
		},
		"dockerfile": {
			task: dockerfileBuildTask(cfg),
		},
		"dockerfile with resources": {
			task:                    dockerfileBuildTask(cfgWithResources),
			containersWithResources: []string{"build"},
		},
		"buildpackv3": {
			task: buildpackV3Build(cfg, v1alpha1.BuildSpec{}, ""),
		},
		"buildpackv3 with resources": {
			task:                    buildpackV3Build(cfgWithResources, v1alpha1.BuildSpec{}, ""),
			containersWithResources: []string{"build"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			knownParams := make(map[string]bool)
			for _, param := range tc.task.Params {
				knownParams[param.Name] = true
			}

			if _, ok := knownParams[v1alpha1.TaskRunParamDestinationImage]; !ok {
				t.Errorf("output image parameter %q not defined", v1alpha1.TaskRunParamDestinationImage)
			}

			for idx, step := range tc.task.Steps {
				t.Run(fmt.Sprintf("steps[%d]", idx), func(t *testing.T) {
					t.Run("image", func(t *testing.T) {
						assertValidTektonParams(t, knownParams, step.Image)
					})

					for argidx, arg := range step.Args {
						t.Run(fmt.Sprintf("args[%d]", argidx), func(t *testing.T) {
							assertValidTektonParams(t, knownParams, arg)
						})
					}

					for cmdidx, cmd := range step.Command {
						t.Run(fmt.Sprintf("args[%d]", cmdidx), func(t *testing.T) {
							assertValidTektonParams(t, knownParams, cmd)
						})
					}
				})
			}

			// valdiate all volumes
			volumeNames := sets.NewString()
			for _, vol := range tc.task.Volumes {
				volumeNames.Insert(vol.Name)
			}

			// Validate each of the expected containers has the resource and
			// the others do not.
			{
				limits := make(map[string]resource.Quantity)
				for _, step := range tc.task.Steps {
					if v, ok := step.Resources.Limits[corev1.ResourceCPU]; ok {
						limits[step.Name] = v
					}
				}

				for _, container := range tc.containersWithResources {
					if _, ok := limits[container]; !ok {
						t.Errorf("step %s should have resources set", container)
					}
					delete(limits, container)
				}

				for k := range limits {
					t.Errorf("step %s should not have resources set", k)
				}
			}

			for _, step := range tc.task.Steps {
				t.Run("step:"+step.Name, func(t *testing.T) {
					// Each step must have a command rather than relying on ENTRYPOINT
					// because Tekton rewrites the comamnd and attempts to resolve
					// them in their controller if unspecified, which means the Tekton
					// controller would need permissions to read from any repository
					// referenced if we didn't require commands.
					// https://github.com/tektoncd/pipeline/blob/0fa0994fe6ef68034842dcf72401d20cfd6057e6/pkg/pod/entrypoint_lookup.go
					testutil.AssertTrue(t, "has command binary", len(step.Command) >= 1)

					// Validate each volume exists
					for _, volume := range step.VolumeMounts {
						if !volumeNames.Has(volume.Name) {
							t.Errorf(
								"volume %s doesn't exist, defined volumes are: %v",
								volume.Name,
								volumeNames.List())
						}
					}
				})
			}
		})
	}
}

func assertValidTektonParams(t *testing.T, params map[string]bool, fieldValue string) {
	t.Helper()

	re, err := regexp.Compile(`\$\(inputs\.params\.(.*?)\)`)
	if err != nil {
		t.Fatal(err)
	}

	matches := re.FindStringSubmatch(fieldValue)

	if len(matches) == 0 {
		return
	}
	// Trim off the full match.
	matches = matches[1:]

	for _, paramName := range matches {
		if _, ok := params[paramName]; !ok {
			t.Errorf("referenced undefined parameter: %q in %q", paramName, fieldValue)
		}
	}
}
