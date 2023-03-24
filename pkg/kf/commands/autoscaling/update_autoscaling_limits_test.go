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

func TestUpdateAutoscalingRuleLimits(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"updated autoscaling limits for App": {
			Space:           "default",
			Args:            []string{"my-app", "1", "3"},
			ExpectedStrings: []string{"updating"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				app := &v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Stopped:  false,
							Replicas: ptr.Int32(99),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								MaxReplicas: ptr.Int32(99),
								Enabled:     true,
							},
						},
					},
				}
				fake.EXPECT().
					Transform(gomock.Any(), "default", "my-app", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, m apps.Mutator) (*v1alpha1.App, error) {
						testutil.AssertNil(t, "mutator error", m(app))
						testutil.AssertEqual(t, "app.spec.instances.replicas", int32(99), *app.Spec.Instances.Replicas)
						testutil.AssertNotNil(t, "app.spec.instances.autoscalingspec", app.Spec.Instances.Autoscaling)
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.maxreplicas", int32(3), *app.Spec.Instances.Autoscaling.MaxReplicas)
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.minreplicas", int32(1), *app.Spec.Instances.Autoscaling.MinReplicas)
						return app, nil
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"wrong number of args": {
			Space:       "default",
			Args:        []string{"20", "80"},
			ExpectedErr: errors.New("accepts 3 arg(s), received 2"),
		},
		"invalid min instances": {
			Space:       "default",
			Args:        []string{"my-ap", "min", "80"},
			ExpectedErr: errors.New("min instances has to be an integer: strconv.ParseInt: parsing \"min\": invalid syntax"),
		},
		"invalid max instances": {
			Space:       "default",
			Args:        []string{"my-ap", "20", "max"},
			ExpectedErr: errors.New("max instances has to be an integer: strconv.ParseInt: parsing \"max\": invalid syntax"),
		},
		"updating app fails": {
			Space:       "default",
			Args:        []string{"my-app", "20", "80"},
			ExpectedErr: errors.New("failed to update autoscaling limits for App: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
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

			fake.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&v1alpha1.App{}, nil)

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
			}

			cmd := NewUpdateAutoscalingLimits(p, fake)
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
