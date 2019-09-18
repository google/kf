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

package manifest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
)

func TestCommand_UnmarshalJSON(t *testing.T) {
	cases := map[string]struct {
		raw             string
		expectedCommand Command
		expectedErr     error
	}{
		"string": {
			raw:             `"rake build $HOME/myscript.rb"`,
			expectedCommand: Command{LauncherPath, "rake build $HOME/myscript.rb"},
		},
		"array": {
			raw:             `["java", "-jar", "path/to/my.jar"]`,
			expectedCommand: Command{"java", "-jar", "path/to/my.jar"},
		},
		"object": {
			raw:         `{"foo":"bar"}`,
			expectedErr: ErrCommandUnmarshal,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			var cmd Command
			err := json.Unmarshal([]byte(tc.raw), &cmd)

			if err != nil || tc.expectedErr != nil {
				testutil.AssertErrorsEqual(t, err, tc.expectedErr)
				return
			}

			testutil.AssertEqual(t, "commands", tc.expectedCommand, cmd)
		})
	}
}

func ExampleCommand_blank() {
	cmd := Command(nil)
	fmt.Printf("Entyrpoint: %v\n", cmd.Entrypoint())
	fmt.Printf("Args: %v\n", cmd.Args())

	//cmd := Command{"/bin/sh", "-e", "echo $HOME"}

	// Output: Entyrpoint: []
	// Args: []
}

func ExampleCommand_filled() {
	cmd := Command{"/bin/sh", "-e", "echo $HOME"}
	fmt.Printf("Entyrpoint: %v\n", cmd.Entrypoint())
	fmt.Printf("Args: %v\n", cmd.Args())

	// Output: Entyrpoint: [/bin/sh]
	// Args: [-e echo $HOME]
}
