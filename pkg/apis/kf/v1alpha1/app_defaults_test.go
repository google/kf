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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
)

func TestAppSpec_SetDefaults(t *testing.T) {
	t.Parallel()

	// Build
	ctx := apis.WithinUpdate(context.TODO(), &App{
		Spec: AppSpec{
			Build: AppSpecBuild{
				UpdateRequests: 1,
			},
		},
	})

	spec := AppSpec{
		Routes: []RouteWeightBinding{
			// Empty weight
			{},
		},
	}
	spec.SetDefaults(ctx)

	// Build
	{
		// Both build specs are nil and therefore equal. The UpdateRequests
		// should simply be copied.
		testutil.AssertEqual(t, "Build", spec.Build.UpdateRequests, 1)
	}

	// Template
	{
		// AppSpecTemplate.SetDefaults adds a default container if Containers
		// is empty.
		testutil.AssertEqual(t, "Template", len(spec.Template.Spec.Containers), 1)
	}

	// Instances
	{
		// AppSpecInstances.SetDefaults will set an empty Replicas to 1.
		testutil.AssertEqual(t, "Instances", spec.Instances.Replicas, ptr.Int32(1))
	}

	// Routes
	{
		// AppSpecRoutes.SetDefaults will set an empty weight to the default
		// value.
		testutil.AssertEqual(t, "Routes", spec.Routes[0].Weight, ptr.Int32(1))
	}
}

func TestAppSpec_SetDefaults_updateRequests(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		old     *AppSpec
		current AppSpec
		want    AppSpec
	}{
		"update autoincrement": {
			old: &AppSpec{
				Template: AppSpecTemplate{},
			},
			current: AppSpec{
				Template: AppSpecTemplate{},
				Instances: AppSpecInstances{
					Replicas: ptr.Int32(1),
				},
			},
			want: AppSpec{
				Template: AppSpecTemplate{
					UpdateRequests: 1,
				},
			},
		},
		"update with increment": {
			old: &AppSpec{
				Template: AppSpecTemplate{},
			},
			current: AppSpec{
				Template: AppSpecTemplate{
					UpdateRequests: 2,
				},
			},
			want: AppSpec{
				Template: AppSpecTemplate{
					UpdateRequests: 2,
				},
			},
		},
		"update with no change": {
			old: &AppSpec{
				Template: AppSpecTemplate{},
				Instances: AppSpecInstances{
					Replicas: ptr.Int32(1),
				},
			},
			current: AppSpec{
				Template: AppSpecTemplate{
					UpdateRequests: 2,
				},
				Instances: AppSpecInstances{
					Replicas: ptr.Int32(1),
				},
			},
			want: AppSpec{
				Template: AppSpecTemplate{
					UpdateRequests: 2,
				},
				Instances: AppSpecInstances{
					Replicas: ptr.Int32(1),
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := context.TODO()
			if tc.old != nil {
				ctx = apis.WithinUpdate(ctx, &App{
					Spec: *tc.old,
				})
			}

			actual := tc.current
			actual.Template.SetDefaults(ctx, &actual)

			testutil.AssertEqual(t, "defaulted", tc.want.Template.UpdateRequests, actual.Template.UpdateRequests)
		})
	}
}

