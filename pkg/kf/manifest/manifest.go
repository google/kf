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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
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
	Instances       *int32            `json:"instances,omitempty"`

	// Container command configuration
	Command string `json:"command,omitempty"`

	Routes      []Route `json:"routes,omitempty"`
	NoRoute     *bool   `json:"no-route,omitempty"`
	RandomRoute *bool   `json:"random-route,omitempty"`
	Task        *bool   `json:"task,omitempty"`

	// HealthCheckTimeout holds the health check timeout.
	// Note the serialized field is just timeout.
	HealthCheckTimeout int `json:"timeout,omitempty"`

	// HealthCheckType holds the type of health check that will be performed to
	// determine if the app is alive. Either port or http, blank means port.
	HealthCheckType string `json:"health-check-type,omitempty"`

	// HealthCheckHTTPEndpoint holds the HTTP endpoint that will receive the
	// get requests to determine liveness if HealthCheckType is http.
	HealthCheckHTTPEndpoint string `json:"health-check-http-endpoint,omitempty"`

	// HealthCheckInvocationTimeout is the timeout in seconds for individual
	// health check requests for HTTP and port health checks. By default this is 1.
	HealthCheckInvocationTimeout int `json:"health-check-invocation-timeout,omitempty"`

	// Metadata contains additional tags for applications and their underlying
	// resources.
	Metadata ApplicationMetadata `json:"metadata,omitempty"`

	// KfApplicationExtension holds fields that aren't officially in cf
	KfApplicationExtension `json:",inline"`
}

type ApplicationMetadata struct {
	// Annotations to set on the app instance.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels to set on the app instance.
	Labels map[string]string `json:"labels,omitempty"`
}

// KfApplicationExtension holds fields that aren't officially in cf
type KfApplicationExtension struct {
	CPU        string              `json:"cpu,omitempty"`
	CPULimit   string              `json:"cpu-limit,omitempty"`
	NoStart    *bool               `json:"no-start,omitempty"`
	Entrypoint string              `json:"entrypoint,omitempty"`
	Args       []string            `json:"args,omitempty"`
	Dockerfile Dockerfile          `json:"dockerfile,omitempty"`
	Build      *v1alpha1.BuildSpec `json:"build,omitempty"`
	Ports      AppPortList         `json:"ports,omitempty"`

	// Allow developers access to the underlying K8s probes because
	// CF is extremely limiting in this respect.

	StartupProbe   *corev1.Probe `json:"startupProbe,omitempty"`
	LivenessProbe  *corev1.Probe `json:"livenessProbe,omitempty"`
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
}

// AppPort represents an open port on an App.
type AppPort struct {
	// Port is the port number to open on the App. It's an int32 to match K8s.
	Port int32 `json:"port"`
	// Protocol is the protocol name of the port, either tcp, http or http2.
	// It's not an L4 protocol, but instead an L7 protocol.
	// The protocol name gets turned into port label that can be tracked
	// by Anthos Service Mesh so they have to be valid Istio protocols:
	// https://istio.io/docs/ops/configuration/traffic-management/protocol-selection/#manual-protocol-selection
	Protocol string `json:"protocol,omitempty"`
}

// AppPortList is a list of AppPort.
type AppPortList []AppPort

// AppDockerImage is the struct for docker configuration.
type AppDockerImage struct {
	Image string `json:"image,omitempty"`
}

// Route is a route name (including hostname, domain, and path) for an application.
type Route struct {
	Route   string `json:"route,omitempty"`
	AppPort int32  `json:"appPort,omitempty"`
}

// Manifest is an application's configuration.
type Manifest struct {
	// RelativePathRoot holds the directory that the manifest's paths are relative
	// to.
	RelativePathRoot string `json:"-"`

	Applications []Application `json:"applications"`
}

// Dockerfile contains the path to a Dockerfile to build.
type Dockerfile struct {
	Path string `json:"path,omitempty"`
}

// NewFromFile creates a Manifest from a manifest file.
func NewFromFile(
	ctx context.Context,
	manifestFile string,
	variables map[string]interface{},
) (*Manifest, error) {
	reader, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	return NewFromReader(ctx, reader, filepath.Dir(manifestFile), variables)
}

