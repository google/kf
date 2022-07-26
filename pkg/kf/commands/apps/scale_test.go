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

func TestNewScaleCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"updates app to exact instances": {
			Space:           "default",
			Args:            []string{"my-app", "-i=3"},
			ExpectedStrings: []string{"Stopped?:", "true", "Replicas:", "3"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform(gomock.Any(), "default", "my-app", gomock.Any()).
					Do(func(_ context.Context, _, _ string, m apps.Mutator) {
						app := v1alpha1.App{}
						app.Spec.Instances.Stopped = true
						testutil.AssertNil(t, "mutator error", m(&app))
						testutil.AssertEqual(t, "app.spec.instances.exactly", int32(3), *app.Spec.Instances.Replicas)

						// Assert stopped wasn't altered
						testutil.AssertEqual(t, "app.spec.instances.stopped", true, app.Spec.Instances.Stopped)
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"async does not wait": {
			Space: "default",
			Args:  []string{"my-app", "--instances=3", "--async"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"no app name": {
			Space:       "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"flags not set, displays current value": {
			Space:           "default",
			Args:            []string{"my-app"},
			ExpectedStrings: []string{"Stopped?:", "false", "Replicas:", "99"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Get(gomock.Any(), "default", "my-app").Return(&v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(99),
						},
					},
				}, nil)
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
		"updating app fails": {
			Space:       "default",
			Args:        []string{"my-app", "-i=3"},
			ExpectedErr: errors.New("failed to scale App: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
		"autoscaling on, update returns error": {
			Space:       "default",
			Args:        []string{"my-app", "-i=30"},
			ExpectedErr: errors.New("failed to scale App: cannot scale App manually when autoscaling is turned on"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				app := v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(00),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								Enabled:     true,
								MaxReplicas: ptr.Int32(2),
								Rules: []v1alpha1.AppAutoscalingRule{
									{
										RuleType: v1alpha1.CPURuleType,
										Target:   ptr.Int32(80),
									},
								},
							},
						},
					},
				}
				fake.EXPECT().
					Transform(gomock.Any(), "default", "my-app", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, m apps.Mutator) (*v1alpha1.App, error) {
						return nil, m(&app)
					})
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

			cmd := NewScaleCommand(p, fake)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
			testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)

		})
	}
}
