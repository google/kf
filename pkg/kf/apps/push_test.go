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

package apps_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	appsfake "github.com/google/kf/v2/pkg/kf/apps/fake"
	secretsfake "github.com/google/kf/v2/pkg/kf/secrets/fake"
	bindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	sourcepackagesfake "github.com/google/kf/v2/pkg/kf/sourcepackages/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/ptr"
)

func TestPush_Logs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName        string
		containerImage *string
		wantErr        error
		logErr         error
		noStart        bool
	}{
		"fetching logs succeeds": {
			appName: "some-app",
		},
		"NoStart gets passed through": {
			appName: "some-app",
			noStart: true,
		},
		"fetching logs returns an error, no error": {
			appName: "some-app",
			wantErr: errors.New("some error"),
			logErr:  errors.New("some error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expectedNamespace := "some-space"

			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeBindings := bindingsfake.NewFakeClient(ctrl)
			fakeSecrets := secretsfake.NewFakeClient(ctrl)

			mockApp := &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:            tc.appName,
					Namespace:       expectedNamespace,
					ResourceVersion: tc.appName + "-version",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Stopped: tc.noStart,
					},
				},
			}

			fakeApps.EXPECT().
				Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
				Return(mockApp, nil)

			fakeApps.EXPECT().
				DeployLogsForApp(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any()).Do(func(ctx context.Context, _ io.Writer, app *v1alpha1.App) {
				testutil.AssertEqual(t, "Name", tc.appName, app.Name)
				testutil.AssertEqual(t, "ResourceVersion", tc.appName+"-version", app.ResourceVersion)
				testutil.AssertEqual(t, "Namespace", expectedNamespace, app.Namespace)
				testutil.AssertEqual(t, "resourceVersion", tc.noStart, app.Spec.Instances.Stopped)
			}).Return(tc.logErr)

			fakeApps.EXPECT().
				Get(gomock.Any(), expectedNamespace, gomock.Any()).
				Return(mockApp, nil).
				AnyTimes()

			fakeSourcePackages := sourcepackagesfake.NewFakeClient(ctrl)

			p := apps.NewPusher(
				fakeApps,
				fakeBindings,
				fakeSecrets,
				fakeSourcePackages,
				nil,
			)

			build := bldPtr(v1alpha1.DockerfileBuild("some-image", "path/to/Dockerfile"))

			gotErr := p.Push(
				context.Background(),
				tc.appName,
				apps.WithPushOutput(&bytes.Buffer{}),
				apps.WithPushBuild(build),
				apps.WithPushContainerImage(tc.containerImage),
				apps.WithPushSpace(expectedNamespace),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Stopped: tc.noStart}),
			)

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

		})
	}
}

