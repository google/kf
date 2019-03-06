package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fakes"
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
			buffer := &bytes.Buffer{}
			fakeLister := &fakes.FakeLister{
				T: t,
				Action: func(namespace string) ([]kf.App, error) {
					t.Helper()
					if namespace != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, namespace)
					}

					var apps []kf.App
					for _, a := range tc.apps {
						apps = append(apps, kf.App{a})
					}
					return apps, tc.listErr
				},
			}

			c := NewAppsCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakeLister)

			gotErr := c.RunE(&cobra.Command{}, nil)
			if tc.wantErr != nil {
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
