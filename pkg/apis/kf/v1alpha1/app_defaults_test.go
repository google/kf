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
	"encoding/json"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
)

func TestAppSpec_SetDefaults_BlankContainer(t *testing.T) {
	t.Parallel()

	app := &App{}
	app.SetDefaults(context.Background())

	testutil.AssertEqual(t, "len(spec.template.spec.containers)", 1, len(app.Spec.Template.Spec.Containers))
	testutil.AssertEqual(t, "spec.template.spec.containers.name", "user-container", app.Spec.Template.Spec.Containers[0].Name)
}

func TestAppSpec_SetDefaults_ResourceLimits_AlreadySet(t *testing.T) {
	t.Parallel()

	wantMem := resource.MustParse("2Gi")
	wantStorage := resource.MustParse("2Gi")
	wantCPU := resource.MustParse("2")

	app := &App{
		Spec: AppSpec{
			Template: AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory:           wantMem,
								corev1.ResourceEphemeralStorage: wantStorage,
								corev1.ResourceCPU:              wantCPU,
							},
						},
					}},
				},
			},
		},
	}

	app.SetDefaults(context.Background())

	appResourceRequests := app.Spec.Template.Spec.Containers[0].Resources.Requests
	testutil.AssertEqual(t, "default memory request", wantMem, appResourceRequests[corev1.ResourceMemory])
	testutil.AssertEqual(t, "default storage request", wantStorage, appResourceRequests[corev1.ResourceEphemeralStorage])
	testutil.AssertEqual(t, "default CPU request", wantCPU, appResourceRequests[corev1.ResourceCPU])
}

func TestSetKfAppContainerDefaults(t *testing.T) {
	defaultContainer := &corev1.Container{}
	SetKfAppContainerDefaults(context.Background(), defaultContainer)

	cases := map[string]struct {
		template *corev1.Container
		expected *corev1.Container
	}{
		"default everything": {
			template: &corev1.Container{},
			expected: &corev1.Container{
				Name: "user-container",
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds:   DefaultHealthCheckProbeTimeout,
					PeriodSeconds:    DefaultHealthCheckPeriodSeconds,
					FailureThreshold: DefaultHealthCheckFailureThreshold,
					SuccessThreshold: 1,
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{},
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:              defaultCPU,
						corev1.ResourceMemory:           defaultMem,
						corev1.ResourceEphemeralStorage: defaultStorage,
					},
				},
			},
		},
		"http path gets defaulted": {
			template: &corev1.Container{
				Name: "some-name",
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds:   DefaultHealthCheckProbeTimeout,
					PeriodSeconds:    DefaultHealthCheckPeriodSeconds,
					FailureThreshold: DefaultHealthCheckFailureThreshold,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{},
					},
				},
			},
			expected: &corev1.Container{
				Name: "some-name",
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds:   DefaultHealthCheckProbeTimeout,
					PeriodSeconds:    DefaultHealthCheckPeriodSeconds,
					FailureThreshold: DefaultHealthCheckFailureThreshold,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: DefaultHealthCheckProbeEndpoint},
					},
				},
				Resources: defaultContainer.Resources,
			},
		},
		"full http doesn't get overwritten": {
			template: &corev1.Container{
				Name: "some-name",
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds:   180,
					PeriodSeconds:    180,
					FailureThreshold: 300,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
					},
				},
			},
			expected: &corev1.Container{
				Name: "some-name",
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds:   180,
					PeriodSeconds:    180,
					FailureThreshold: 300,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
					},
				},
				Resources: defaultContainer.Resources,
			},
		},
		"resources don't get overwritten": {
			template: &corev1.Container{
				Name: "some-name",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("2"),
						corev1.ResourceMemory:           resource.MustParse("2Gi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
					},
				},
			},
			expected: &corev1.Container{
				Name:           "some-name",
				ReadinessProbe: defaultContainer.ReadinessProbe,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("2"),
						corev1.ResourceMemory:           resource.MustParse("2Gi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			SetKfAppContainerDefaults(context.TODO(), tc.template)

			testutil.AssertEqual(t, "expected", tc.expected, tc.template)
		})
	}
}

func TestAppSpec_SetSourceDefaults(t *testing.T) {
	cases := map[string]struct {
		old     *SourceSpec
		current SourceSpec
		want    SourceSpec
	}{
		"update autoincrement": {
			old: &SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			current: SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
			want: SourceSpec{
				UpdateRequests: 1,
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
		},
		"update with increment": {
			old: &SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			current: SourceSpec{
				UpdateRequests: 2,
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
			want: SourceSpec{
				UpdateRequests: 2,
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
		},
		"update no source change": {
			old: &SourceSpec{
				UpdateRequests: 3,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			current: SourceSpec{
				UpdateRequests: 3,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			want: SourceSpec{
				UpdateRequests: 3,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
		},
		"update post missing updaterequests": {
			old: &SourceSpec{
				UpdateRequests: 3,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			current: SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
			want: SourceSpec{
				UpdateRequests: 4,
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
		},
		"kf restage": {
			old: &SourceSpec{
				UpdateRequests: 3,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			current: SourceSpec{
				UpdateRequests: 4,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
			want: SourceSpec{
				UpdateRequests: 4,
				ContainerImage: SourceSpecContainerImage{Image: "mysql"},
			},
		},
		"create": {
			current: SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
			want: SourceSpec{
				ContainerImage: SourceSpecContainerImage{Image: "sqlite3"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := context.TODO()
			if tc.old != nil {
				ctx = apis.WithinUpdate(ctx, &App{
					Spec: AppSpec{
						Source: *tc.old,
					},
				})
			}

			actual := &AppSpec{Source: tc.current}
			actual.SetSourceDefaults(ctx)

			testutil.AssertEqual(t, "defaulted", tc.want, actual.Source)
		})
	}
}

func TestAppSpec_SetServiceBindingsDefaults(t *testing.T) {
	cases := map[string]struct {
		old     *[]AppSpecServiceBinding
		current []AppSpecServiceBinding
		want    []AppSpecServiceBinding
	}{
		"binding": {
			current: []AppSpecServiceBinding{
				{
					Instance: "instance",
				},
			},
			want: []AppSpecServiceBinding{
				{
					BindingName: "instance",
					Instance:    "instance",
					Parameters:  json.RawMessage("null"),
				},
			},
		},
		"binding update": {
			old: &[]AppSpecServiceBinding{
				{
					BindingName: "some-binding",
					Instance:    "instance",
					Parameters:  json.RawMessage("null"),
				},
			},
			current: []AppSpecServiceBinding{
				{
					Instance: "instance",
				},
			},
			want: []AppSpecServiceBinding{
				{
					BindingName: "instance",
					Instance:    "instance",
					Parameters:  json.RawMessage("null"),
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := context.TODO()
			if tc.old != nil {
				ctx = apis.WithinUpdate(ctx, &App{
					Spec: AppSpec{
						ServiceBindings: *tc.old,
					},
				})
			}

			actual := &AppSpec{ServiceBindings: tc.current}
			actual.SetServiceBindingDefaults(ctx)

			testutil.AssertEqual(t, "defaulted", tc.want, actual.ServiceBindings)
		})
	}
}