func TestPush(t *testing.T) {
	// XXX: Unfortunately, the FakeDynamicClient panics when it is given an
	// object that isn't purely primitives. The real dynamic client is far
	// more permissive. This makes testing rough... We're going to lean on
	// integration tests for now.
	t.Parallel()

	appWithSourcePackage := &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-app",
			Namespace: "default",
		},
		Spec: v1alpha1.AppSpec{
			Build: v1alpha1.AppSpecBuild{
				Spec: &v1alpha1.BuildSpec{
					SourcePackage: corev1.LocalObjectReference{
						Name: "some-source-package",
					},
				},
			},
		},
	}

	type fakes struct {
		appsClient           *appsfake.FakeClient
		sourcePackagesClient *sourcepackagesfake.FakeClient
	}

	fakeImage := "some-image"

	fakeStack := config.StackV3Definition{
		Name:       "some-stack",
		BuildImage: "gcr.io/google/builder",
		RunImage:   "gcr.io/google/runner",
	}

	fakeBuildpack := "some-buildpack"

	fakeBuild := v1alpha1.BuildpackV3Build(fakeImage, fakeStack, []string{fakeBuildpack})

	for tn, tc := range map[string]struct {
		appName   string
		srcImage  string
		buildpack string
		opts      apps.PushOptions
		setup     func(t *testing.T, f *fakes)
		assert    func(t *testing.T, err error)
	}{
		"pushes app to a configured space": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
				apps.WithPushSpace("some-space"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "space", "some-space", newApp.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app to default space": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "space", "default", newApp.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with exact instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Replicas: ptr.Int32(9)}),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Instances.Replicas = ptr.Int32(9)
						newApp = merge(newApp, oldApp)
						testutil.Assert(t, gomock.Not(gomock.Nil()), newApp.Spec.Instances.Replicas)
						testutil.AssertEqual(t, "instances.Replicas", int32(9), *newApp.Spec.Instances.Replicas)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app but leaves exact instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Instances.Replicas = ptr.Int32(9)
						newApp = merge(newApp, oldApp)
						testutil.AssertEqual(t, "instances.Replicas", int32(9), *newApp.Spec.Instances.Replicas)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with default of exactly 1 instance": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						newApp = merge(newApp, oldApp)
						testutil.AssertEqual(t, "instances.Replicas", int32(1), *newApp.Spec.Instances.Replicas)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with buildpack": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
			},
		},
		"pushes app with proper Service config": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
				apps.WithPushSpace("myns"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "service.Name", "some-app", newApp.Name)
						testutil.AssertEqual(t, "service.Kind", "App", newApp.Kind)
						testutil.AssertEqual(t, "service.APIVersion", "kf.dev/v1alpha1", newApp.APIVersion)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"properly configures buildpackBuild": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushBuild(&fakeBuild),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "space", "default", newApp.Namespace)

						var buildpack *string
						for _, p := range newApp.Spec.Build.Spec.Params {
							if p.Name == "BUILDPACK" {
								s := p.Value
								buildpack = &s
							}
						}

						testutil.AssertNotNil(t, "buildpack", buildpack)
						testutil.AssertEqual(t, "buildpack", "some-buildpack", *buildpack)

					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes a container image": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushContainerImage(&fakeImage),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertNotNil(t, "containerImage", newApp.Spec.Build.Image)
						testutil.AssertEqual(t, "containerImage", "some-image", *newApp.Spec.Build.Image)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes a docker build": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushBuild(bldPtr(v1alpha1.DockerfileBuild("some-image", "path/to/Dockerfile"))),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {

						var path *string

						for _, p := range newApp.Spec.Build.Spec.Params {
							if p.Name == "DOCKERFILE" {
								s := p.Value
								path = &s
							}
						}

						testutil.AssertNotNil(t, "path", path)
						testutil.AssertEqual(t, "path", "path/to/Dockerfile", *path)

					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with routes": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushRoutes([]v1alpha1.RouteWeightBinding{
					{
						Weight: ptr.Int32(1),
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "host-1",
						},
					},
					{
						Weight: ptr.Int32(1),
						RouteSpecFields: v1alpha1.RouteSpecFields{
							Hostname: "host-2",
						},
					},
				}),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "Routes", []v1alpha1.RouteWeightBinding{
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Hostname: "host-1",
								},
							},
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Hostname: "host-2",
								},
							},
						}, newApp.Spec.Routes)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"pushes app with default route": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushGenerateDefaultRoute(true),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newObj *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "len(Routes)", 1, len(newObj.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "", newObj.Spec.Routes[0].Domain)
						testutil.AssertEqual(t, "Routes.Hostname", "some-app", newObj.Spec.Routes[0].Hostname)
						testutil.AssertEqual(t, "Routes.Path", "", newObj.Spec.Routes[0].Path)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"default route does not overwrite existing route": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushGenerateDefaultRoute(true),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newObj *v1alpha1.App, merge apps.Merger) {
						app := &v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteWeightBinding{
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Domain: "existing.com",
								},
							},
						}
						app = merge(app, app)

						testutil.AssertEqual(t, "len(Routes)", 1, len(app.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "existing.com", app.Spec.Routes[0].Domain)
						testutil.AssertEqual(t, "Routes.Hostname", "", app.Spec.Routes[0].Hostname)
						testutil.AssertEqual(t, "Routes.Path", "", app.Spec.Routes[0].Path)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"pushes app with random route": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushGenerateRandomRoute(true),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						newApp = merge(newApp, newApp)
						testutil.AssertEqual(t, "len(Routes)", 1, len(newApp.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "", newApp.Spec.Routes[0].Domain)
						testutil.Assert(t, gomock.Not(gomock.Eq("")), newApp.Spec.Routes[0].Hostname)
						testutil.AssertEqual(t, "Routes.Path", "", newApp.Spec.Routes[0].Path)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"random route does not overwrite existing route": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushGenerateRandomRoute(true),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := &v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteWeightBinding{
							{
								Weight: ptr.Int32(1),
								RouteSpecFields: v1alpha1.RouteSpecFields{
									Domain: "existing.com",
								},
							},
						}
						app = merge(app, app)

						testutil.AssertEqual(t, "len(Routes)", 1, len(app.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "existing.com", app.Spec.Routes[0].Domain)
						testutil.AssertEqual(t, "Routes.Hostname", "", app.Spec.Routes[0].Hostname)
						testutil.AssertEqual(t, "Routes.Path", "", app.Spec.Routes[0].Path)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"random route does not exceed length": {
			appName: "some-app-which-has-a-long-name-whoops",
			opts: apps.PushOptions{
				apps.WithPushGenerateRandomRoute(true),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						newApp = merge(newApp, newApp)
						testutil.AssertTrue(t, "hostname length limit", len(newApp.Spec.Routes[0].Hostname) < 64)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"deployer returns an error": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to push App: some-error"), err)
			},
		},
		"NoStart sets stopped": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushAppSpecInstances(v1alpha1.AppSpecInstances{Stopped: true}),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "app.Spec.Instances.Stopped", true, newApp.Spec.Instances.Stopped)

					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"increments Spec.Template.UpdateRequests": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushBuild(&fakeBuild),
				apps.WithPushSpace("some-namespace"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Template.UpdateRequests = 99
						merge(newApp, oldApp)

						testutil.AssertEqual(
							t,
							"Spec.Template.UpdateRequests",
							100,
							newApp.Spec.Template.UpdateRequests,
						)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"disk quota uses manifest value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {

				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := setupResourcedApp()
						result := merge(app, appWithContainer())
						assertOnResource(t, "DiskQuota", corev1.ResourceEphemeralStorage, "1Gi", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"disk quota uses existing value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {

				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := appWithContainer()
						oldApp := setupResourcedApp()
						result := merge(app, oldApp)
						assertOnResource(t, "DiskQuota", corev1.ResourceEphemeralStorage, "1Gi", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"memory limit uses manifest value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {

				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := setupResourcedApp()
						result := merge(app, appWithContainer())
						assertOnResource(t, "MemoryLimit", corev1.ResourceMemory, "1Gi", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"memory limit uses existing value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {

				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := appWithContainer()
						oldApp := setupResourcedApp()
						result := merge(app, oldApp)
						assertOnResource(t, "MemoryLimit", corev1.ResourceMemory, "1Gi", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"cpu cores uses manifest value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {

				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := setupResourcedApp()
						result := merge(app, appWithContainer())
						assertOnResource(t, "CPU Cores", corev1.ResourceCPU, "1", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"cpu cores uses existing value": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						app := appWithContainer()
						oldApp := setupResourcedApp()
						result := merge(app, oldApp)
						assertOnResource(t, "CPU Cores", corev1.ResourceCPU, "1", result)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"App's BuildSpec references the SourcePackage": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushSourcePath("testdata"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) (*v1alpha1.App, error) {
						// Ensure the newApp has the SourcePackage name set.
						testutil.AssertEqual(t, "source package", "some-app-0", newApp.Spec.Build.Spec.SourcePackage.Name)

						oldApp := setupResourcedApp()
						result := merge(newApp, oldApp)
						testutil.AssertEqual(t, "source package", "some-app-1", result.Spec.Build.Spec.SourcePackage.Name)
						return result, nil
					})

				f.sourcePackagesClient.EXPECT().
					UploadSourcePath(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"uploads data to the uploads API server fails": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushSourcePath("testdata"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(appWithSourcePackage, nil)

				f.sourcePackagesClient.EXPECT().
					UploadSourcePath(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		"uploads data to the uploads API server": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushSourcePath("testdata"),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(appWithSourcePackage, nil)

				f.sourcePackagesClient.EXPECT().
					UploadSourcePath(
						gomock.Any(),
						gomock.Not(gomock.Eq("")),
						gomock.Any(),
					)
			},
		},
		"annotations-are-canonical": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushAnnotations(map[string]string{
					"new1k": "new1v",
				}),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Annotations = map[string]string{
							"old1k": "old1v",
						}

						merge(newApp, oldApp)

						testutil.AssertEqual(
							t,
							"annotations",
							map[string]string{
								"new1k": "new1v",
							},
							newApp.Annotations,
						)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"labels-are-canonical": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSpace("default"),
				apps.WithPushLabels(map[string]string{
					"new1k": "new1v",
				}),
			},
			setup: func(t *testing.T, f *fakes) {
				f.appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, space string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Labels = map[string]string{
							"old1k": "old1v",
						}

						merge(newApp, oldApp)

						testutil.AssertEqual(
							t,
							"labels",
							map[string]string{
								"new1k": "new1v",
							},
							newApp.Labels,
						)
					}).Return(&v1alpha1.App{}, nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.assert == nil {
				tc.assert = func(t *testing.T, err error) {}
			}
			if tc.setup == nil {
				tc.setup = func(t *testing.T, f *fakes) {
					f.appsClient.EXPECT().
						Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
						Return(&v1alpha1.App{}, nil)
				}
			}

			ctrl := gomock.NewController(t)
			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeBindings := bindingsfake.NewFakeClient(ctrl)
			fakeSecrets := secretsfake.NewFakeClient(ctrl)

			fakeApps.EXPECT().
				DeployLogsForApp(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeApps.EXPECT().
				Get(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&v1alpha1.App{}, nil).
				AnyTimes()

			fakeSourcePackages := sourcepackagesfake.NewFakeClient(ctrl)

			f := &fakes{
				appsClient:           fakeApps,
				sourcePackagesClient: fakeSourcePackages,
			}

			tc.setup(t, f)

			p := apps.NewPusher(
				fakeApps,
				fakeBindings,
				fakeSecrets,
				fakeSourcePackages,
				nil,
			)
			opts := append([]apps.PushOption{apps.WithPushOutput(&bytes.Buffer{})}, tc.opts...)
			gotErr := p.Push(context.Background(), tc.appName, opts...)
			tc.assert(t, gotErr)
			if gotErr != nil {
				return
			}

		})
	}
}

func TestPush_ServiceInstanceBindings(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName   string
		srcImage  string
		buildpack string
		opts      apps.PushOptions
		setup     func(t *testing.T, appsClient *appsfake.FakeClient, bindingsClient *bindingsfake.FakeClient, secretsClient *secretsfake.FakeClient)
		assert    func(t *testing.T, err error)
	}{
		"does not overwrite existing ServiceBindings": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushServiceBindings([]v1alpha1.ServiceInstanceBinding{createServiceInstanceBinding("some-app", "some-service", "some-space")}),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient, bindingsClient *bindingsfake.FakeClient, secretsClient *secretsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.App{}, nil)

				bindingsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.ServiceInstanceBinding{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"uses new ServiceBindings": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushServiceBindings([]v1alpha1.ServiceInstanceBinding{createServiceInstanceBinding("some-app", "some-service", "some-space")}),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient, bindingsClient *bindingsfake.FakeClient, secretsClient *secretsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.App{}, nil)

				bindingsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, apierrs.NewNotFound(v1alpha1.Resource("serviceinstancebinding"), gomock.Any().String()))
				bindingsClient.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.ServiceInstanceBinding{}, nil)
				secretsClient.EXPECT().CreateParamsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&v1.Secret{}, nil)
				bindingsClient.EXPECT().
					WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.ServiceInstanceBinding{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.assert == nil {
				tc.assert = func(t *testing.T, err error) {}
			}
			if tc.setup == nil {
				tc.setup = func(t *testing.T, appsClient *appsfake.FakeClient, bindingsClient *bindingsfake.FakeClient, secretsClient *secretsfake.FakeClient) {
					appsClient.EXPECT().
						Upsert(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
						Return(&v1alpha1.App{}, nil)
				}
			}

			ctrl := gomock.NewController(t)
			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeBindings := bindingsfake.NewFakeClient(ctrl)
			fakeSecrets := secretsfake.NewFakeClient(ctrl)

			fakeApps.EXPECT().
				DeployLogsForApp(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeApps.EXPECT().
				Get(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&v1alpha1.App{}, nil).
				AnyTimes()

			fakeSourcePackages := sourcepackagesfake.NewFakeClient(ctrl)

			tc.setup(t, fakeApps, fakeBindings, fakeSecrets)

			p := apps.NewPusher(
				fakeApps,
				fakeBindings,
				fakeSecrets,
				fakeSourcePackages,
				nil,
			)
			opts := append([]apps.PushOption{apps.WithPushOutput(&bytes.Buffer{})}, tc.opts...)
			gotErr := p.Push(context.Background(), tc.appName, opts...)
			tc.assert(t, gotErr)
			if gotErr != nil {
				return
			}

		})
	}
}