func TestAppSpec_SetDefaults_BlankContainer(t *testing.T) {
	t.Parallel()

	app := &App{}
	app.SetDefaults(context.Background())

	testutil.AssertEqual(t, "len(spec.template.spec.containers)", 1, len(app.Spec.Template.Spec.Containers))
	testutil.AssertEqual(t, "spec.template.spec.containers.name", DefaultUserContainerName, app.Spec.Template.Spec.Containers[0].Name)
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
							Limits: corev1.ResourceList{
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
	appResourceLimits := app.Spec.Template.Spec.Containers[0].Resources.Limits
	testutil.AssertEqual(t, "default memory request", wantMem, appResourceLimits[corev1.ResourceMemory])
	testutil.AssertEqual(t, "default storage request", wantStorage, appResourceLimits[corev1.ResourceEphemeralStorage])
	testutil.AssertEqual(t, "default CPU request", wantCPU, appResourceLimits[corev1.ResourceCPU])
}

func TestSetKfAppContainerDefaults(t *testing.T) {
	t.Parallel()

	defaultContainer := &corev1.Container{}
	SetKfAppContainerDefaults(context.Background(), defaultContainer)

	cases := map[string]struct {
		template *corev1.Container
		expected *corev1.Container
	}{
		"default everything except ReadinessProbe": {
			template: &corev1.Container{},
			expected: &corev1.Container{
				Name:           DefaultUserContainerName,
				ReadinessProbe: nil,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:              defaultCPU,
						corev1.ResourceMemory:           defaultMem,
						corev1.ResourceEphemeralStorage: defaultStorage,
					},
					Limits: corev1.ResourceList{
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
					ProbeHandler: corev1.ProbeHandler{
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
					ProbeHandler: corev1.ProbeHandler{
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
					ProbeHandler: corev1.ProbeHandler{
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
					ProbeHandler: corev1.ProbeHandler{
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
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("3"),
						corev1.ResourceMemory:           resource.MustParse("3Gi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("3Gi"),
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
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("3"),
						corev1.ResourceMemory:           resource.MustParse("3Gi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("3Gi"),
					},
				},
			},
		},
		"limits set from requests don't get overwritten": {
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
					Limits: corev1.ResourceList{
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

func TestAppSpec_SetBuildDefaults(t *testing.T) {
	t.Parallel()

	buildpackBuildSpec := &BuildSpec{
		BuildTaskRef: buildpackV3BuildTaskRef(),
	}

	dockerfileBuildSpec := &BuildSpec{
		BuildTaskRef: dockerfileBuildTaskRef(),
	}

	cases := map[string]struct {
		old     *AppSpec
		current AppSpec
		want    AppSpec
	}{
		"update autoincrement": {
			old: &AppSpec{
				Build: AppSpecBuild{
					Spec: buildpackBuildSpec,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: dockerfileBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 1,
					Spec:           dockerfileBuildSpec,
				},
			},
		},
		"update with increment": {
			old: &AppSpec{
				Build: AppSpecBuild{
					Spec: buildpackBuildSpec,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 2,
					Spec:           dockerfileBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 2,
					Spec:           dockerfileBuildSpec,
				},
			},
		},
		"update no build change": {
			old: &AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 3,
					Spec:           buildpackBuildSpec,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 3,
					Spec:           buildpackBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 3,
					Spec:           buildpackBuildSpec,
				},
			},
		},
		"update post missing updaterequests": {
			old: &AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 3,
					Spec:           buildpackBuildSpec,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: dockerfileBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 4,
					Spec:           dockerfileBuildSpec,
				},
			},
		},
		"kf restage": {
			old: &AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 3,
					Spec:           buildpackBuildSpec,
				},
			},
			current: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 4,
					Spec:           buildpackBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					UpdateRequests: 4,
					Spec:           buildpackBuildSpec,
				},
			},
		},
		"create": {
			current: AppSpec{
				Build: AppSpecBuild{
					Spec: buildpackBuildSpec,
				},
			},
			want: AppSpec{
				Build: AppSpecBuild{
					Spec: buildpackBuildSpec,
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := context.TODO()
			if tc.old != nil {
				ctx = apis.WithinUpdate(ctx, &App{
					Spec: *tc.old,
				})
			}

			actual := tc.current
			actual.SetBuildDefaults(ctx)

			testutil.AssertEqual(t, "defaulted", tc.want.Build, actual.Build)
		})
	}
}

func TestAppSpec_SetRouteDefaults(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routes []RouteWeightBinding
		want   []RouteWeightBinding
	}{
		"blank defaults": {
			routes: []RouteWeightBinding{
				{},
			},
			want: []RouteWeightBinding{
				{
					RouteSpecFields: RouteSpecFields{
						Path: "/",
					},
					Weight: &defaultRouteWeight,
				},
			},
		},

		"populated gets no defaults": {
			routes: []RouteWeightBinding{
				{
					RouteSpecFields: RouteSpecFields{
						Hostname: "host",
						Domain:   "domain",
						Path:     "/foo",
					},
					Weight: ptr.Int32(42),
				},
			},
			want: []RouteWeightBinding{
				{
					RouteSpecFields: RouteSpecFields{
						Hostname: "host",
						Domain:   "domain",
						Path:     "/foo",
					},
					Weight: ptr.Int32(42),
				},
			},
		},

		"duplicates get merged": {
			routes: []RouteWeightBinding{
				{
					RouteSpecFields: RouteSpecFields{Path: "/foo"},
					Weight:          ptr.Int32(2),
				},
				{
					RouteSpecFields: RouteSpecFields{Path: "/foo"},
					Weight:          ptr.Int32(2),
				},
			},
			want: []RouteWeightBinding{
				{
					RouteSpecFields: RouteSpecFields{Path: "/foo"},
					Weight:          ptr.Int32(4),
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			appSpec := &AppSpec{
				Routes: tc.routes,
			}

			appSpec.SetRouteDefaults(context.Background())

			testutil.AssertEqual(t, "spec", tc.want, appSpec.Routes)
		})
	}
}

func TestAppSpecInstances_SetDefaults(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		spec AppSpecInstances
		want AppSpecInstances
	}{
		"defers to replicas": {
			spec: AppSpecInstances{
				Replicas:          ptr.Int32(99),
				DeprecatedExactly: ptr.Int32(101),
			},
			want: AppSpecInstances{
				Replicas: ptr.Int32(99),
			},
		},
		"copies exactly to replicas": {
			spec: AppSpecInstances{
				Replicas:          nil,
				DeprecatedExactly: ptr.Int32(99),
			},
			want: AppSpecInstances{
				Replicas: ptr.Int32(99),
			},
		},
		"defaults to 1 replica": {
			spec: AppSpecInstances{
				Replicas:          nil,
				DeprecatedExactly: nil,
			},
			want: AppSpecInstances{
				Replicas: ptr.Int32(1),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.spec.SetDefaults(context.Background())
			testutil.AssertEqual(t, "AppSpecInstances", tc.want, tc.spec)
		})
	}
}
