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

// Command is a union type to represent the command to run in the contianer.
// If the manifest specifies a string then it is converted to a launched process
// in the contianer. However, if the user enters a Docker style array then the
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
		*c = []string{LauncherPath, commandString}
		return nil
	}

	return ErrCommandUnmarshal
}
