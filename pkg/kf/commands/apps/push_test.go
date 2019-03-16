package apps

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
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
		dockerImage       string
		path              string
		serviceAccount    string
		wantErr           error
		pusherErr         error
	}{
		"uses configured properties": {
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			dockerImage:       "some-docker-image",
			serviceAccount:    "some-service-account",
			path:              "some-path",
		},
		"service create error": {
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			pusherErr:         errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			path:              "some-path",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakePusher := fake.NewFakePusher(ctrl)

			fakePusher.
				EXPECT().
				Push(gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName string, opts ...kf.PushOption) error {
					if appName != tc.args[0] {
						t.Fatalf("expected appName %s, got %s", tc.args[0], appName)
					}

					if ns := kf.PushOptions(opts).Namespace(); ns != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, ns)
					}

					if p := kf.PushOptions(opts).Path(); filepath.Base(p) != tc.path {
						t.Fatalf("expected path %s, got %s", filepath.Base(tc.path), p)
					}
					if p := kf.PushOptions(opts).Path(); !filepath.IsAbs(p) {
						t.Fatalf("expected path to be an absolute: %s", p)
					}

					if cr := kf.PushOptions(opts).ContainerRegistry(); cr != tc.containerRegistry {
						t.Fatalf("expected container registry %s, got %s", tc.containerRegistry, cr)
					}

					if cr := kf.PushOptions(opts).DockerImage(); cr != tc.dockerImage {
						t.Fatalf("expected docker image %s, got %s", tc.dockerImage, cr)
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
			c.Flags().Set("docker-image", tc.dockerImage)
			c.Flags().Set("service-account", tc.serviceAccount)
			c.Flags().Set("path", tc.path)
			gotErr := c.RunE(c, tc.args)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			ctrl.Finish()
		})
	}
}
