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
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake"
	cbuildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
)

func TestUploadBuildpacks(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Setup       func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient)
	}{
		"uses correct container registry and path": {
			Setup: func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient) {
				c.Flags().Set("path", "/some-path")
				c.Flags().Set("container-registry", "some-registry.io")
				fakeClient.EXPECT().Create("/some-path", "some-registry.io").Return("some-image", nil)
				fakeClient.EXPECT().UploadBuildTemplate("some-image")
			},
		},
		"converts relative path to absolute": {
			Setup: func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient) {
				c.Flags().Set("path", "some-path")
				fakeClient.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(path, containerRegistry string) {
					if !filepath.IsAbs(path) {
						t.Fatalf("expetec path to be absolute: %s", path)
					}
				})
				fakeClient.EXPECT().UploadBuildTemplate(gomock.Any())
			},
		},
		"converts empty path to current directory": {
			Setup: func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient) {
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				fakeClient.EXPECT().Create(cwd, gomock.Any())
				fakeClient.EXPECT().UploadBuildTemplate(gomock.Any())
			},
		},
		"returns error when upload fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient) {
				c.Flags().Set("path", "/some-path")
				c.Flags().Set("container-registry", "some-registry.io")
				fakeClient.EXPECT().Create("/some-path", "some-registry.io")
				fakeClient.EXPECT().UploadBuildTemplate(gomock.Any()).Return(errors.New("some-error"))
			},
		},
		"returns error when create fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, c *cobra.Command, fakeClient *fake.FakeClient) {
				fakeClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("some-error"))
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeClient := fake.NewFakeClient(ctrl)

			c := cbuildpacks.NewUploadBuildpacks(
				&config.KfParams{
					Namespace: tc.Namespace,
					Output:    &bytes.Buffer{},
				},
				fakeClient,
			)

			if tc.Setup != nil {
				tc.Setup(t, c, fakeClient)
			}

			gotErr := c.Execute()
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			ctrl.Finish()
		})
	}
}
