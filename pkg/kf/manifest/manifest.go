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
	"path/filepath"
	"strings"

	"github.com/google/kf/pkg/internal/envutil"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// Application is a configuration for a single 12-factor-app.
type Application struct {
	Name       string            `yaml:"name,omitempty"`
	Path       string            `yaml:"path,omitempty"`
	Buildpacks []string          `yaml:"buildpacks,omitempty"`
	Docker     AppDockerImage    `yaml:"docker,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Services   []string          `yaml:"services,omitempty"`
	MinScale   *int              `yaml:"minScale,omitempty"`
	MaxScale   *int              `yaml:"maxScale,omitempty"`
	Routes     []Route           `yaml:"routes,omitempty"`

	// HealthCheckTimeout holds the health check timeout.
	// Note the serialized field is just timeout.
	HealthCheckTimeout int `yaml:"timeout,omitempty"`

	// HealthCheckType holds the type of health check that will be performed to
	// determine if the app is alive. Either port or http, blank means port.
	HealthCheckType string `yaml:"health-check-type,omitempty"`

	// HealthCheckHTTPEndpoint holds the HTTP endpoint that will receive the
	// get requests to determine liveness if HealthCheckType is http.
	HealthCheckHTTPEndpoint string `yaml:"health-check-http-endpoint,omitempty"`
}

// AppDockerImage is the struct for docker configuration.
type AppDockerImage struct {
	Image string `yaml:"image,omitempty"`
}

// Route is a route name (including hostname, domain, and path) for an application.
type Route struct {
	Route string `yaml:"route,omitempty"`
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
	if err = yaml.UnmarshalStrict(bytes, &m); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: manifest file contains unsupported config: %v", err)
	} else if err = yaml.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// New creates a Manifest for a single app.
func New(appName string) (*Manifest, error) {
	if appName == "" {
		return nil, errors.New("appName cannot be empty")
	}

	return &Manifest{
		Applications: []Application{
			{
				Name: appName,
			},
		},
	}, nil
}

// CheckForManifest will optionally return a Manifest given a directory.
func CheckForManifest(directory string) (*Manifest, error) {
	dirFile, err := os.Stat(directory)
	if err != nil {
		return nil, err
	}

	if !dirFile.IsDir() {
		return nil, fmt.Errorf("expected %s to be a directory", directory)
	}

	for _, fileName := range []string{"manifest.yml", "manifest.yaml"} {
		filePath := filepath.Join(directory, fileName)

		if _, err := os.Stat(filePath); err == nil {
			return NewFromFile(filePath)
		}
	}

	return nil, nil
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

// Override overrides values using corresponding non-empty values from overrides.
// Environment variables are extended with override taking priority.
func (app *Application) Override(overrides *Application) error {
	appEnv := envutil.MapToEnvVars(app.Env)
	overrideEnv := envutil.MapToEnvVars(overrides.Env)
	combined := append(appEnv, overrideEnv...)

	if err := mergo.Merge(app, overrides, mergo.WithOverride); err != nil {
		return err
	}

	if len(combined) > 0 {
		app.Env = envutil.EnvVarsToMap(envutil.DeduplicateEnvVars(combined))
	}

	return nil
}

// Buildpack joings toegether the buildpacks in order as a CSV to be compatible
// with buildpacks v3.
func (app *Application) Buildpack() string {
	return strings.Join(app.Buildpacks, ",")
}
