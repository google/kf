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
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeSource(t *testing.T) {

	makePodTemplate := func(env []corev1.EnvVar) v1alpha1.AppSpecTemplate {
		return v1alpha1.AppSpecTemplate{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Env: env,
					},
				},
			},
		}
	}

	makeAppSpecBuild := func(updateRequests int, params []v1alpha1.BuildParam, env []corev1.EnvVar) v1alpha1.AppSpecBuild {
		return v1alpha1.AppSpecBuild{
			UpdateRequests: updateRequests,
			Spec: &v1alpha1.BuildSpec{
				Params: params,
				Env:    env,
			},
		}
	}

	makeSpace := func(env []corev1.EnvVar) v1alpha1.Space {
		return v1alpha1.Space{
			Status: v1alpha1.SpaceStatus{
				BuildConfig: v1alpha1.SpaceStatusBuildConfig{
					Env: env,
				},
			},
		}
	}

	appObjectMeta := metav1.ObjectMeta{
		Name:      "mybuildpackapp",
		Namespace: "myspace",
	}

	cases := map[string]struct {
		app   v1alpha1.App
		space v1alpha1.Space
	}{
		"empty app and space": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					// No container set
					Build: makeAppSpecBuild(1, nil, nil),
				},
			},
			space: makeSpace(nil),
		},
		"buildpack": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Template: makePodTemplate(nil),
					Build: makeAppSpecBuild(37, []v1alpha1.BuildParam{
						{
							Name:  "some-param",
							Value: "cool-value",
						},
					},
						[]corev1.EnvVar{
							{Name: "some-env-var", Value: "cool-env-value"},
						}),
				},
			},
			space: makeSpace(nil),
		},
		"cascading env": {
			space: makeSpace([]corev1.EnvVar{
				{Name: "CASCADE", Value: "space"},
			}),
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Template: makePodTemplate([]corev1.EnvVar{
						{Name: "CASCADE", Value: "app"},
					}),
					Build: makeAppSpecBuild(1, nil, []corev1.EnvVar{
						{Name: "CASCADE", Value: "build"},
					}),
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			build, err := MakeBuild(&tc.app, &tc.space)
			testutil.AssertNil(t, "MakeSource error", err)
			testutil.AssertGoldenJSONContext(t, "build", build, map[string]interface{}{
				"app":   tc.app,
				"space": tc.space,
			})
		})
	}
}
