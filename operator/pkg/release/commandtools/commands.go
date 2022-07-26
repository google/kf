/*
Copyright 2020 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commandtools

import (
	"log"
	"os/exec"
	"syscall"
)

// ExecuteTerminalCommand - Execute command with arguments and log its output. Return command output
// Terminates if the command returns non-zero exit code
func ExecuteTerminalCommand(command string, arguments ...string) string {
	cmd := exec.Command(command, arguments...)
	// Combine both standard outputs for simplicity
	// Long-term, we want to avoid command invocation and integrate libraries and tools
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Command output:\n%s", string(output))
		exitCode := 0 // Default exit code to 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Command has exited with an exit code different from 0
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
			log.Printf("Exit code: %d", exitCode)
			log.Printf("stderr: %s", exitErr.Stderr)
		}
		log.Fatalf("Failed to run command \"%s %v\". Error: %v", command, arguments, err)
	}

	return string(output)
}

// RunGsCpUtil - Run gsutil command with given arguments
func RunGsCpUtil(arguments ...string) string {
	// Invoke ko resolve command
	args := append([]string{"cp"}, arguments...)
	return ExecuteTerminalCommand("gsutil", args...)
}
