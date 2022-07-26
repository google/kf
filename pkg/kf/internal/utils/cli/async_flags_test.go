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

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func ExampleAsyncFlags() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running async?", async.IsAsync())
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	cmd.ExecuteC()

	cmd.SetArgs([]string{"--async"})
	cmd.ExecuteC()

	// Output: Running async? false
	// Running async? true
}

func ExampleAsyncFlags_IsSynchronous() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running sync?", async.IsSynchronous())
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	cmd.ExecuteC()

	cmd.SetArgs([]string{"--async"})
	cmd.ExecuteC()

	// Output: Running sync? true
	// Running sync? false
}

func callbackTrueAfter(n int) func() (bool, error) {
	times := 0
	return func() (bool, error) {
		times += 1
		if times >= n {
			return true, nil
		}
		return false, nil
	}
}

func TestAsyncFlags_WaitFor(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      func() (context.Context, context.CancelFunc)
		action   string
		callback func() (bool, error)
	}
	cases := []struct {
		name    string
		args    args
		async   bool
		wantW   string
		wantErr error
	}{
		{
			name: "polls until true",
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.Background(), nil
				},
				action:   "waiting for true",
				callback: callbackTrueAfter(3),
			},
			async: false,
			wantW: "waiting for true...\nSuccess\n",
		},
		{
			name: "times out with context",
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.WithTimeout(context.Background(), 2*time.Millisecond)
				},
				action:   "waiting for true",
				callback: func() (bool, error) { return false, nil },
			},
			async:   false,
			wantW:   "waiting for true...\n",
			wantErr: fmt.Errorf("waiting for true timed out"),
		},
		{

			name: "returns error",
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.Background(), nil
				},
				action:   "waiting for true",
				callback: func() (bool, error) { return false, fmt.Errorf("error during action") },
			},
			async:   false,
			wantW:   "waiting for true...\n",
			wantErr: fmt.Errorf("error during action"),
		},
		{

			name: "asynchronous",
			args: args{
				ctx: func() (context.Context, context.CancelFunc) {
					return context.Background(), nil
				},
				action:   "waiting for true",
				callback: func() (bool, error) { return false, fmt.Errorf("not invoked") },
			},
			async: true,
			wantW: "waiting for true asynchronously\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			flags := &AsyncFlags{
				async: c.async,
			}
			w := &bytes.Buffer{}

			ctx, cancel := c.args.ctx()
			if cancel != nil {
				defer cancel()
			}
			actualErr := flags.WaitFor(ctx, w, c.args.action, time.Millisecond, c.args.callback)
			testutil.AssertErrorsEqual(t, c.wantErr, actualErr)

			testutil.AssertEqual(t, "output", c.wantW, w.String())
		})
	}
}

func TestAsyncFlags_AwaitAndLog(t *testing.T) {
	type fields struct {
		async bool
	}
	type args struct {
		action   string
		callback func() error
	}
	tests := map[string]struct {
		fields  fields
		args    args
		wantW   string
		wantErr error
	}{
		"async": {
			fields: fields{
				async: true,
			},
			args: args{
				action: "deleting foo in space bar",
				callback: func() error {
					return nil
				},
			},
			wantW: "deleting foo in space bar asynchronously\n",
		},
		"synchronous, no error": {
			fields: fields{
				async: false,
			},
			args: args{
				action: "deleting foo in space bar",
				callback: func() error {
					return nil
				},
			},
			wantW: "deleting foo in space bar...\nSuccess\n",
		},
		"synchronous, error": {
			fields: fields{
				async: false,
			},
			args: args{
				action: "deleting foo in space bar",
				callback: func() error {
					return errors.New("test-error")
				},
			},
			wantW:   "deleting foo in space bar...\n",
			wantErr: errors.New("test-error"),
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			flags := &AsyncFlags{
				async: tc.fields.async,
			}
			w := &bytes.Buffer{}

			actualErr := flags.AwaitAndLog(w, tc.args.action, tc.args.callback)
			testutil.AssertErrorsEqual(t, actualErr, tc.wantErr)

			testutil.AssertEqual(t, "output", tc.wantW, w.String())
		})
	}
}
