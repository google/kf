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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewScaleCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Namespace       string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeClient)
	}{
		"updates app to exact instances": {
			Namespace:       "default",
			Args:            []string{"my-app", "-i=3"},
			ExpectedStrings: []string{"Stopped?:", "true", "Exactly:", "3"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform("default", "my-app", gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						someInt := 9
						app := v1alpha1.App{}
						app.Spec.Instances.Min = &someInt
						app.Spec.Instances.Max = &someInt
						app.Spec.Instances.Stopped = true
						testutil.AssertNil(t, "mutator error", m(&app))
						testutil.AssertEqual(t, "app.spec.instances.exactly", 3, *app.Spec.Instances.Exactly)

						// Assert stopped wasn't altered
						testutil.AssertEqual(t, "app.spec.instances.stopped", true, app.Spec.Instances.Stopped)
						testutil.AssertEqual(t, "app.spec.instances.min", true, app.Spec.Instances.Min == nil)
						testutil.AssertEqual(t, "app.spec.axstances.max", true, app.Spec.Instances.Max == nil)
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"updates app auto scaling": {
			Namespace:       "default",
			Args:            []string{"my-app", "--min=3", "--max=5"},
			ExpectedStrings: []string{"Stopped?:", "true", "Min:", "3", "Min:", "5"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform("default", "my-app", gomock.Any()).
					Do(func(_, _ string, m apps.Mutator) {
						min, max := 3, 5
						app := v1alpha1.App{}
						app.Spec.Instances.Min = &min
						app.Spec.Instances.Max = &max
						app.Spec.Instances.Stopped = true
						testutil.AssertNil(t, "mutator error", m(&app))

						testutil.AssertEqual(t, "app.spec.instances.exactly", true, app.Spec.Instances.Exactly == nil)
						testutil.AssertEqual(t, "app.spec.instances.min", 3, *app.Spec.Instances.Min)
						testutil.AssertEqual(t, "app.spec.axstances.max", 5, *app.Spec.Instances.Max)

						// Assert stopped wasn't altered
						testutil.AssertEqual(t, "app.spec.instances.stopped", true, app.Spec.Instances.Stopped)
					})
				fake.EXPECT().WaitForConditionKnativeServiceReadyTrue(gomock.Any(), "default", "my-app", gomock.Any())
			},
		},
		"async does not wait": {
			Namespace: "default",
			Args:      []string{"my-app", "--min=3", "--max=5", "--async"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		"no app name": {
			Namespace:   "default",
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"flags not set, displays current value": {
			Namespace:       "default",
			Args:            []string{"my-app"},
			ExpectedStrings: []string{"Stopped?:", "false", "Exactly:", "99"},
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				exactly := 99
				fake.EXPECT().Get("default", "my-app").Return(&v1alpha1.App{
					Spec: v1alpha1.AppSpec{
						Instances: v1alpha1.AppSpecInstances{
							Exactly: &exactly,
						},
					},
				}, nil)
			},
		},
		"getting app fails": {
			Namespace:   "default",
			Args:        []string{"my-app"},
			ExpectedErr: errors.New("failed to get app: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().Get("default", "my-app").Return(nil, errors.New("some-error"))
			},
		},
		"autoscale and exact flags set": {
			Namespace:   "default",
			Args:        []string{"my-app", "--instances=3", "--max=9"},
			ExpectedErr: errors.New("failed to scale app: expected exactly one, got both: exactly, max"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform("default", "my-app", gomock.Any()).
					DoAndReturn(func(_, _ string, m apps.Mutator) (*v1alpha1.App, error) {
						return nil, m(&v1alpha1.App{})
					})
			},
		},
		"min greater than max": {
			Namespace:   "default",
			Args:        []string{"my-app", "--min=8", "--max=5"},
			ExpectedErr: errors.New("failed to scale app: max must be >= min: max, min"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform("default", "my-app", gomock.Any()).
					DoAndReturn(func(_, _ string, m apps.Mutator) (*v1alpha1.App, error) {
						return nil, m(&v1alpha1.App{})
					})
			},
		},
		"updating app fails": {
			Namespace:   "default",
			Args:        []string{"my-app", "-i=3"},
			ExpectedErr: errors.New("failed to scale app: some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient) {
				fake.EXPECT().
					Transform(gomock.Any(), gomock.Any(), gomock.Any()).
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
				Namespace: tc.Namespace,
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

			ctrl.Finish()
		})
	}
}