func createServiceInstanceBinding(appName, serviceInstance, namespace string) v1alpha1.ServiceInstanceBinding {
	return v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1alpha1.MakeServiceBindingName(appName, serviceInstance),
			Namespace: namespace,
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: appName,
				},
			},
			InstanceRef: v1.LocalObjectReference{
				Name: serviceInstance,
			},
			ParametersFrom: v1.LocalObjectReference{
				Name: v1alpha1.MakeServiceBindingParamsSecretName(appName, serviceInstance),
			},
		},
	}
}

func appWithContainer() *v1alpha1.App {
	app := &v1alpha1.App{}
	app.Spec.Template.Spec.Containers = []corev1.Container{{}}
	return app
}

func setupResourcedApp() *v1alpha1.App {
	resourceList := corev1.ResourceList{
		corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
		corev1.ResourceMemory:           resource.MustParse("1Gi"),
		corev1.ResourceCPU:              resource.MustParse("1"),
	}
	app := appWithContainer()
	app.Spec.Template.Spec.Containers[0].Resources.Requests = resourceList
	return app
}

func getResource(app *v1alpha1.App, k corev1.ResourceName) *resource.Quantity {
	containers := app.Spec.Template.Spec.Containers

	if len(containers) == 0 {
		return nil
	}

	v, ok := containers[0].Resources.Requests[k]
	if !ok {
		return nil
	}
	return &v
}

