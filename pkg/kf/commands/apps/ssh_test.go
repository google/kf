// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apps

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	fakeexecstreamer "github.com/google/kf/v2/pkg/kf/injection/clients/execstreamer/fake"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

func TestNewSSHCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		space string
		args  []string

		wantErr        error
		wantExec       bool
		wantSpace      string
		wantSelector   metav1.ListOptions
		wantExecOpts   corev1.PodExecOptions
		wantStreamOpts remotecommand.StreamOptions
	}{
		"invalid space": {
			space:    "",
			args:     []string{"my-app"},
			wantExec: false,
			wantErr:  errors.New(config.EmptySpaceError),
		},
		"no arg": {
			space:    "my-space",
			args:     []string{},
			wantExec: false,
			wantErr:  errors.New("accepts 1 arg(s), received 0"),
		},
		"pod": {
			space:     "my-space",
			args:      []string{"pod/my-custom-pod"},
			wantExec:  true,
			wantSpace: "my-space",
			wantSelector: metav1.ListOptions{
				FieldSelector: "metadata.name=my-custom-pod",
			},
			wantExecOpts: corev1.PodExecOptions{
				Command:   []string{"/bin/bash"},
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
				Container: "user-container",
			},
			wantStreamOpts: remotecommand.StreamOptions{
				Tty: true,
			},
		},
		"app": {
			space:     "my-space",
			args:      []string{"my-app"},
			wantExec:  true,
			wantSpace: "my-space",
			wantSelector: metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/component=app-server,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=my-app",
			},
			wantExecOpts: corev1.PodExecOptions{
				Command:   []string{"/bin/bash"},
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
				Container: "user-container",
			},
			wantStreamOpts: remotecommand.StreamOptions{
				Tty: true,
			},
		},
		"fully custom": {
			space: "my-space",
			args: []string{
				"pod/my-app-pod",
				"-c", "/bin/sh",
				"-c", "myscript.sh",
				"-T",
				"--container", "istio-proxy",
			},
			wantExec:  true,
			wantSpace: "my-space",
			wantSelector: metav1.ListOptions{
				FieldSelector: "metadata.name=my-app-pod",
			},
			wantExecOpts: corev1.PodExecOptions{
				Command:   []string{"/bin/sh", "myscript.sh"},
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       false,
				Container: "istio-proxy",
			},
			wantStreamOpts: remotecommand.StreamOptions{
				Tty: false,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Space: tc.space,
			}

			ctx := fakeinjection.WithInjection(context.Background(), t)

			gotExec := false
			fakeexecstreamer.Get(ctx).StreamF = func(
				ctx context.Context,
				selector metav1.ListOptions,
				execOpts corev1.PodExecOptions,
				streamOpts remotecommand.StreamOptions,
			) error {
				gotExec = true

				testutil.AssertEqual(t, "selector", tc.wantSelector, selector)
				testutil.AssertEqual(t, "execOpts", tc.wantExecOpts, execOpts)
				// for stream opts, ignore stdin/stdout/stderr
				streamOpts.Stderr = nil
				streamOpts.Stdin = nil
				streamOpts.Stdout = nil
				testutil.AssertEqual(t, "streamOpts", tc.wantStreamOpts, streamOpts)
				return nil
			}

			cmd := NewSSHCommand(p)

			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			_, actualErr := cmd.ExecuteC()
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "ExecStreamer ran", tc.wantExec, gotExec)
		})
	}
}
