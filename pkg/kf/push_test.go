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
	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/builds"
	buildsfake "github.com/google/kf/pkg/kf/builds/fake"
	kffake "github.com/google/kf/pkg/kf/fake"
	"github.com/google/kf/pkg/kf/internal/envutil"
	kfi "github.com/google/kf/pkg/kf/internal/kf"
	"github.com/google/kf/pkg/kf/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPush_BadConfig(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName  string
		srcImage string
		wantErr  error
		opts     kf.PushOptions
	}{
		"empty app name, returns error": {
			srcImage: "some-image", wantErr: errors.New("invalid app name"),
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"empty source image, returns error": {
			appName: "some-app",
			wantErr: errors.New("invalid source image"),
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			p := kf.NewPusher(
				nil, // Deployer - Should not be used
				nil, // Logs - Should not be used
				nil, // BuildClient - Should not be used
			)

			gotErr := p.Push(tc.appName, tc.srcImage, tc.opts...)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			if !kfi.ConfigError(gotErr) {
				t.Fatal("wanted error to be a ConfigError")
			}
		})
	}
}

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

			fakeDeployer := kffake.NewFakeDeployer(ctrl)
			fakeDeployer.EXPECT().
				Deploy(gomock.Not(gomock.Nil()), gomock.Any()).
				Return(&serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: tc.appName + "-version",
					},
				}, nil)

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.EXPECT().
				DeployLogs(
					gomock.Not(gomock.Nil()), // out,
					tc.appName,               // appName
					tc.appName+"-version",    // resourceVersion
					expectedNamespace,        // namespace
				).
				Return(tc.logErr)

			fakeBuilds := buildsfake.NewFakeClient(ctrl)
			mockSuccessfulBuild(fakeBuilds)

			p := kf.NewPusher(
				fakeDeployer,
				fakeLogs,
				fakeBuilds,
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
		setup     func(t *testing.T, d *kffake.FakeDeployer, bc *buildsfake.FakeClient)
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
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Do(func(service serving.Service, opts ...kf.DeployOption) {
						testutil.AssertEqual(t, "namespace", "some-namespace", kf.DeployOptions(opts).Namespace())
					}).
					Return(&serving.Service{}, nil)

				mockSuccessfulBuild(bc)
			},
		},
		"pushes app to default namespace": {
			appName:  "some-app",
			srcImage: "some-image",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Do(func(service serving.Service, opts ...kf.DeployOption) {
						testutil.AssertEqual(t, "namespace", "default", kf.DeployOptions(opts).Namespace())
						testutil.AssertEqual(t, "service.Namespace", "default", service.Namespace)
					}).
					Return(&serving.Service{}, nil)

				mockSuccessfulBuild(bc)
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
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Do(func(service serving.Service, opts ...kf.DeployOption) {
						ka := apps.NewFromService(&service)

						testutil.AssertEqual(t, "service.Name", "some-app", service.Name)
						testutil.AssertEqual(t, "service.Kind", "Service", service.Kind)
						testutil.AssertEqual(t, "service.APIVersion", "serving.knative.dev/v1alpha1", service.APIVersion)
						testutil.AssertRegexp(t, "Spec.Container.Image", `^some-reg.io/app-myns-some-app:\d+`, ka.GetImage())
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", ka.GetServiceAccount())
					}).
					Return(&serving.Service{}, nil)

				mockSuccessfulBuild(bc)
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
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				bc.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(name string, template build.TemplateInstantiationSpec, opts ...builds.CreateOption) {
						// PopulateTemplate is used by Create to generate the build
						// we use it here to make sure all the correct values are populated.
						b := builds.PopulateTemplate(name, template, opts...)

						testutil.AssertEqual(t, "namespace", "default", b.Namespace)
						testutil.AssertEqual(t, "Spec.ServiceAccountName", "some-service-account", b.Spec.ServiceAccountName)
						testutil.AssertEqual(t, "image", "some-image", b.Spec.Source.Custom.Image)
						testutil.AssertEqual(t, "Spec.Template.Name", "buildpack", b.Spec.Template.Name)
						testutil.AssertEqual(t, "Spec.Template.Kind", build.ClusterBuildTemplateKind, b.Spec.Template.Kind)

						args := make(map[string]string)
						for _, arg := range b.Spec.Template.Arguments {
							args[arg.Name] = arg.Value
						}
						testutil.AssertRegexp(t, "image name", `^some-reg.io/app-default-some-app:[0-9]{19}$`, args["IMAGE"])
						testutil.AssertEqual(t, "buildpack", "some-buildpack", args["BUILDPACK"])

					}).Return(nil, nil)

				bc.EXPECT().
					Tail(gomock.Any(), gomock.Any()).
					Return(nil)

				bc.EXPECT().
					Status(gomock.Any(), gomock.Any()).
					Return(true, nil)

				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Return(&serving.Service{}, nil)
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
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Do(func(service serving.Service, opts ...kf.DeployOption) {
						actual := envutil.GetServiceEnvVars(&service)
						envutil.SortEnvVars(actual)
						testutil.AssertEqual(t, "envs",
							[]corev1.EnvVar{{Name: "ENV1", Value: "val1"}, {Name: "ENV2", Value: "val2"}},
							actual,
						)
					}).
					Return(&serving.Service{}, nil)
				mockSuccessfulBuild(bc)
			},
		},
		"deployer returns an error": {
			appName:  "some-app",
			srcImage: "some-image",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().Deploy(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
				mockSuccessfulBuild(bc)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to deploy: some-error"), err)
			},
		},
		"set ports to h2c for gRPC": {
			appName:  "some-app",
			srcImage: "some-image",
			setup: func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
				fakeDeployer.EXPECT().
					Deploy(gomock.Any(), gomock.Any()).
					Do(func(service serving.Service, opts ...kf.DeployOption) {
						ka := apps.NewFromService(&service)

						testutil.AssertEqual(
							t,
							"container.ports",
							[]corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}},
							ka.GetContainerPorts(),
						)
					}).
					Return(&serving.Service{}, nil)
				mockSuccessfulBuild(bc)
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
				tc.setup = func(t *testing.T, fakeDeployer *kffake.FakeDeployer, bc *buildsfake.FakeClient) {
					fakeDeployer.EXPECT().
						Deploy(gomock.Not(gomock.Nil()), gomock.Any()).
						Return(&serving.Service{}, nil)

					mockSuccessfulBuild(bc)
				}
			}

			ctrl := gomock.NewController(t)

			fakeDeployer := kffake.NewFakeDeployer(ctrl)

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.EXPECT().
				DeployLogs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeBuilds := buildsfake.NewFakeClient(ctrl)

			tc.setup(t, fakeDeployer, fakeBuilds)

			p := kf.NewPusher(
				fakeDeployer,
				fakeLogs,
				fakeBuilds,
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

// mockSuccessfulBuild mocks a successful
func mockSuccessfulBuild(bc *buildsfake.FakeClient) {
	bc.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil)

	bc.EXPECT().
		Tail(gomock.Any(), gomock.Any()).
		Return(nil)

	bc.EXPECT().
		Status(gomock.Any(), gomock.Any()).
		Return(true, nil)
}
