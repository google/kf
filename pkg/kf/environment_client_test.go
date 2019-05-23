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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestEnvironmentClient_List(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		appName        string
		expectedValues map[string]string
		opts           []kf.ListEnvOption
		wantErr        error
		setup          func(t *testing.T, fake *kffake.FakeLister)
	}{
		"empty app name": {
			wantErr: errors.New("invalid app name"),
		},
		"listing fails": {
			appName: "some-app",
			wantErr: errors.New("some-error"),
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"unknown app": {
			appName: "some-app",
			wantErr: errors.New("expected 1 app, but found 0"),
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any())
			},
		},
		"custom namespace": {
			appName: "some-app",
			opts: []kf.ListEnvOption{
				kf.WithListEnvNamespace("some-namespace"),
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Do(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", kf.ListOptions(opts).Namespace())
				}).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
		"requests app name": {
			appName: "some-app",
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).DoAndReturn(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "app name", "some-app", kf.ListOptions(opts).AppName())
				}).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
		"empty results": {
			appName: "some-app",
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
			wantErr: nil, // Ensure it doesn't faile on empty results
		},
		"with results": {
			appName: "some-app",
			expectedValues: map[string]string{
				"name-0": "value-0",
				"name-1": "value-1",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs(
					"some-app",
					map[string]string{
						"name-0": "value-0",
						"name-1": "value-1",
					},
				)}, nil)
			},
			wantErr: nil, // Ensure it doesn't faile on empty results
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := kffake.NewFakeLister(ctrl)

			if tc.setup != nil {
				tc.setup(t, fake)
			}

			c := kf.NewEnvironmentClient(
				fake,
				nil, // Serving Factory. Not used with List.
			)

			values, gotErr := c.List(tc.appName, tc.opts...)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			testutil.AssertEqual(t, "env value", tc.expectedValues, values)
			ctrl.Finish()
		})
	}
}

func TestEnvironmentClient_Config_Set(t *testing.T) {
	t.Parallel()

	setupEnvClientTests(t, "Set", func(c kf.EnvironmentClient, appName, namespace string, m map[string]string) error {
		var opts []kf.SetEnvOption
		if namespace != "" {
			opts = append(opts, kf.WithSetEnvNamespace(namespace))
		}
		return c.Set(appName, m, opts...)
	})
}

func TestEnvironmentClient_Set(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		values         map[string]string
		expectedValues map[string]string
		setup          func(t *testing.T, fake *kffake.FakeLister)
	}{
		"adds new envs": {
			expectedValues: map[string]string{
				"name-0": "value-0",
				"name-1": "value-1",
			},
			values: map[string]string{
				"name-0": "value-0",
				"name-1": "value-1",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
		"appends new envs": {
			expectedValues: map[string]string{
				"name-0": "value-0",
				"name-1": "value-1",
				"name-2": "value-2",
			},
			values: map[string]string{
				"name-1": "value-1",
				"name-2": "value-2",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", map[string]string{"name-0": "value-0"})}, nil)
			},
		},
		"clobbers old envs": {
			expectedValues: map[string]string{
				"name-0": "new",
				"name-1": "value-1",
			},
			values: map[string]string{
				"name-0": "new",
				"name-1": "value-1",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", map[string]string{"name-0": "value-0"})}, nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := kffake.NewFakeLister(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeLister)
			}

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				actualEnvs := action.(ktesting.CreateAction).GetObject().(*serving.Service).Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env

				// Converting to map will clobber. Make sure we don't have duplicates
				testutil.AssertEqual(t, "service.env len", len(tc.expectedValues), len(actualEnvs))
				testutil.AssertEqual(t, "service.env", tc.expectedValues, envutil.EnvVarsToMap(actualEnvs))

				return false, nil, nil
			}))

			c := kf.NewEnvironmentClient(
				fakeLister,
				fakeServing,
			)

			gotErr := c.Set("some-app", tc.values)
			testutil.AssertErrorsEqual(t, nil, gotErr)
			ctrl.Finish()
		})
	}
}

