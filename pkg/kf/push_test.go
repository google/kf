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

package kf_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/apps"
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	kffake "github.com/google/kf/pkg/kf/fake"
	"github.com/google/kf/pkg/kf/internal/envutil"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPush_Logs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName  string
		srcImage string
		wantErr  error
		logErr   error
	}{
		"fetching logs succeeds": {
			appName:  "some-app",
			srcImage: "some-image",
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

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.EXPECT().
				DeployLogs(
					gomock.Not(gomock.Nil()), // out,
					tc.appName,               // appName
					tc.appName+"-version",    // resourceVersion
					expectedNamespace,        // namespace
				).
				Return(tc.logErr)

			fakeApps := appsfake.NewFakeClient(ctrl)
			fakeApps.EXPECT().
				Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
				Return(&v1alpha1.App{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: tc.appName + "-version",
					},
				}, nil)

			p := kf.NewPusher(
				fakeLogs,
				fakeApps,
			)

			gotErr := p.Push(
				tc.appName,
				tc.srcImage,
				kf.WithPushNamespace(expectedNamespace),
				kf.WithPushContainerRegistry("some-container-registry"),
				kf.WithPushServiceAccount("some-service-account"),
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
		opts      kf.PushOptions
		setup     func(t *testing.T, appsClient *appsfake.FakeClient)
		assert    func(t *testing.T, err error)
	}{
		"pushes app to a configured namespace": {
			appName:  "some-app",
			srcImage: "some-image",
			opts: kf.PushOptions{
				kf.WithPushNamespace("some-namespace"),
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "some-namespace", newObj.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app to default namespace": {
			appName:  "some-app",
			srcImage: "some-image",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "default", newObj.Namespace)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"pushes app with buildpack": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
				kf.WithPushBuildpack("some-buildpack"),
			},
		},
		"pushes app with proper Service config": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: kf.PushOptions{
				kf.WithPushNamespace("myns"),
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						ka := apps.NewFromApp(newObj)

						testutil.AssertEqual(t, "service.Name", "some-app", newObj.Name)
						testutil.AssertEqual(t, "service.Kind", "App", newObj.Kind)
						testutil.AssertEqual(t, "service.APIVersion", "kf.dev/v1alpha1", newObj.APIVersion)
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", ka.GetServiceAccount())
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"properly configures build": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: kf.PushOptions{
				kf.WithPushNamespace("default"),
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
				kf.WithPushBuildpack("some-buildpack"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						testutil.AssertEqual(t, "namespace", "default", newObj.Namespace)
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", newObj.Spec.Template.Spec.ServiceAccountName)
						testutil.AssertEqual(t, "image", "some-image", newObj.Spec.Source.BuildpackBuild.Source)
						testutil.AssertEqual(t, "buildpack", "some-buildpack", newObj.Spec.Source.BuildpackBuild.Buildpack)

					}).Return(&v1alpha1.App{}, nil)

				// appsClient.EXPECT().
				// 	Tail(gomock.Any(), gomock.Any()).
				// 	Return(nil)

				// appsClient.EXPECT().
				// 	Status(gomock.Any(), gomock.Any()).
				// 	Return(true, nil)
			},
		},
		"pushes app with environment variables": {
			appName:   "some-app",
			srcImage:  "some-image",
			buildpack: "some-buildpack",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushEnvironmentVariables(map[string]string{"ENV1": "val1", "ENV2": "val2"}),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						actual := envutil.GetAppEnvVars(newObj)
						envutil.SortEnvVars(actual)
						testutil.AssertEqual(t, "envs",
							[]corev1.EnvVar{{Name: "ENV1", Value: "val1"}, {Name: "ENV2", Value: "val2"}},
							actual,
						)
					}).
					Return(&v1alpha1.App{}, nil)
			},
		},
		"deployer returns an error": {
			appName:  "some-app",
			srcImage: "some-image",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create app: some-error"), err)
			},
		},
		"set ports to h2c for gRPC": {
			appName:  "some-app",
			srcImage: "some-image",
			setup: func(t *testing.T, appsClient *appsfake.FakeClient) {
				appsClient.EXPECT().
					Upsert(gomock.Not(gomock.Nil()), gomock.Any(), gomock.Any()).
					Do(func(namespace string, newObj *v1alpha1.App, merge apps.Merger) {
						ka := apps.NewFromApp(newObj)

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
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushGrpc(true),
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

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.EXPECT().
				DeployLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeApps := appsfake.NewFakeClient(ctrl)

			tc.setup(t, fakeApps)

			p := kf.NewPusher(
				fakeLogs,
				fakeApps,
			)

			gotErr := p.Push(tc.appName, tc.srcImage, tc.opts...)
			tc.assert(t, gotErr)
			if gotErr != nil {
				return
			}

			ctrl.Finish()
		})
	}
}
