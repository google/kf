// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/commands/group"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func TestNewCommandGroup(t *testing.T) {
	cases := map[string]struct {
		commands        []*cobra.Command
		groupName       string
		args            []string
		expectedStrings []string
		expectedErr     error
	}{
		"Use": {
			groupName: "group",
			commands: []*cobra.Command{
				&cobra.Command{
					Use: "use-a",
				},
				&cobra.Command{
					Use: "use-b",
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			commandGroup, err := group.NewCommandGroup(tc.groupName, tc.commands...)
			testutil.AssertEqual(t, "error", tc.expectedErr, err)

			var b bytes.Buffer
			commandGroup.SetArgs(tc.args)
			commandGroup.SetOutput(&b)
			commandGroup.Execute()
			fmt.Fprintf(os.Stderr, b.String())
			if tc.expectedStrings != nil {
				actualStrings := strings.Split(b.String(), "\n")
				testutil.AssertEqual(t, "strings", tc.expectedStrings, actualStrings)
			}
		})
	}
}