func TestEnvironmentClient_Config_Unset(t *testing.T) {
	t.Parallel()

	setupEnvClientTests(t, "Unset", func(c kf.EnvironmentClient, appName, namespace string, m map[string]string) error {
		var (
			opts  []kf.UnsetEnvOption
			names []string
		)
		for name := range m {
			names = append(names, name)
		}

		if namespace != "" {
			opts = append(opts, kf.WithUnsetEnvNamespace(namespace))
		}

		return c.Unset(appName, names, opts...)
	})
}

func TestEnvironmentClient_Unset(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		names          []string
		expectedValues map[string]string
		setup          func(t *testing.T, fake *kffake.FakeLister)
	}{
		"remove all envs": {
			names: []string{
				"name-0",
				"name-1",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", map[string]string{"name-0": "value-0", "name-1": "value-1"})}, nil)
			},
		},
		"remove some envs": {
			expectedValues: map[string]string{
				"name-1": "value-1",
			},
			names: []string{
				"name-0",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", map[string]string{"name-0": "value-0", "name-1": "value-1"})}, nil)
			},
		},
		"remove non-existing env": {
			expectedValues: map[string]string{
				"name-0": "value-0",
			},
			names: []string{
				"not-there",
				"name-1",
			},
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", map[string]string{"name-0": "value-0", "name-1": "value-1"})}, nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := kffake.NewFakeLister(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeLister)
			}

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				actualEnvs := action.(ktesting.CreateAction).GetObject().(*serving.Service).Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env

				// Converting to map will clobber. Make sure we don't have duplicates
				testutil.AssertEqual(t, "service.env len", len(tc.expectedValues), len(actualEnvs))
				testutil.AssertEqual(t, "service.env", tc.expectedValues, envutil.EnvVarsToMap(actualEnvs))

				return false, nil, nil
			}))

			c := kf.NewEnvironmentClient(
				fakeLister,
				fakeServing,
			)

			gotErr := c.Unset("some-app", tc.names)
			testutil.AssertErrorsEqual(t, nil, gotErr)
			ctrl.Finish()
		})
	}
}

func setupEnvClientTests(t *testing.T, prefix string, f func(c kf.EnvironmentClient, appName, namespace string, m map[string]string) error) {
	for tn, tc := range map[string]struct {
		appName           string
		expectedNamespace string
		values            map[string]string
		expectedValues    map[string]string
		updateErr         error
		wantErr           error
		setup             func(t *testing.T, fake *kffake.FakeLister)
	}{
		"empty app name": {
			wantErr: errors.New("invalid app name"),
		},
		"custom namespace": {
			appName:           "some-app",
			expectedNamespace: "some-namespace",
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Do(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", kf.ListOptions(opts).Namespace())
				}).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
		"requests app name": {
			appName:           "some-app",
			expectedNamespace: "default",
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).DoAndReturn(func(opts ...kf.ListOption) {
					testutil.AssertEqual(t, "app name", "some-app", kf.ListOptions(opts).AppName())
				}).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
		"listing fails": {
			appName:           "some-app",
			expectedNamespace: "default",
			wantErr:           errors.New("some-error"),
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"unknown app": {
			appName:           "some-app",
			expectedNamespace: "default",
			wantErr:           errors.New("expected 1 app, but found 0"),
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any())
			},
		},
		"updating fails": {
			appName:           "some-app",
			expectedNamespace: "default",
			wantErr:           errors.New("some-error"),
			updateErr:         errors.New("some-error"),
			setup: func(t *testing.T, fake *kffake.FakeLister) {
				fake.EXPECT().List(gomock.Any()).Return([]serving.Service{buildServiceWithEnvs("some-app", nil)}, nil)
			},
		},
	} {
		t.Run(prefix+": "+tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := kffake.NewFakeLister(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeLister)
			}

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "action.namespace", tc.expectedNamespace, action.GetNamespace())
				testutil.AssertEqual(t, "action.verb", "update", action.GetVerb())
				testutil.AssertEqual(t, "action.resources", "services", action.GetResource().Resource)

				return tc.updateErr != nil, nil, tc.updateErr
			}))

			c := kf.NewEnvironmentClient(
				fakeLister,
				fakeServing,
			)

			gotErr := f(c, tc.appName, tc.expectedNamespace, tc.values)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}
			ctrl.Finish()
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
