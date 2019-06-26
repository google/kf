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
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/knative/pkg/apis"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		"configured namespace": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List("some-namespace")
			},
		},
		"formats multiple services": {
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]serving.Service{
						{ObjectMeta: metav1.ObjectMeta{Name: "service-a"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "service-b"}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in namespace: "
				header2 := "Found 2 apps in namespace "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, header2, "service-a", "service-b"})
			},
		},
		"shows app as deleting": {
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				dt := metav1.Now()
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]serving.Service{
						{ObjectMeta: metav1.ObjectMeta{Name: "service-a", DeletionTimestamp: &dt}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header1 := "Getting apps in namespace: "
				header2 := "Found 1 apps in namespace "
				testutil.AssertContainsAll(t, buffer.String(), []string{header1, header2, "service-a", "Deleting"})
			},
		},
		"list applications error, returns error": {
			wantErr: errors.New("some-error"),
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"filters out configurations without a name": {
			setup: func(t *testing.T, fakeLister *fake.FakeClient) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]serving.Service{
						{Status: serving.ServiceStatus{Status: duckv1beta1.Status{Conditions: []apis.Condition{{Type: "Ready", Status: "should-not-see-this"}}}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "service-b"}},
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

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			ctrl.Finish()
		})
	}
}
