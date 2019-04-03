package apps

import (
	"bytes"
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
)

func TestUnsetEnvCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace       string
		Args            []string
		ExpectedStrings []string
		ExpectedErr     error
		Setup           func(t *testing.T, fake *fake.FakeEnvironmentClient)
	}{
		"wrong number of params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"unsetting variables fails": {
			Args:        []string{"app-name", "NAME"},
			ExpectedErr: errors.New("some-error"),
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Unset("app-name", gomock.Any(), gomock.Any()).Return(errors.New("some-error"))
			},
		},
		"custom namespace": {
			Args:      []string{"app-name", "NAME"},
			Namespace: "some-namespace",
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Unset(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(appName string, names []string, opts ...kf.UnsetEnvOption) {
					testutil.AssertEqual(t, "namespace", "some-namespace", kf.UnsetEnvOptions(opts).Namespace())
				})
			},
		},
		"unsets values": {
			Args: []string{"app-name", "NAME"},
			Setup: func(t *testing.T, fake *fake.FakeEnvironmentClient) {
				fake.EXPECT().Unset("app-name", []string{"NAME"}, gomock.Any())
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fake := fake.NewFakeEnvironmentClient(ctrl)

			if tc.Setup != nil {
				tc.Setup(t, fake)
			}

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Output:    buf,
				Namespace: tc.Namespace,
			}

			cmd := NewUnsetEnvCommand(p, fake)
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
