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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/buildpacks"
	"github.com/google/kf/pkg/kf/buildpacks/fake"
	cbuildpacks "github.com/google/kf/pkg/kf/commands/buildpacks"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestBuildpacks(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fake *fake.FakeClient, params *config.KfParams)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args:        []string{"arg-1"},
		},
		"no space chosen": {
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
			Args:        []string{},
		},
		"listing fails": {
			Namespace:   "my-space",
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, fake *fake.FakeClient, params *config.KfParams) {
				params.TargetSpace.Spec.BuildpackBuild.BuilderImage = "my-image"

				fake.EXPECT().List("my-image").Return(nil, errors.New("some-error"))
			},
		},
		"lists each buildpack": {
			Namespace: "my-space",
			Setup: func(t *testing.T, fake *fake.FakeClient, params *config.KfParams) {
				params.TargetSpace.Spec.BuildpackBuild.BuilderImage = "my-image"

				fake.EXPECT().List("my-image").Return([]buildpacks.Buildpack{{ID: "bp-1"}, {ID: "bp-2"}}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"bp-1", "bp-2"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeClient(ctrl)

			params := &config.KfParams{
				Namespace: tc.Namespace,
			}
			params.SetTargetSpaceToDefault()

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

			gotErr := cmd.Execute()
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			if tc.BufferF != nil {
				tc.BufferF(t, &buffer)
			}

			ctrl.Finish()
		})
	}
}
