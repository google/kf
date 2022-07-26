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
	"os"
)

func ExampleSuggestNextAction() {

	SuggestNextAction(NextAction{
		Description: "View logs",
		Commands: []string{
			"kf logs my-app",
		},
	})

	PrintNextActions(os.Stdout)

	// Output:
	// TIP: These other commands could be useful:
	// View logs:  kf logs my-app
}

func ExampleSuggestNextAction_duplicate() {

	for i := 0; i < 10; i++ {
		SuggestNextAction(NextAction{
			Description: "View logs",
			Commands: []string{
				"kf logs my-app",
			},
		})
	}

	PrintNextActions(os.Stdout)

	// Output:
	// TIP: These other commands could be useful:
	// View logs:  kf logs my-app
}
