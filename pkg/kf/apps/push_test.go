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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/apps"
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPush_Logs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName        string
		srcImage       string
		containerImage string
		wantErr        error
		logErr         error
		noStart        bool
	}{
		"fetching logs succeeds": {
			appName:  "some-app",
			srcImage: "some-image",
		},
		"NoStart gets passed through": {
			appName:  "some-app",
			srcImage: "some-image",
			noStart:  true,
		},
		"fetching logs returns an error, no error": {
			appName:  "some-app",
			srcImage: "some-image",
			wantErr:  errors.New("some error"),
			logErr:   errors.New("some error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expectedNamespace := "some-namespace"

			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeApps.EXPECT().
				Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
				Return(&v1alpha1.App{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: tc.appName + "-version",
					},
				}, nil)
			fakeApps.EXPECT().
				DeployLogs(
					gomock.Not(gomock.Nil()), // out,
					tc.appName,               // appName
					tc.appName+"-version",    // resourceVersion
					expectedNamespace,        // namespace
					tc.noStart,               // NoStart
				).
				Return(tc.logErr)

			p := apps.NewPusher(
				fakeApps,
			)

			gotErr := p.Push(
				tc.appName,
				apps.WithPushSourceImage(tc.srcImage),
				apps.WithPushContainerImage(tc.containerImage),
				apps.WithPushNamespace(expectedNamespace),
				apps.WithPushContainerRegistry("some-container-registry"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushNoStart(tc.noStart),
			)

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			ctrl.Finish()
		})
	}
}

func TestPush(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		appName   string
		srcImage  string
		buildpack string
		opts      apps.PushOptions
		setup     func(t *testing.T, appsClient *appsfake.FakeClient)
		assert    func(t *testing.T, err error)
	}{
		"pushes app to a configured namespace": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushNamespace("some-namespace"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "some-namespace", newApp.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app to default namespace": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "default", newApp.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with exact instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushExactScale(intPtr(9)),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Instances.Exactly = intPtr(9)
						newApp = merge(newApp, oldApp)
						testutil.Assert(t, gomock.Not(gomock.Nil()), newApp.Spec.Instances.Exactly)
						testutil.AssertEqual(t, "instances.Exactly", 9, *newApp.Spec.Instances.Exactly)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app but leaves exact instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Instances.Exactly = intPtr(9)
						newApp = merge(newApp, oldApp)
						testutil.AssertEqual(t, "instances.Exactly", 9, *newApp.Spec.Instances.Exactly)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with min and max instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushMinScale(intPtr(9)),
				apps.WithPushMaxScale(intPtr(11)),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "instances.Min", 9, *newApp.Spec.Instances.Min)
						testutil.AssertEqual(t, "instances.Max", 11, *newApp.Spec.Instances.Max)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app but leaves min and max instances": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						oldApp.Spec.Instances.Min = intPtr(9)
						oldApp.Spec.Instances.Max = intPtr(11)
						newApp = merge(newApp, oldApp)
						testutil.AssertEqual(t, "instances.Min", 9, *newApp.Spec.Instances.Min)
						testutil.AssertEqual(t, "instances.Max", 11, *newApp.Spec.Instances.Max)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with default of exactly 1 instance": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						oldApp := &v1alpha1.App{}
						newApp = merge(newApp, oldApp)
						testutil.AssertEqual(t, "instances.Exactly", 1, *newApp.Spec.Instances.Exactly)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with buildpack": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushBuildpack("some-buildpack"),
			},
		},
		"pushes app with proper Service config": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushNamespace("myns"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						ka := apps.NewFromApp(newApp)

						testutil.AssertEqual(t, "service.Name", "some-app", newApp.Name)
						testutil.AssertEqual(t, "service.Kind", "App", newApp.Kind)
						testutil.AssertEqual(t, "service.APIVersion", "kf.dev/v1alpha1", newApp.APIVersion)
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", ka.GetServiceAccount())
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"properly configures buildpackBuild source": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushNamespace("default"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
				apps.WithPushBuildpack("some-buildpack"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "default", newApp.Namespace)
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", newApp.Spec.Template.Spec.ServiceAccountName)
						testutil.AssertEqual(t, "image", "some-image", newApp.Spec.Source.BuildpackBuild.Source)
						testutil.AssertEqual(t, "buildpack", "some-buildpack", newApp.Spec.Source.BuildpackBuild.Buildpack)

					}).Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with environment variables": {
			appName:   "some-app",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushEnvironmentVariables(map[string]string{"ENV1": "val1", "ENV2": "val2"}),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						actual := envutil.GetAppEnvVars(newApp)
						envutil.SortEnvVars(actual)
						testutil.AssertEqual(t, "envs",
							[]corev1.EnvVar{{Name: "ENV1", Value: "val1"}, {Name: "ENV2", Value: "val2"}},
							actual,
						)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes a container image": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushContainerImage("some-image"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "containerImage", "some-image", newApp.Spec.Source.ContainerImage.Image)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with routes": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushRoutes([]v1alpha1.RouteSpecFields{{Hostname: "host-1"}, {Hostname: "host-2"}}),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "Routes", []v1alpha1.RouteSpecFields{
							{Hostname: "host-1"}, {Hostname: "host-2"},
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
				apps.WithPushDefaultRouteDomain("example.com"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "len(Routes)", 1, len(newObj.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "example.com", newObj.Spec.Routes[0].Domain)
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
				apps.WithPushDefaultRouteDomain("example.com"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						app := &v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteSpecFields{
							{Domain: "existing.com"},
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
				apps.WithPushRandomRouteDomain("example.com"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						newApp = merge(newApp, newApp)
						testutil.AssertEqual(t, "len(Routes)", 1, len(newApp.Spec.Routes))
						testutil.AssertEqual(t, "Routes.Domain", "example.com", newApp.Spec.Routes[0].Domain)
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
				apps.WithPushRandomRouteDomain("example.com"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						app := &v1alpha1.App{}
						app.Spec.Routes = []v1alpha1.RouteSpecFields{
							{Domain: "existing.com"},
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
		"deployer returns an error": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to push app: some-error"), err)
			},
		},
		"set ports to h2c for gRPC": {
			appName: "some-app",
			opts: apps.PushOptions{
				apps.WithPushSourceImage("some-image"),
				apps.WithPushContainerRegistry("some-reg.io"),
				apps.WithPushGrpc(true),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						ka := apps.NewFromApp(newApp)

						testutil.AssertEqual(
							t,
							"container.ports",
							[]corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}},
							ka.GetContainerPorts(),
						)
					}).
					Return(&v1alpha1.App{}, nil)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"NoStart sets stopped": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: apps.PushOptions{
				apps.WithPushNamespace("default"),
				apps.WithPushNoStart(true),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newApp *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "app.Spec.Instances.Stopped", true, newApp.Spec.Instances.Stopped)

					}).Return(&v1alpha1.App{}, nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.assert == nil {
				tc.assert = func(t *testing.T, err error) {}
			}
			if tc.setup == nil {
				tc.setup = func(t *testing.T, appsClient *appsfake.FakeClient) {
					appsClient.EXPECT().
						Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
						Return(&v1alpha1.App{}, nil)
				}
			}

			ctrl := gomock.NewController(t)
			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeApps.EXPECT().
				DeployLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			tc.setup(t, fakeApps)

			p := apps.NewPusher(fakeApps)
			gotErr := p.Push(tc.appName, tc.opts...)
			tc.assert(t, gotErr)
			if gotErr != nil {
				return
			}

			ctrl.Finish()
		})
	}
}

func intPtr(i int) *int {
	return &i
}
