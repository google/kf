package buildpacks_test

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/buildpack/pack"
)

func TestBuilderCreator(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Dir               string
		ContainerRegistry string
		ExpectedErr       error
		Creator           buildpacks.BuilderFactoryCreate
	}{
		"empty dir": {
			Dir:               "",
			ContainerRegistry: "some-reg.io",
			ExpectedErr:       errors.New("dir must not be empty"),
		},
		"empty container registry": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "",
			ExpectedErr:       errors.New("containerRegistry must not be empty"),
		},
		"returns an error if creating fails": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "some-registry.io",
			Creator:           func(f pack.CreateBuilderFlags) error { return errors.New("some-error") },
			ExpectedErr:       errors.New("some-error"),
		},
		"sets the flags up": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "some-registry.io",
			Creator: func(f pack.CreateBuilderFlags) error {
				testutil.AssertEqual(t, "publish", true, f.Publish)
				testutil.AssertEqual(t, "BuilderTomlPath", "some-path/builder.toml", f.BuilderTomlPath)
				testutil.AssertEqual(t, "RepoName", "some-path/builder.toml", f.BuilderTomlPath)
				testutil.AssertRegexp(t, "RepoName", `some-registry.io/buildpack-builder:[0-9]+`, f.RepoName)
				return nil
			},
		},
		"appends builder.toml if necessary": {
			Dir:               "some-path",
			ContainerRegistry: "some-registry.io",
			Creator: func(f pack.CreateBuilderFlags) error {
				testutil.AssertEqual(t, "BuilderTomlPath", "some-path/builder.toml", f.BuilderTomlPath)
				return nil
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Creator == nil {
				tc.Creator = func(f pack.CreateBuilderFlags) error { return nil }
			}

			b := buildpacks.NewBuilderCreator(tc.Creator)
			image, gotErr := b.Create(tc.Dir, tc.ContainerRegistry)
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			testutil.AssertRegexp(t, "image name", tc.ContainerRegistry+`/buildpack-builder:[0-9]+`, image)
		})
	}
}