// NewFromReader creates a Manifest from a reader.
func NewFromReader(
	ctx context.Context,
	reader io.Reader,
	relativePathRoot string,
	variables map[string]interface{},
) (*Manifest, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	bytes, err = ApplySubstitution(bytes, variables)
	if err != nil {
		return nil, err
	}

	m := Manifest{}
	if strictError := yaml.UnmarshalStrict(bytes, &m); strictError != nil {

		// Fallback if strict unmarshaling doesn't work.
		if err := yaml.Unmarshal(bytes, &m); err != nil {
			return nil, fmt.Errorf("manifest appears invalid: %v", err)
		}

		// Print the strict error only if regular unmarshaling worked; the
		// YAML library wraps the type of error that caused the issue so
		// there's not a good way to tell if UnmarshalStrict failed due to
		// invalid structure or invalid configuration.
		logging.FromContext(ctx).Warnf("Manifest file contains unsupported config: %v\n", strictError)
	}

	m.RelativePathRoot = relativePathRoot

	return &m, nil
}

// New creates a Manifest for a single app relative to the current working
// directory.
func New(appName string) (*Manifest, error) {
	if appName == "" {
		return nil, errors.New("appName cannot be empty")
	}

	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Manifest{
		RelativePathRoot: path,
		Applications: []Application{
			{
				Name: appName,
			},
		},
	}, nil
}

// CheckForManifest looks for a manifest in the working directory.
// It returns a pointer to a Manifest if one is found, nil otherwise.
// An error is also returned in case an unexpected error occurs.
func CheckForManifest(
	ctx context.Context,
	variables map[string]interface{},
) (*Manifest, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, fileName := range []string{"manifest.yml", "manifest.yaml"} {
		filePath := filepath.Join(path, fileName)

		if _, err := os.Stat(filePath); err == nil {
			return NewFromFile(ctx, filePath, variables)
		}
	}

	return nil, nil
}

// App returns an Application by name.
func (m Manifest) App(name string) (*Application, error) {
	// If the manifest only has one App, override the App's name
	if len(m.Applications) == 1 {
		// To avoid changing m.Applications directly
		app := m.Applications[0]
		if name != "" {
			app.Name = name
		}
		return &app, nil
	}

	appNames := sets.NewString()
	for _, app := range m.Applications {
		if app.Name == name {
			return &app, nil
		}
		appNames.Insert(app.Name)
	}

	return nil, fmt.Errorf("the manifest doesn't have an App named %q, available names are: %q", name, appNames.List())
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

	// Default ports to TCP if not otherwise specified because HTTP traffic should
	// still be correctly classified under TCP, but not the other way around.
	for idx, port := range app.Ports {
		if port.Protocol == "" {
			app.Ports[idx].Protocol = protocolTCP
		}
	}

	if err := app.Validate(context.Background()); err.Error() != "" {
		return err
	}

	return nil
}

// WriteWarnings prints a message to the given writer if the user is
// using any kf specific fields in their configuration, any of the fields
// are deprecated or if the name had to have underscores swapped out.
func (app *Application) WriteWarnings(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	unofficialFields, err := kmp.CompareSetFields(app.KfApplicationExtension, KfApplicationExtension{})
	if err != nil {
		return err
	}

	if len(unofficialFields) != 0 {
		sort.Strings(unofficialFields)

		logger.Warnf("The field(s) %v are Kf-specific manifest extensions and may change.", unofficialFields)
	}

	for _, port := range app.Ports {
		if port.Protocol == protocolTCP {
			logger.Warn("Kf supports TCP ports but currently only HTTP Routes. " +
				"TCP ports can be reached on the App's cluster-internal app-<name>.<space>.svc.cluster.local address.")
			break // only show once
		}
	}

	// CF supports names with underscores however K8s does not. There might be
	// some existing apps that are migrated that use underscores. Replace any
	// underscore with a hyphen and throw up a warning.
	if strings.Contains(app.Name, "_") {
		logger.Warn("Underscores ('_') in names are not allowed in Kubernetes. Replacing with hyphens ('-')...")
		app.Name = strings.ReplaceAll(app.Name, "_", "-")
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

// BuildpacksSlice returns the buildpacks as a slice of strings.
// The legacy buildpack field is checked.
func (app *Application) BuildpacksSlice() []string {
	if len(app.Buildpacks) > 0 {
		return app.Buildpacks
	}

	if app.LegacyBuildpack != "" {
		return []string{app.LegacyBuildpack}
	}

	return nil
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
