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

package logs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/logs"
	"github.com/google/kf/v2/pkg/kf/logs/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func TestLogsCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		Space  string
		Args   []string
		Setup  func(t *testing.T, fake *fake.FakeTailer)
		Assert func(t *testing.T, cmd *cobra.Command, err error)
	}{
		"missing app name": {
			Assert: func(t *testing.T, cmd *cobra.Command, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts 1 arg(s), received 0"), err)
			},
		},
		"missing space": {
			Args: []string{"some-app"},
			Assert: func(t *testing.T, cmd *cobra.Command, err error) {
				testutil.AssertErrorsEqual(t, errors.New(config.EmptySpaceError), err)
			},
		},
		"tailer returns error": {
			Args:  []string{"some-app"},
			Space: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeTailer) {
				fake.EXPECT().
					Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
			Assert: func(t *testing.T, cmd *cobra.Command, err error) {
				testutil.AssertEqual(t, "SilenceUsage", true, cmd.SilenceUsage)
				testutil.AssertErrorsEqual(t, errors.New("failed to tail logs: some-error"), err)
			},
		},
		"don't silence usage for configuration errors": {
			Args:  []string{"some-app"},
			Space: "some-space",
			Setup: func(t *testing.T, fake *fake.FakeTailer) {
				fake.EXPECT().
					Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(utils.ConfigErr{Reason: "some-error"})
			},
			Assert: func(t *testing.T, cmd *cobra.Command, err error) {
				testutil.AssertEqual(t, "SilenceUsage", false, cmd.SilenceUsage)
			},
		},
		"uses configuration": {
			Space: "some-space",
			Args:  []string{"some-app", "-n=15"},
			Setup: func(t *testing.T, fake *fake.FakeTailer) {
				fake.EXPECT().
					Tail(gomock.Not(gomock.Nil()), "some-app", gomock.Not(gomock.Nil()), gomock.Any()).
					Do(func(ctx context.Context, appName string, out io.Writer, opts ...logs.TailOption) {
						testutil.AssertEqual(t, "space", "some-space", logs.TailOptions(opts).Space())
						testutil.AssertEqual(t, "number lines", 15, logs.TailOptions(opts).NumberLines())
						testutil.AssertEqual(t, "follow", true, logs.TailOptions(opts).Follow())
					})
			},
		},
		"tail logs for App": {
			Space: "some-space",
			Args:  []string{"some-app"},
			Setup: func(t *testing.T, fake *fake.FakeTailer) {
				fake.EXPECT().
					Tail(gomock.Not(gomock.Nil()), "some-app", gomock.Not(gomock.Nil()), gomock.Any()).
					Do(func(ctx context.Context, appName string, out io.Writer, opts ...logs.TailOption) {
						testutil.AssertEqual(t, "component name", "app-server", logs.TailOptions(opts).ComponentName())
						testutil.AssertEqual(t, "containter name", "user-container", logs.TailOptions(opts).ContainerName())
					})
			},
		},
		"tail logs for Task": {
			Space: "some-space",
			Args:  []string{"some-app", "--task"},
			Setup: func(t *testing.T, fake *fake.FakeTailer) {
				fake.EXPECT().
					Tail(gomock.Not(gomock.Nil()), "some-app", gomock.Not(gomock.Nil()), gomock.Any()).
					Do(func(ctx context.Context, appName string, out io.Writer, opts ...logs.TailOption) {
						testutil.AssertEqual(t, "component name", "task", logs.TailOptions(opts).ComponentName())
						testutil.AssertEqual(t, "containter name", "step-user-container", logs.TailOptions(opts).ContainerName())
					})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Setup == nil {
				tc.Setup = func(t *testing.T, fake *fake.FakeTailer) {
					// NOP
				}
			}
			if tc.Assert == nil {
				tc.Assert = func(t *testing.T, cmd *cobra.Command, err error) {
					testutil.AssertNil(t, "err", err)
				}
			}

			ctrl := gomock.NewController(t)
			fake := fake.NewFakeTailer(ctrl)
			tc.Setup(t, fake)

			var buf bytes.Buffer
			cmd := NewLogsCommand(
				&config.KfParams{Space: tc.Space},
				fake,
			)
			cmd.SetArgs(tc.Args)
			cmd.SetOutput(&buf)

			gotErr := cmd.Execute()

			tc.Assert(t, cmd, gotErr)
			if gotErr != nil {
				return
			}

		})
	}
}
