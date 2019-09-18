// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"encoding/json"
	"errors"
)

const (
	// LauncherPath is the path to the buildpack launcher in a container.
	// It starts: https://github.com/buildpack/lifecycle/tree/master/cmd/launcher
	// which in turn launches a custom process or the default one set by the
	// buildpack.
	LauncherPath = "/lifecycle/launcher"
)

var (
	// ErrCommandUnmarshal is the error produced if Command couldn't be
	// deserialized from JSON.
	ErrCommandUnmarshal = errors.New("couldn't unmarshal command")
)

// NewLauncherCommand creates a command that gets run by the container launcher
// process. This command could either be a shell command, or a named process
// e.g. "web".
func NewLauncherCommand(launcherProcess string) Command {
	return Command{LauncherPath, launcherProcess}
}

// Command is a union type to represent the command to run in the container.
// If the manifest specifies a string then it is converted to a launched process
// in the container. However, if the user enters a Docker style array then the
// arguments are passed directly.
type Command []string

// UnmarshalJSON implements json.Unmarshaler
func (c *Command) UnmarshalJSON(data []byte) error {
	var commandArray []string
	if err := json.Unmarshal(data, &commandArray); err == nil {
		*c = commandArray
		return nil
	}

	var commandString string
	if err := json.Unmarshal(data, &commandString); err == nil {
		*c = NewLauncherCommand(commandString)
		return nil
	}

	return ErrCommandUnmarshal
}

// Entrypoint returns the container entrypoint if it's defined or nil.
// This can be set on a Pod's command field.
func (c *Command) Entrypoint() []string {
	if len(*c) > 0 {
		return (*c)[0:1]
	}

	return nil
}

// Args returns the container args if they're defined or nil.
func (c *Command) Args() []string {
	if len(*c) > 1 {
		return (*c)[1:]
	}

	return nil
}
