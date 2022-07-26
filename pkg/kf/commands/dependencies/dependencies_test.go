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

package dependencies

import (
	"net/url"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewDependencyCommand_sanity(t *testing.T) {
	rootCmd := NewDependencyCommand()

	var commands []*cobra.Command
	commands = append(commands, rootCmd)
	commands = append(commands, rootCmd.Commands()...)

	for _, command := range commands {
		t.Run("command:"+command.Name(), func(t *testing.T) {
			testutil.AssertTrue(t, "command is hidden", command.Hidden)

			_, skipVersionCheckOk := command.Annotations[config.SkipVersionCheckAnnotation]
			testutil.AssertTrue(t, "skip version check", skipVersionCheckOk)

			testutil.AssertContainsAll(t, command.Long, []string{documentationOnly})
		})
	}
}

func TestNewDependencies(t *testing.T) {
	knownNames := sets.NewString()

	for _, dependency := range newDependencies() {
		t.Run("dependency:"+dependency.Name, func(t *testing.T) {
			// assert no name conflicts
			depNames := dependency.names()
			if overlap := knownNames.Intersection(depNames); len(overlap) > 0 {
				t.Errorf("expected no name overlap, but %v were already declared", overlap.List())
			}
			knownNames = knownNames.Union(depNames)

			// assert InfoURL is valid
			if _, err := url.ParseRequestURI(dependency.InfoURL); err != nil {
				t.Errorf("InfoURL isn't a valid URL: %v", err)
			}

			// assert the resolve functions exist, trying to
			// invoke they may not work outside of a build
			testutil.AssertTrue(t, "ResolveVersion is not nil", dependency.ResolveVersion != nil)
			testutil.AssertTrue(t, "ResolveURL is not nil", dependency.ResolveURL != nil)
		})
	}
}
