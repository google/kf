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

package main

import (
	"os"

	"github.com/spf13/cobra"
)

// NewWriteResultCommand creates a command that will write the results of built to the file at provided location
func NewWriteResultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "write-result RESULT PATH",
		Short: "write-result file RESULT to file at PATH",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, path := args[0], args[1]
			os.WriteFile(path, []byte(result), 0777)

			return nil
		},
	}
}
