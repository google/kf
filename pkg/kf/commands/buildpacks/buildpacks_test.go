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
