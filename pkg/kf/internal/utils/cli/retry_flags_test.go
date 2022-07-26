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

package utils

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func ExampleRetryFlags() {
	var retry RetryFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			times := 0
			retry.Retry(func() error {
				fmt.Println("Running:", times)
				times++
				return errors.New("trigger-retry")
			})
		},
	}
	retry.Add(cmd, 3, 10*time.Millisecond)

	cmd.SetArgs([]string{})
	cmd.ExecuteC()

	// Output: Running: 0
	// Running: 1
	// Running: 2
}

func TestRetryFlags(t *testing.T) {
	t.Parallel()

	testErr1 := errors.New("test1")
	testErr2 := errors.New("test2")

	cases := map[string]struct {
		retryTimes int
		results    []error
		wantErr    error
	}{
		"first good": {
			retryTimes: 5,
			results:    []error{nil},
			wantErr:    nil,
		},
		"eventually good": {
			retryTimes: 5,
			results:    []error{testErr1, testErr2, nil},
			wantErr:    nil,
		},
		"return last err on failure": {
			retryTimes: 2,
			results:    []error{testErr1, testErr2},
			wantErr:    testErr2,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			var retry RetryFlags
			retry.retries = tc.retryTimes

			count := 0
			gotErr := retry.Retry(func() error {
				defer func() { count++ }()
				return tc.results[count]
			})

			testutil.AssertEqual(t, "calls", len(tc.results), count)
			testutil.AssertEqual(t, "error", tc.wantErr, gotErr)
		})
	}
}

func TestRetryFlags_AddRetryForK8sPropagation(t *testing.T) {
	// validity check for the defaults
	var retry RetryFlags

	cmd := &cobra.Command{}
	retry.AddRetryForK8sPropagation(cmd)

	testutil.AssertTrue(t, "retries multiple times", retry.retries >= 2)
	testutil.AssertTrue(t, "delay >= 1 second", retry.delay >= 1*time.Second)
}
