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

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	kffake "github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/envutil"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/golang/mock/gomock"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeploy(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Opts    kf.DeployOptions
		Service serving.Service
		Setup   func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector)
		Assert  func(t *testing.T, s serving.Service, err error)
	}{
		"AppLister returns an error": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				fakeLister.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, s serving.Service, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to list apps: some-error"), err)
			},
		},
		"SystemEnvInjectorInterface returns an error": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				fakeLister.EXPECT().List(gomock.Any()).Return(nil, nil)
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any()).Return(errors.New("some-error"))
			},
			Assert: func(t *testing.T, s serving.Service, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to inject system environment variables: some-error"), err)
			},
		},
		"Use service name in AppLister": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				fakeLister.EXPECT().List(gomock.Any()).Do(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "app name", "some-app", kf.ListOptions(opts).AppName())
				})
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return false, nil, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any())
			},
		},
		"Creates service when app isn't present": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				expectedService := &serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-app",
					},
				}

				fakeLister.EXPECT().List(gomock.Any()).Do(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "namespace", "default", kf.ListOptions(opts).Namespace())
				})
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
					testutil.AssertEqual(t, "verb", "create", action.GetVerb())
					testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
					testutil.AssertEqual(t, "service", expectedService, service)
					return true, service, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(expectedService)
			},
			Assert: func(t *testing.T, s serving.Service, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "service", serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-app",
					},
				}, s)
			},
		},
		"Updates service when app is present": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				expectedService := &serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "some-app",
						ResourceVersion: "some-version",
					},
				}
				// Initialize structs to allow proper comparison.
				envutil.SetServiceEnvVars(expectedService, nil)

				fakeLister.EXPECT().List(gomock.Any()).Return([]serving.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "some-app",
							ResourceVersion: "some-version",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "some-other-app",
						},
					},
				}, nil)
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
					testutil.AssertEqual(t, "verb", "update", action.GetVerb())
					testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
					envutil.SetServiceEnvVars(expectedService, nil)
					testutil.AssertEqual(t, "service", expectedService, service)
					return true, service, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(expectedService)
			},
			Assert: func(t *testing.T, s serving.Service, err error) {
				testutil.AssertNil(t, "err", err)
				expectedService := serving.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "some-app",
						ResourceVersion: "some-version",
					},
				}
				envutil.SetServiceEnvVars(&expectedService, nil)
				testutil.AssertEqual(t, "service", expectedService, s)
			},
		},
		"custom namespace": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Opts: kf.DeployOptions{
				kf.WithDeployNamespace("custom-namespace"),
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				fakeLister.EXPECT().List(gomock.Any()).Do(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "namespace", "custom-namespace", kf.ListOptions(opts).Namespace())
				})
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", "custom-namespace", action.GetNamespace())
					return false, nil, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any())
			},
		},
		"Creating service fails": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				fakeLister.EXPECT().List(gomock.Any()).AnyTimes()
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("some-error")
				}))
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any())
			},
			Assert: func(t *testing.T, s serving.Service, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create service: some-error"), err)
			},
		},
		"Replace existing environment variables": {
			Service: buildServiceWithEnvs("some-app", map[string]string{
				"ENV1": "val1",
				"ENV2": "val2",
			}),
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				t.Skip()
				fakeLister.EXPECT().List(gomock.Any()).Return([]serving.Service{
					buildServiceWithEnvs("some-app", map[string]string{
						"ENV1": "old1",
						"ENV2": "val2",
					}),
				}, nil)
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
					envs := envutil.GetServiceEnvVars(service)
					envutil.SortEnvVars(envs)

					testutil.AssertEqual(t, "envs", []corev1.EnvVar{
						{Name: "ENV1", Value: "va1"},
						{Name: "ENV2", Value: "val2"},
					}, envs)
					return false, nil, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any())
			},
		},
		"Leave existing environment variables": {
			Service: buildServiceWithEnvs("some-app", map[string]string{
				"ENV1": "val1",
				"ENV2": "val2",
			}),
			Setup: func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				t.Skip()
				fakeLister.EXPECT().List(gomock.Any()).Return([]serving.Service{
					buildServiceWithEnvs("some-app", map[string]string{
						"ENV3": "val3",
						"ENV4": "val4",
					}),
				}, nil)
				cfake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
					envs := envutil.GetServiceEnvVars(service)
					envutil.SortEnvVars(envs)

					testutil.AssertEqual(t, "envs", []corev1.EnvVar{
						{Name: "ENV1", Value: "va1"},
						{Name: "ENV2", Value: "val2"},
						{Name: "ENV3", Value: "va3"},
						{Name: "ENV4", Value: "val4"},
					}, envs)
					return false, nil, nil
				}))
				fakeInjector.EXPECT().InjectSystemEnv(gomock.Any())
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Setup == nil {
				tc.Setup = func(t *testing.T, fakeLister *kffake.FakeLister, cfake *fake.FakeServingV1alpha1, fakeInjector *kffake.FakeSystemEnvInjector) {
				}
			}
			if tc.Assert == nil {
				tc.Assert = func(t *testing.T, s serving.Service, err error) {}
			}

			ctrl := gomock.NewController(t)
			fakeLister := kffake.NewFakeLister(ctrl)
			fakeInjector := kffake.NewFakeSystemEnvInjector(ctrl)

			fakec := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			tc.Setup(t, fakeLister, fakec, fakeInjector)

			service, gotErr := kf.NewDeployer(fakeLister, fakec, fakeInjector).Deploy(tc.Service, tc.Opts...)
			tc.Assert(t, service, gotErr)
			if gotErr != nil {
				ctrl.Finish()
			}
		})
	}
}

func buildServiceWithEnvs(appName string, envs map[string]string) serving.Service {
	s := serving.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
	}
	envutil.SetServiceEnvVars(&s, envutil.MapToEnvVars(envs))
	return s
}
