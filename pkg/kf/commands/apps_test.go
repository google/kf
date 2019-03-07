package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
)

func TestAppsCommand(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		namespace string
		wantErr   error
		listErr   error
		apps      []string
	}{
		{
			name:      "configured namespace",
			namespace: "somenamespace",
		},
		{
			name: "formats multiple services",
			apps: []string{"service-a", "service-b"},
		},
		{
			name:    "list applications error, returns error",
			listErr: errors.New("some-error"),
			wantErr: errors.New("some-error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeLister := fake.NewFakeLister(ctrl)

			fakeRecorder := fakeLister.
				EXPECT().
				List(gomock.Any()).
				DoAndReturn(func(opts ...kf.ListOption) ([]kf.App, error) {
					t.Helper()
					if namespace := kf.ListOptions(opts).Namespace(); namespace != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, namespace)
					}

					var apps []kf.App
					for _, a := range tc.apps {
						apps = append(apps, kf.App{a})
					}
					return apps, tc.listErr
				})

			buffer := &bytes.Buffer{}

			c := NewAppsCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakeLister)

			gotErr := c.RunE(&cobra.Command{}, nil)
			if tc.wantErr != nil {
				// We don't really care if Push was invoked if we want an
				// error.
				fakeRecorder.AnyTimes()

				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}
				return
			}

			for _, app := range tc.apps {
				if strings.Index(buffer.String(), app) < 0 {
					t.Fatalf("wanted output: %s: got:\n%v", app, buffer.String())
				}
			}
		})
	}
}
