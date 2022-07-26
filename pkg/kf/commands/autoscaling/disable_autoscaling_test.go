// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package autoscaling

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/ptr"
)

func TestNewDisableAutoscalingCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"disable autoscaling for App succeed": {
			Space:           "default",
			Args:            []string{"my-app"},
			ExpectedStrings: []string{"Disabling"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				app := &v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(99),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								Enabled:     true,
								MaxReplicas: ptr.Int32(3),
							},
						},
					},
				}
				fake.EXPECT().Get(gomock.Any(), "default", "my-app").Return(app, nil)
				fake.EXPECT().
					Transform(gomock.Any(), "default", "my-app", gomock.Any()).
					Do(func(_ context.Context, _, _ string, m apps.Mutator) {
						testutil.AssertNil(t, "mutator error", m(app))
						testutil.AssertNotNil(t, "app.spec.instances.autoscalingspec", app.Spec.Instances.Autoscaling)
						testutil.AssertFalse(t, "app.spec.instances.autoscalingspec.enabled", app.Spec.Instances.Autoscaling.Enabled)
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"getting app fails": {
			Space:       "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("failed to get App: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Get(gomock.Any(), "default", "my-app").Return(nil, errors.New("some-error"))
			},
		},
		"no app name": {
			Space:       "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"already disabled, displays current value": {
			Space:           "default",
			Args:            []string{"my-app"},
			ExpectedStrings: []string{"Enabled?:", "false"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(99),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								Enabled:     false,
								MaxReplicas: ptr.Int32(99),
							},
						},
					},
				}, nil)
			},
		},
		"updating app fails": {
			Space:       "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("failed to disable autoscaling for App: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(00),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								Enabled:     true,
								MaxReplicas: ptr.Int32(99),
							},
						},
					},
				}, nil)
				fake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			fake := fake.NewFakeClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
			}

			cmd := NewDisableAutoscaling(p, fake)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)

		})
	}
}
