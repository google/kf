package apps

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/golang/mock/gomock"
)

func TestPushCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		args              []string
		namespace         string
		containerRegistry string
		serviceAccount    string
		wantErr           error
		wantUsageOnErr    bool
		pusherErr         error
	}{
		"uses configured namespace": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
		"container registry not configured, returns error": {
			args:           []string{"app-name"},
			wantErr:        errors.New("container registry is not set"),
			serviceAccount: "some-service-account",
			wantUsageOnErr: true,
		},
		"service account not configured, returns error": {
			args:              []string{"app-name"},
			wantErr:           errors.New("service account is not set"),
			containerRegistry: "some-reg.io",
			wantUsageOnErr:    true,
		},
		"service create error": {
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			pusherErr:         errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakePusher := fake.NewFakePusher(ctrl)

			fakeRecorder := fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName string, opts ...kf.PushOption) error {
					if appName != tc.args[0] {
						t.Fatalf("expected appName %s, got %s", tc.args[0], appName)
					}

					if ns := kf.PushOptions(opts).Namespace(); ns != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, ns)
					}

					if path := kf.PushOptions(opts).Path(); path != "" {
						t.Fatalf("expected path to be empty, got %s", path)
					}

					if cr := kf.PushOptions(opts).ContainerRegistry(); cr != tc.containerRegistry {
						t.Fatalf("expected container registry %s, got %s", tc.containerRegistry, cr)
					}

					if sa := kf.PushOptions(opts).ServiceAccount(); sa != tc.serviceAccount {
						t.Fatalf("expected service account %s, got %s", tc.serviceAccount, sa)
					}
					return tc.pusherErr
				})

			buffer := &bytes.Buffer{}

			c := NewPushCommand(&config.KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakePusher)

			c.Flags().Set("container-registry", tc.containerRegistry)
			c.Flags().Set("service-account", tc.serviceAccount)
			gotErr := c.RunE(c, tc.args)
			if tc.wantErr != nil || gotErr != nil {
				// We don't really care if Push was invoked if we want an
				// error.
				fakeRecorder.AnyTimes()

				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				if !tc.wantUsageOnErr != c.SilenceUsage {
					t.Fatalf("wanted %v, got %v", !tc.wantUsageOnErr, c.SilenceUsage)
				}

				return
			}
		})
	}
}
