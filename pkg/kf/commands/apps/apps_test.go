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

package apps

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestAppsCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace string
		wantErr   error
		args      []string
		setup     func(t *testing.T, fakeLister *fake.FakeClient)
		assert    func(t *testing.T, buffer *bytes.Buffer)
	}{
		"invalid number of args": {
			args:    []string{"invalid"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"returns error when missing namespace": {
			wantErr: errors.New(utils.EmptyNamespaceError),
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List("some-namespace")
			},
		},
		"configured namespace": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List("some-namespace")
			},
		},
		"formats multiple apps": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{
						{ObjectMeta: metav1.ObjectMeta{Name: "app-a"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "app-b"}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "app-b"})
			},
		},
		"shows app ready": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{
						{ObjectMeta: metav1.ObjectMeta{Name: "app-a"}, Status: happyStatus()},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "ready"})
			},
		},
		"shows app not ready": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{
						{ObjectMeta: metav1.ObjectMeta{Name: "app-a"}, Status: v1alpha1.AppStatus{}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "not ready"})
			},
		},
		"shows app stopped": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Instances.Stopped = true

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "stopped"})
			},
		},
		"shows app as deleting": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				dt := metav1.Now()
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{
						{ObjectMeta: metav1.ObjectMeta{Name: "app-a", DeletionTimestamp: &dt}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "deleting"})
			},
		},
		"shows app exact instances": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Instances.Exactly = intPtr(99)

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "99"})
			},
		},
		"shows app min and max instances": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Instances.Min = intPtr(99)
				app.Spec.Instances.Max = intPtr(101)

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "99 - 101"})
			},
		},
		"shows app min instances": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Instances.Min = intPtr(99)

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "99 - ∞"})
			},
		},
		"shows app max instances": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Instances.Max = intPtr(101)

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "app-a", "0 - 101"})
			},
		},
		"shows app urls": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Spec.Routes = append(app.Spec.Routes, v1alpha1.RouteSpecFields{Domain: "example.com"})
				app.Spec.Routes = append(app.Spec.Routes, v1alpha1.RouteSpecFields{Hostname: "somehost", Domain: "example.com"})
				app.Spec.Routes = append(app.Spec.Routes, v1alpha1.RouteSpecFields{Hostname: "somehost", Domain: "example.com", Path: "somepath"})

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in space "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, "example.com/, somehost.example.com/, somehost.example.com/somepath"})
			},
		},
		"shows cluster URL": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				app := v1alpha1.App{}
				app.Name = "app-a"
				app.Status.Address = &duckv1alpha1.Addressable{
					Addressable: duckv1beta1.Addressable{
						URL: &apis.URL{
							Host:   "app-a.some-namespace.svc.cluster.local",
							Scheme: "http",
						},
					},
				}

				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{app}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"http://app-a.some-namespace.svc.cluster.local"})
			},
		},

		"listing apps fails": {
			namespace: "some-namespace",
			wantErr:   errors.New("some-error"),
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"filters out apps without a name": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]v1alpha1.App{
						{Status: v1alpha1.AppStatus{Status: duckv1beta1.Status{Conditions: []apis.Condition{{Type: "Ready", Status: "should-not-see-this"}}}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "app-b"}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				if strings.Contains(buffer.String(), "should-not-see-this") {
					t.Fatalf("expected app to be filtered out")
				}
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeLister)
			}

			buffer := &bytes.Buffer{}

			c := NewAppsCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeLister)
			c.SetOutput(buffer)

			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}
			testutil.AssertEqual(t, "SilenceUsage", true, c.SilenceUsage)

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			ctrl.Finish()
		})
	}
}

func happyStatus() v1alpha1.AppStatus {
	s := v1alpha1.AppStatus{}
	s.InitializeConditions()
	for i := range s.Conditions {
		s.Conditions[i].Status = corev1.ConditionTrue
	}
	return s
}
