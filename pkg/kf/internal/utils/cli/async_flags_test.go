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

	"github.com/spf13/cobra"
)

func ExampleAsyncFlags() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running async?", async.IsAsync())
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	cmd.ExecuteC()

	cmd.SetArgs([]string{"--async"})
	cmd.ExecuteC()

	// Output: Running async? false
	// Running async? true
}

func ExampleAsyncFlags_IsSynchronous() {
	var async AsyncFlags

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running sync?", async.IsSynchronous())
		},
	}
	async.Add(cmd)

	cmd.SetArgs([]string{})
	cmd.ExecuteC()

	cmd.SetArgs([]string{"--async"})
	cmd.ExecuteC()

	// Output: Running sync? true
	// Running sync? false
}
