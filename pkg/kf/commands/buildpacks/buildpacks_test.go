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

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake"
	cbuildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
)

func TestBuildpacks(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Args        []string
		Setup       func(t *testing.T, fake *fake.FakeBuildpackLister)
		BufferF     func(t *testing.T, buffer *bytes.Buffer)
	}{
		"wrong number of args": {ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
			Args: []string{"arg-1"},
		},
		"listing failes": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, fake *fake.FakeBuildpackLister) {
				fake.EXPECT().List(gomock.Any()).Return(nil, errors.New("some-error"))
			},
		},
		"custom namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fake *fake.FakeBuildpackLister) {
				fake.EXPECT().List(gomock.Any()).Do(func(opts ...buildpacks.BuildpackListOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", buildpacks.BuildpackListOptions(opts).Namespace())
				})
			},
		},
		"lists each buildpack": {
			Setup: func(t *testing.T, fake *fake.FakeBuildpackLister) {
				fake.EXPECT().List(gomock.Any()).Return([]string{"bp-1", "bp-2"}, nil)
			},
			BufferF: func(t *testing.T, buffer *bytes.Buffer) {
				testutil.AssertContainsAll(t, buffer.String(), []string{"bp-1", "bp-2"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeBuildpackLister(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			var buffer bytes.Buffer
			cmd := cbuildpacks.NewBuildpacks(
				&config.KfParams{
					Namespace: tc.Namespace,
					Output:    &buffer,
				},
				fake,
			)
			cmd.SetArgs(tc.Args)

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