func assertOnResource(
	t *testing.T,
	name string,
	r corev1.ResourceName,
	expectedQuantity string,
	actual *v1alpha1.App,
) {
	t.Helper()

	v := actual.Spec.Template.Spec.Containers[0].Resources.Requests[r]
	testutil.AssertEqual(t, name, resource.MustParse(expectedQuantity), v)
}

func bldPtr(build v1alpha1.BuildSpec) *v1alpha1.BuildSpec {
	return &build
}

func TestJoinRepositoryImage(t *testing.T) {
	cases := map[string]struct {
		repo  string
		image string
		want  string
	}{
		"normal": {
			repo:  "gcr.io/myrepo",
			image: "src-my-app:latest",
			want:  "gcr.io/myrepo/src-my-app:latest",
		},
		"trailing slash": {
			repo:  "gcr.io/myrepo/",
			image: "src-my-app:latest",
			want:  "gcr.io/myrepo/src-my-app:latest",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.AssertEqual(
				t,
				"image",
				tc.want,
				apps.JoinRepositoryImage(tc.repo, tc.image),
			)
		})
	}
}

func Test_pusher_CreatePlaceholderApp(t *testing.T) {
	t.Parallel()

	type mocks struct {
		appsClient *appsfake.FakeClient
	}

	var (
		mockAppName      = "testapp"
		mockAppNamespace = "default"
		mockPushOptions  = []apps.PushOption{
			apps.WithPushSpace(mockAppNamespace),
		}
		notFoundError = apierrs.NewNotFound(schema.GroupResource{}, "")
	)

	cases := map[string]struct {
		// Command args
		appName string
		opts    []apps.PushOption

		// Environment mocks and assertions
		setup   func(t *testing.T, mocks *mocks)
		wantErr error
	}{
		"does nothing if App exists": {
			appName: mockAppName,
			opts:    mockPushOptions,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.appsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&v1alpha1.App{}, nil).
					Times(1)
			},
		},
		"creates App if not exists": {
			appName: mockAppName,
			opts:    mockPushOptions,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.appsClient.EXPECT().
					Get(gomock.Any(), mockAppNamespace, mockAppName).
					Return(nil, notFoundError).
					Times(1)

				mocks.appsClient.EXPECT().
					Create(gomock.Any(), mockAppNamespace, gomock.Any()).
					Return(nil, nil).
					Times(1)

				mocks.appsClient.EXPECT().
					WaitForConditionReadyTrue(gomock.Any(), mockAppNamespace, mockAppName, gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
		},
		"creates a valid placeholder App": {
			appName: mockAppName,
			opts:    mockPushOptions,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.appsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, notFoundError).
					AnyTimes()

				mocks.appsClient.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_, _ interface{}, app *v1alpha1.App) (interface{}, interface{}) {
						app.SetDefaults(context.TODO())
						aerr := app.Validate(context.TODO())
						testutil.AssertEqual(t, "validation errors", "", aerr.Error())

						return app, nil
					}).MinTimes(1)

				mocks.appsClient.EXPECT().
					WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil).
					AnyTimes()
			},
		},
		"fails if App wait fails": {
			appName: mockAppName,
			opts:    mockPushOptions,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.appsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, notFoundError).
					AnyTimes()

				mocks.appsClient.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil).
					AnyTimes()

				mocks.appsClient.EXPECT().
					WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("test failure")).
					AnyTimes()
			},
			wantErr: errors.New("couldn't wait for App placeholder: test failure"),
		},
		"fails if placeholder can't be created": {
			appName: mockAppName,
			opts:    mockPushOptions,
			setup: func(t *testing.T, mocks *mocks) {
				mocks.appsClient.EXPECT().
					Get(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, notFoundError).
					AnyTimes()

				mocks.appsClient.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("test failure")).
					AnyTimes()
			},
			wantErr: errors.New("couldn't create App placeholder: test failure"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			m := &mocks{
				appsClient: appsfake.NewFakeClient(ctrl),
			}

			p := apps.NewPusher(
				m.appsClient,
				nil,
				nil,
				nil,
				nil,
			)

			tc.setup(t, m)

			opts := append([]apps.PushOption{apps.WithPushOutput(&bytes.Buffer{})}, tc.opts...)
			_, gotErr := p.CreatePlaceholderApp(context.TODO(), tc.appName, opts...)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
		})
	}
}
