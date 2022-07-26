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

func TestCreateAutoscalingRuleCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Space           string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"Created autoscaling rule for App": {
			Space:           "default",
			Args:            []string{"my-app", "CPU", "20", "80"},
			ExpectedStrings: []string{"Creating"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				app := &v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Replicas: ptr.Int32(99),
							Autoscaling: v1alpha1.AppSpecAutoscaling{
								MaxReplicas: ptr.Int32(99),
							},
						},
					},
				}
				fake.EXPECT().
					Transform(gomock.Any(), "default", "my-app", gomock.Any()).
					Do(func(_ context.Context, _, _ string, m apps.Mutator) {
						testutil.AssertNil(t, "mutator error", m(app))
						testutil.AssertNotNil(t, "app.spec.instances.autoscalingspec", app.Spec.Instances.Autoscaling)
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.maxreplicas", int32(99), *app.Spec.Instances.Autoscaling.MaxReplicas)
						testutil.AssertNotNil(t, "app.spec.instances.autoscalingspec.rules", app.Spec.Instances.Autoscaling.Rules)
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.rules", 1, len(app.Spec.Instances.Autoscaling.Rules))
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.rules[0].ruletype", v1alpha1.CPURuleType, app.Spec.Instances.Autoscaling.Rules[0].RuleType)
						testutil.AssertEqual(t, "app.spec.instances.autoscalingspec.rules[0].ruletype", int32(50), *app.Spec.Instances.Autoscaling.Rules[0].Target)
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"wrong number of args": {
			Space:       "default",
			Args:        []string{"CPU", "20", "80"},
			ExpectedErr: errors.New("accepts 4 arg(s), received 3"),
		},
		"invalid min threshold": {
			Space:       "default",
			Args:        []string{"my-ap", "CPU", "min", "80"},
			ExpectedErr: errors.New("min threshold has to be an integer: strconv.ParseInt: parsing \"min\": invalid syntax"),
		},
		"invalid max threshold": {
			Space:       "default",
			Args:        []string{"my-ap", "CPU", "20", "max"},
			ExpectedErr: errors.New("max threshold has to be an integer: strconv.ParseInt: parsing \"max\": invalid syntax"),
		},
		"updating app failed": {
			Space:       "default",
			Args:        []string{"my-app", "CPU", "20", "80"},
			ExpectedErr: errors.New("failed to add autoscaling rule for App: some-error"),
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

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.Space,
			}

			cmd := NewCreateAutoscalingRule(p, fake)
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
