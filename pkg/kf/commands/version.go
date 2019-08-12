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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand returns a command that displays the version.
func NewVersionCommand(versionStr, goos string) *cobra.Command {
	return newVersion(func(v version, cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "kf version", versionStr, goos)
		return nil
	})
}

// Version is filled in via ldflags.
var Version = "dev"
