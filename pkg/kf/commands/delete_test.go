package commands

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/golang/mock/gomock"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		namespace string
		appName   string
		wantErr   error
		deleteErr error
	}{
		{
			name:      "deletes given app in namespace",
			namespace: "some-namespace",
			appName:   "some-app",
		},
		{
			name:      "delete app error",
			wantErr:   errors.New("some error"),
			deleteErr: errors.New("some error"),
			appName:   "some-app",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeDeleter := fake.NewFakeDeleter(ctrl)

			fakeRecorder := fakeDeleter.
				EXPECT().
				Delete(gomock.Any(), gomock.Any()).
				DoAndReturn(func(appName string, opts ...kf.DeleteOption) error {
					if appName != tc.appName {
						t.Fatalf("wanted appName %s, got %s", tc.appName, appName)
					}

					if ns := kf.DeleteOptions(opts).Namespace(); ns != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, ns)
					}
					return tc.deleteErr
				})

			buffer := &bytes.Buffer{}
			c := NewDeleteCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakeDeleter)

			gotErr := c.RunE(c, []string{tc.appName})
			if tc.wantErr != nil || gotErr != nil {
				// We don't really care if Push was invoked if we want an
				// error.
				fakeRecorder.AnyTimes()

				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				if !c.SilenceUsage {
					t.Fatalf("wanted %v, got %v", true, c.SilenceUsage)
				}

				return
			}
		})
	}
}
