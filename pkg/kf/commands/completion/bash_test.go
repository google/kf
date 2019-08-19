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

package completion

import (
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func TestAddBashCompletion(t *testing.T) {
	root := &cobra.Command{
		Use: "root",
	}

	child := &cobra.Command{
		Use: "child",
	}
	MarkArgCompletionSupported(child, "spaces")

	root.AddCommand(child)

	// Sanity check
	testutil.AssertEqual(t, "bash completion", "", root.BashCompletionFunction)

	AddBashCompletion(root)

	testutil.AssertNotBlank(t, "bash completion", root.BashCompletionFunction)
	testutil.AssertContainsAll(t, root.BashCompletionFunction, []string{
		"root_child",
		"__kf_name_spaces",
		"__kf_custom_func",
	})
}
