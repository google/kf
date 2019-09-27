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
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/kf/pkg/internal/envutil"
	"github.com/imdario/mergo"
	"knative.dev/pkg/kmp"
	"sigs.k8s.io/yaml"
)

// Application is a configuration for a single 12-factor-app.
type Application struct {
	Name            string            `json:"name,omitempty"`
	Path            string            `json:"path,omitempty"`
	LegacyBuildpack string            `json:"buildpack,omitempty"`
	Buildpacks      []string          `json:"buildpacks,omitempty"`
	Stack           string            `json:"stack,omitempty"`
	Docker          AppDockerImage    `json:"docker,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	Services        []string          `json:"services,omitempty"`
	DiskQuota       string            `json:"disk_quota,omitempty"`
	Memory          string            `json:"memory,omitempty"`
	Instances       *int              `json:"instances,omitempty"`

	// Container command configuration
	Command string `json:"command,omitempty"`

	Routes      []Route `json:"routes,omitempty"`
	NoRoute     *bool   `json:"no-route,omitempty"`
	RandomRoute *bool   `json:"random-route,omitempty"`

	// HealthCheckTimeout holds the health check timeout.
	// Note the serialized field is just timeout.
	HealthCheckTimeout int `json:"timeout,omitempty"`

	// HealthCheckType holds the type of health check that will be performed to
	// determine if the app is alive. Either port or http, blank means port.
	HealthCheckType string `json:"health-check-type,omitempty"`

	// HealthCheckHTTPEndpoint holds the HTTP endpoint that will receive the
	// get requests to determine liveness if HealthCheckType is http.
	HealthCheckHTTPEndpoint string `json:"health-check-http-endpoint,omitempty"`

	// KfApplicationExtension holds fields that aren't officially in cf
	KfApplicationExtension `json:",inline"`
}

// KfApplicationExtension holds fields that aren't officially in cf
type KfApplicationExtension struct {
	// TODO(#95): These aren't CF proper. How do we expose these in the manifest?

	CPU string `json:"cpu,omitempty"`

	MinScale *int  `json:"min-scale,omitempty"`
	MaxScale *int  `json:"max-scale,omitempty"`
	NoStart  *bool `json:"no-start,omitempty"`

	EnableHTTP2 *bool `json:"enable-http2,omitempty"`

	Entrypoint string   `json:"entrypoint,omitempty"`
	Args       []string `json:"args,omitempty"`
}

// AppDockerImage is the struct for docker configuration.
type AppDockerImage struct {
	Image string `json:"image,omitempty"`
}

// Route is a route name (including hostname, domain, and path) for an application.
type Route struct {
	Route string `json:"route,omitempty"`
}

// Manifest is an application's configuration.
type Manifest struct {
	Applications []Application `json:"applications"`
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

	if overrides.RandomRoute != nil {
		app.RandomRoute = overrides.RandomRoute
	}

	if overrides.NoRoute != nil {
		app.NoRoute = overrides.NoRoute

		if *app.NoRoute && app.RandomRoute != nil && *app.RandomRoute {
			return errors.New("can not use random-route and no-route together")
		}
	}
	if len(overrides.Routes) > 0 {
		app.Routes = overrides.Routes
	}

	if err := mergo.Merge(app, overrides, mergo.WithOverride); err != nil {
		return err
	}

	if len(combined) > 0 {
		app.Env = envutil.EnvVarsToMap(envutil.DeduplicateEnvVars(combined))
	}

	if err := app.Validate(context.Background()); err.Error() != "" {
		return err
	}

	return nil
}

// WarnUnofficialFields prints a message to the given writer if the user is
// using any kf specific fields in their configuration.
func (app *Application) WarnUnofficialFields(w io.Writer) error {
	// TODO(#95) Warn the user about using unofficial fields that are subject to
	// change.
	unofficialFields, err := kmp.CompareSetFields(app.KfApplicationExtension, KfApplicationExtension{})
	if err != nil {
		return err
	}

	if len(unofficialFields) != 0 {
		sort.Strings(unofficialFields)

		fmt.Fprintln(w)
		fmt.Fprintf(w, `WARNING! The field(s) %v are Kf extensions to the manifest and are subject to change.
See https://github.com/google/kf/issues/95 for more info.`, unofficialFields)
		fmt.Fprintln(w)
	}

	return nil
}

// Buildpack joins together the buildpacks in order as a CSV to be compatible
// with buildpacks v3. If no buildpacks are specified, the legacy buildpack
// field is checked.
func (app *Application) Buildpack() string {
	if len(app.Buildpacks) > 0 {
		return strings.Join(app.Buildpacks, ",")
	}

	return app.LegacyBuildpack
}

// CommandEntrypoint gets an override for the entrypoint of the container.
func (app *Application) CommandEntrypoint() []string {
	if app.Entrypoint != "" {
		return []string{app.Entrypoint}
	}

	return nil
}

// CommandArgs returns the container args if they're defined or nil.
func (app *Application) CommandArgs() []string {
	if len(app.Args) > 0 {
		return app.Args
	}

	if app.Command != "" {
		return []string{app.Command}
	}

	return nil
}
