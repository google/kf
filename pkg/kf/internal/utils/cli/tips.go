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
	"fmt"
	"io"
	"sync"

	"github.com/google/kf/v2/pkg/kf/describe"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NextAction is a user-facing suggested next action that can be taken after
// running a particular command.
type NextAction struct {
	// Description holds the user-readable description of the next action.
	// This should be short and start with a capital letter.
	Description string
	// Commands are a list of commands to show the user. These commands should be
	// equivalent e.g. a URL, a kubectl and a kf command that all achieve the
	// same action could be put together.
	Commands []string
}

var actionsSync sync.Mutex
var nextActions []NextAction

// SuggestNextAction adds a next action to take to the global list of next
// actions.
func SuggestNextAction(t NextAction) {
	actionsSync.Lock()
	defer actionsSync.Unlock()

	nextActions = append(nextActions, t)
}

// PrintNextActions prints all the available next actions and clears
// their values.
func PrintNextActions(w io.Writer) {
	actionsSync.Lock()
	defer actionsSync.Unlock()

	if len(nextActions) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s: These other commands could be useful:\n", TipColor.Sprintf("TIP"))

	describe.TabbedWriter(w, func(w io.Writer) {
		alreadyPrinted := sets.NewString()

		for _, action := range nextActions {
			// deduplicate by description
			if alreadyPrinted.Has(action.Description) {
				continue
			}
			alreadyPrinted.Insert(action.Description)

			first := true
			for _, cmd := range action.Commands {
				if first {
					fmt.Fprintf(w, "%s:\t%s\n", action.Description, cmd)
					first = false
				} else {
					fmt.Fprintf(w, "%s\t%s\n", "", cmd)
				}
			}
		}
	})

	nextActions = nil
}
