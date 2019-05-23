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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Application is a configuration for a single 12-factor-app.
type Application struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}

// Manifest is an application's configuration.
type Manifest struct {
	Applications []Application `yaml:"applications"`
}

// NewFromFile creates a Manifest from a manifest file.
func NewFromFile(manifestFile string) (*Manifest, error) {
	reader, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	return NewFromReader(reader)
}

// NewFromReader creates a Manifest from a reader.
func NewFromReader(reader io.Reader) (*Manifest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// TODO: validate manifest
	m := Manifest{}
	if err = yaml.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// NewFromFileOrDefault creates a Manifest from a file.
// A backup appName can be provded if the given file does not exist.
func NewFromFileOrDefault(manifestFile, appName string) (*Manifest, error) {
	if manifestFile == "" && appName == "" {
		return nil, errors.New("manifestFile and appName cannot both be empty")
	}

	if manifestFile != "" {
		if _, err := os.Stat(manifestFile); err == nil {
			return NewFromFile(manifestFile)
		}

		// Fall through to default Manifest
	}

	if appName == "" {
		return nil, fmt.Errorf("manifest file %s did not exist. appName required", manifestFile)
	}

	return &Manifest{
		Applications: []Application{
			{
				Name: appName,
			},
		},
	}, nil
}

// App returns an Application by name.
func (m Manifest) App(name string) (*Application, error) {
	for _, app := range m.Applications {
		if app.Name == name {
			return &app, nil
		}
	}

	return nil, fmt.Errorf("no app %s found in the Manifest", name)
}
