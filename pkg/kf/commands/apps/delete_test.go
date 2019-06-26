// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apps

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/commands/config"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace string
		appName   string
		wantErr   error
		deleteErr error
	}{
		"deletes given app in namespace": {
			namespace: "some-namespace",
			appName:   "some-app",
		},
		"delete app error": {
			wantErr:   errors.New("some error"),
			deleteErr: errors.New("some error"),
			appName:   "some-app",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeDeleter := fake.NewFakeClient(ctrl)

			fakeRecorder := fakeDeleter.
				EXPECT().
				DeleteInForeground(gomock.Any(), gomock.Any()).
				DoAndReturn(func(namespace, appName string) error {
					if appName != tc.appName {
						t.Fatalf("wanted appName %s, got %s", tc.appName, appName)
					}

					if ns := namespace; ns != tc.namespace {
						t.Fatalf("expected namespace %s, got %s", tc.namespace, ns)
					}
					return tc.deleteErr
				})

			buffer := &bytes.Buffer{}
			c := NewDeleteCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeDeleter)
			c.SetOutput(buffer)

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
