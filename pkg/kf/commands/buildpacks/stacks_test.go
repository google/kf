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

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	cbuildpacks "github.com/google/kf/v2/pkg/kf/commands/buildpacks"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	injection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestStacks(t *testing.T) {
	t.Parallel()

	enableAppDevExBuilds := make(kfconfig.FeatureFlagToggles)
	enableAppDevExBuilds.SetAppDevExperienceBuilds(true)

	for tn, tc := range map[string]struct {
		Space           string
		TargetSpace     *v1alpha1.Space
		ExpectedErr     error
		Args            []string
		EnableADXBuilds bool
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"no space chosen": {
			ExpectedErr: errors.New(config.EmptySpaceError),
			Args:        []string{},
		},
		"lists each stack": {
			Space: "some-space",
			TargetSpace: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					BuildConfig: v1alpha1.SpaceStatusBuildConfig{
						StacksV2: kfconfig.StackV2List{
							{
								Name:        "cflinuxfs3",
								Description: "A CF Compatible Stack",
								Image:       "cloudfoundry/cflinuxfs3:latest",
							},
						},
						StacksV3: kfconfig.StackV3List{
							{
								Name:        "google-golang",
								Description: "A stack for golang by Google",
								BuildImage:  "gcr.io/buildpacks/go",
								RunImage:    "gcr.io/buildpacks/slim",
							},
						},
					},
				},
			},
		},
		"lists only V3 stack when appDevExperience builds are enabled": {
			Space:           "some-space",
			EnableADXBuilds: true,
			TargetSpace: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					BuildConfig: v1alpha1.SpaceStatusBuildConfig{
						StacksV2: kfconfig.StackV2List{
							{
								Name:        "cflinuxfs3",
								Description: "A CF Compatible Stack",
								Image:       "cloudfoundry/cflinuxfs3:latest",
							},
						},
						StacksV3: kfconfig.StackV3List{
							{
								Name:        "google-golang",
								Description: "A stack for golang by Google",
								BuildImage:  "gcr.io/buildpacks/go",
								RunImage:    "gcr.io/buildpacks/slim",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			params := &config.KfParams{
				Space:       tc.Space,
				TargetSpace: tc.TargetSpace,
			}

			ctx := injection.WithInjection(context.Background(), t)
			if tc.EnableADXBuilds {
				ff := make(kfconfig.FeatureFlagToggles)
				ff.SetAppDevExperienceBuilds(true)
				ctx = testutil.WithFeatureFlags(ctx, t, ff)
			}

			var buffer bytes.Buffer
			cmd := cbuildpacks.NewStacksCommand(params)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buffer)
			cmd.SetContext(ctx)

			gotErr := cmd.Execute()
			testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)

			testutil.AssertGolden(t, "output", buffer.Bytes())
		})
	}
}
