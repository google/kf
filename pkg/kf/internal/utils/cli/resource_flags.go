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

package utils

import "github.com/spf13/cobra"

// ResourceFlags is a flag set for intaking container resource requests (e.g. cpu/memory/disk).
type ResourceFlags struct {
	cpu    string
	disk   string
	memory string
}

// Add adds the resource flags to the Cobra command.
func (flags *ResourceFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&flags.cpu,
		"cpu-cores",
		"",
		"Amount of dedicated CPU cores to give the Task (for example 256M, 1024M, 1G).",
	)

	cmd.Flags().StringVarP(
		&flags.disk,
		"disk-quota",
		"k",
		"",
		"Amount of dedicated disk space to give the Task (for example 256M, 1024M, 1G).",
	)

	cmd.Flags().StringVarP(
		&flags.memory,
		"memory-limit",
		"m",
		"",
		"Amount of dedicated memory to give the Task (for example 256M, 1024M, 1G).",
	)
}

// CPU returns the cpu flag value.
func (flags *ResourceFlags) CPU() string {
	return flags.cpu
}

// Disk returns the disk flag value.
func (flags *ResourceFlags) Disk() string {
	return flags.disk
}

// Memory returns the memory flag value.
func (flags *ResourceFlags) Memory() string {
	return flags.memory
}
