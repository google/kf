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
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func ExampleAsyncFlags() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := fmt.Println("Running async?", async.IsAsync()); err != nil {
				panic(err)
			}
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	if _, err := cmd.ExecuteC(); err != nil {
		panic(err)
	}

	cmd.SetArgs([]string{"--async"})
	if _, err := cmd.ExecuteC(); err != nil {
		panic(err)
	}

	// Output: Running async? false
	// Running async? true
}

func ExampleAsyncFlags_IsSynchronous() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := fmt.Println("Running sync?", async.IsSynchronous()); err != nil {
				panic(err)
			}
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	if _, err := cmd.ExecuteC(); err != nil {
		panic(err)
	}

	cmd.SetArgs([]string{"--async"})
	if _, err := cmd.ExecuteC(); err != nil {
		panic(err)
	}

	// Output: Running sync? true
	// Running sync? false
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

			testutil.AssertEqual(t, "output", w.String(), tc.wantW)
		})
	}
}
