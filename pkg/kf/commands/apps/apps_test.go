package apps

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/golang/mock/gomock"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
)

func TestAppsCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace string
		wantErr   error
		listErr   error
		apps      []string
	}{
		"configured namespace": {
			namespace: "somenamespace",
		},
		"formats multiple services": {
			apps: []string{"service-a", "service-b"},
		},
		"list applications error, returns error": {
			listErr: errors.New("some-error"),
			wantErr: errors.New("some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeLister := fake.NewFakeLister(ctrl)

			fakeRecorder := fakeLister.
				EXPECT().
				List(gomock.Any()).
				DoAndReturn(func(opts ...kf.ListOption) ([]serving.Service, error) {
					t.Helper()
					if namespace := kf.ListOptions(opts).Namespace(); namespace != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, namespace)
					}

					var apps []serving.Service
					for _, a := range tc.apps {
						s := serving.Service{}
						s.Name = a
						apps = append(apps, s)
					}
					return apps, tc.listErr
				})

			buffer := &bytes.Buffer{}

			c := NewAppsCommand(&config.KfParams{
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

			header := fmt.Sprintf("Getting apps in namespace: %s", tc.namespace)
			if strings.Index(buffer.String(), header) < 0 {
				t.Fatalf("wanted header: %s: got:\n%v", header, buffer.String())
			}

			for _, app := range tc.apps {
				if strings.Index(buffer.String(), app) < 0 {
					t.Fatalf("wanted output: %s: got:\n%v", app, buffer.String())
				}
			}
		})
	}
}
