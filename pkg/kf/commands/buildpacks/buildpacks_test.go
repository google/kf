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

package buildpacks_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/buildpacks"
	"github.com/google/kf/v2/pkg/kf/buildpacks/fake"
	cbuildpacks "github.com/google/kf/v2/pkg/kf/commands/buildpacks"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestBuildpacks(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Space           string
		ExpectedErr     error
		EnableADXBuilds bool
		Args            []string
		Setup           func(t *testing.T, fake *fake.FakeClient, params *config.KfParams)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"no space chosen": {
			ExpectedErr: errors.New(config.EmptySpaceError),
			Args:        []string{},
		},
		"listing fails": {
			Space: "my-space",
			Setup: func(t *testing.T, fake *fake.FakeClient, params *config.KfParams) {
				params.TargetSpace.Status.BuildConfig.StacksV3 = kfconfig.StackV3List{
					{
						Name:       "custom-stack",
						BuildImage: "my-image",
					},
				}
				fake.EXPECT().List("my-image").Return(nil, errors.New("some-error"))
			},
		},
		"lists each buildpack": {
			Space: "my-space",
			Setup: func(t *testing.T, fake *fake.FakeClient, params *config.KfParams) {
				params.TargetSpace.Status.BuildConfig.StacksV3 = kfconfig.StackV3List{
					{
						Name:       "custom-stack",
						BuildImage: "my-image",
					},
				}

				params.TargetSpace.Status.BuildConfig.BuildpacksV2 = kfconfig.BuildpackV2List{
					{
						Name: "java-buildpack",
						URL:  "https://path/to/java-buildpack",
					},
					{
						Name: "go-buildpack",
						URL:  "https://path/to/go-buildpack",
					},
				}

				fake.EXPECT().List("my-image").Return([]buildpacks.Buildpack{{ID: "bp-1"}, {ID: "bp-2"}}, nil)
			},
		},
		"lists each buildpack with appDevExperienceBuilds enabled": {
			Space:           "my-space",
			EnableADXBuilds: true,
			Setup: func(t *testing.T, fake *fake.FakeClient, params *config.KfParams) {
				params.TargetSpace.Status.BuildConfig.StacksV3 = kfconfig.StackV3List{
					{
						Name:       "custom-stack",
						BuildImage: "my-image",
					},
				}

				params.TargetSpace.Status.BuildConfig.BuildpacksV2 = kfconfig.BuildpackV2List{
					{
						Name: "java-buildpack",
						URL:  "https://path/to/java-buildpack",
					},
					{
						Name: "go-buildpack",
						URL:  "https://path/to/go-buildpack",
					},
				}

				fake.EXPECT().List("my-image").Return([]buildpacks.Buildpack{{ID: "bp-1"}, {ID: "bp-2"}}, nil)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeClient(ctrl)

			params := &config.KfParams{
				Space:       tc.Space,
				TargetSpace: &v1alpha1.Space{},
			}

			ctx := injection.WithInjection(context.Background(), t)
			if tc.EnableADXBuilds {
				ff := make(kfconfig.FeatureFlagToggles)
				ff.SetAppDevExperienceBuilds(true)
				ctx = testutil.WithFeatureFlags(ctx, t, ff)
			}

			if tc.Setup != nil {
				tc.Setup(t, fake, params)
			}

			var buffer bytes.Buffer
			cmd := cbuildpacks.NewBuildpacksCommand(
				params,
				fake,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)
			cmd.SetContext(ctx)

			gotErr := cmd.Execute()
			testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)

			testutil.AssertGolden(t, "output", buffer.Bytes())
		})
	}
}
