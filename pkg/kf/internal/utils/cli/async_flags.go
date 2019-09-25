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
	"github.com/spf13/cobra"
)

// AsyncFlags is a flag set for managing whether or not a command runs
// asynchronously.
type AsyncFlags struct {
	async bool
}

// Add adds the async flag to the Cobra command.
func (flags *AsyncFlags) Add(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&flags.async,
		"async",
		"",
		false,
		"Don't wait for the action to complete on the server before returning",
	)
}

// IsAsync returns true if the user wanted the operation to run asynchronously.
func (flags *AsyncFlags) IsAsync() bool {
	return flags.async
}

// IsSynchronous returns true if the user wants the operation to be completed
// synchronously.
func (flags *AsyncFlags) IsSynchronous() bool {
	return !flags.async
}
