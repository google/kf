package buildpacks_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake"
	cbuildpacks "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
)

func TestUploadBuildpacks(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace   string
		ExpectedErr error
		Setup       func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader)
	}{
		"uses correct container registry and path": {
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				c.Flags().Set("path", "/some-path")
				c.Flags().Set("container-registry", "some-registry.io")
				fakeBuilderCreator.EXPECT().Create("/some-path", "some-registry.io").Return("some-image", nil)
				fakeBuildTemplateUploader.EXPECT().UploadBuildTemplate("some-image", gomock.Any())
			},
		},
		"converts relative path to absolute": {
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				c.Flags().Set("path", "some-path")
				fakeBuilderCreator.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(path, containerRegistry string) {
					if !filepath.IsAbs(path) {
						t.Fatalf("expetec path to be absolute: %s", path)
					}
				})
				fakeBuildTemplateUploader.EXPECT().UploadBuildTemplate(gomock.Any(), gomock.Any())
			},
		},
		"converts empty path to current directory": {
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				fakeBuilderCreator.EXPECT().Create(cwd, gomock.Any())
				fakeBuildTemplateUploader.EXPECT().UploadBuildTemplate(gomock.Any(), gomock.Any())
			},
		},
		"passes namespace": {
			Namespace: "some-namespace",
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				fakeBuilderCreator.EXPECT().Create(gomock.Any(), gomock.Any())
				fakeBuildTemplateUploader.EXPECT().UploadBuildTemplate(gomock.Any(), gomock.Any()).Do(func(imageName string, opts ...buildpacks.UploadBuildTemplateOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", buildpacks.UploadBuildTemplateOptions(opts).Namespace())
				})
			},
		},
		"returns error when upload fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				c.Flags().Set("path", "/some-path")
				c.Flags().Set("container-registry", "some-registry.io")
				fakeBuilderCreator.EXPECT().Create("/some-path", "some-registry.io")
				fakeBuildTemplateUploader.EXPECT().UploadBuildTemplate(gomock.Any(), gomock.Any()).Return(errors.New("some-error"))
			},
		},
		"returns error when create fails": {
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, c *cobra.Command, fakeBuilderCreator *fake.FakeBuilderCreator, fakeBuildTemplateUploader *fake.FakeBuildTemplateUploader) {
				fakeBuilderCreator.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("some-error"))
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeBuilderCreator := fake.NewFakeBuilderCreator(ctrl)
			fakeBuildTemplateUploader := fake.NewFakeBuildTemplateUploader(ctrl)

			c := cbuildpacks.NewUploadBuildpacks(
				&config.KfParams{
					Namespace: tc.Namespace,
					Output:    &bytes.Buffer{},
				},
				fakeBuilderCreator,
				fakeBuildTemplateUploader,
			)

			if tc.Setup != nil {
				tc.Setup(t, c, fakeBuilderCreator, fakeBuildTemplateUploader)
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
