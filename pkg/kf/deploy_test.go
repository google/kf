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
	appsfake "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/internal/envutil"
	"github.com/google/kf/pkg/kf/testutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploy(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Opts    kf.DeployOptions
		Service serving.Service
		Setup   func(t *testing.T, fakeAppsClient *appsfake.FakeClient)
		Assert  func(t *testing.T, s *serving.Service, err error)
	}{
		"AppsClient returns an error": {
			Service: serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app",
				},
			},
			Setup: func(t *testing.T, fakeAppsClient *appsfake.FakeClient) {
				fakeAppsClient.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			Assert: func(t *testing.T, s *serving.Service, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create service: some-error"), err)
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
			Setup: func(t *testing.T, fakeAppsClient *appsfake.FakeClient) {
				fakeAppsClient.EXPECT().Upsert("custom-namespace", gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"Replace existing environment variables": {
			Service: buildServiceWithEnvs("some-app", map[string]string{
				"ENV1": "val1",
				"ENV2": "val2",
			}),
			Setup: func(t *testing.T, fakeAppsClient *appsfake.FakeClient) {
				old := buildServiceWithEnvs("some-app", map[string]string{
					"ENV1": "old1",
					"ENV2": "val2",
				})

				fakeAppsClient.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ns string, newObj *serving.Service, merger apps.Merger) {
					merged := merger(newObj, &old)

					envs := envutil.GetServiceEnvVars(merged)
					envutil.SortEnvVars(envs)

					testutil.AssertEqual(t, "envs", []corev1.EnvVar{
						{Name: "ENV1", Value: "val1"},
						{Name: "ENV2", Value: "val2"},
					}, envs)

				})
			},
		},
		"Leave existing environment variables": {
			Service: buildServiceWithEnvs("some-app", map[string]string{
				"ENV1": "val1",
				"ENV2": "val2",
			}),
			Setup: func(t *testing.T, fakeAppsClient *appsfake.FakeClient) {
				old := buildServiceWithEnvs("some-app", map[string]string{
					"ENV3": "val3",
					"ENV4": "val4",
				})

				fakeAppsClient.EXPECT().Upsert(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ns string, newObj *serving.Service, merger apps.Merger) {
					merged := merger(newObj, &old)

					envs := envutil.GetServiceEnvVars(merged)
					envutil.SortEnvVars(envs)

					testutil.AssertEqual(t, "envs", []corev1.EnvVar{
						{Name: "ENV1", Value: "val1"},
						{Name: "ENV2", Value: "val2"},
						{Name: "ENV3", Value: "val3"},
						{Name: "ENV4", Value: "val4"},
					}, envs)
				})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Setup == nil {
				tc.Setup = func(t *testing.T, fakeAppsClient *appsfake.FakeClient) {}
			}
			if tc.Assert == nil {
				tc.Assert = func(t *testing.T, s *serving.Service, err error) {}
			}

			ctrl := gomock.NewController(t)
			fakeAppsClient := appsfake.NewFakeClient(ctrl)

			tc.Setup(t, fakeAppsClient)

			service, gotErr := kf.NewDeployer(fakeAppsClient).Deploy(tc.Service, tc.Opts...)
			tc.Assert(t, service, gotErr)
			if gotErr != nil {
				ctrl.Finish()
			}
		})
	}
}

func buildServiceWithEnvs(appName string, envs map[string]string) serving.Service {
	app := apps.NewKfApp()
	app.SetName(appName)
	app.SetEnvVars(envutil.MapToEnvVars(envs))
	return *app.ToService()
}
