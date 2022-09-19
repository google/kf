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
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf"
	kfapis "github.com/google/kf/v2/pkg/apis/kf"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"knative.dev/pkg/apis"
)

const (
	protocolHTTP2 = "http2"
	protocolHTTP  = "http"
	protocolTCP   = "tcp"
)

// Validate checks for errors in the Application's fields.
func (app *Application) Validate(ctx context.Context) (errs *apis.FieldError) {
	// validate container execution
	if app.Command != "" {
		if len(app.Args) > 0 {
			errs = errs.Also(apis.ErrMultipleOneOf("command", "args"))
		}

		if app.Entrypoint != "" {
			errs = errs.Also(apis.ErrMultipleOneOf("entrypoint", "command"))
		}
	}

	// validate buildpacks
	if len(app.Buildpacks) > 0 {
		if app.LegacyBuildpack != "" {
			errs = errs.Also(apis.ErrMultipleOneOf("buildpack", "buildpacks"))
		}
	}

	errs = errs.Also(app.Ports.Validate(ctx).ViaField("ports"))

	errs = errs.Also(app.Metadata.Validate(ctx).ViaField("metadata"))

	okRoutePorts := sets.NewInt(0) // 0 means default
	for _, port := range app.Ports {
		okRoutePorts.Insert(int(port.Port))
	}
	for i, route := range app.Routes {
		if !okRoutePorts.Has(int(route.AppPort)) {
			errs = errs.Also(apis.ErrInvalidValue("must match a declared port", "appPort").ViaFieldIndex("routes", i))
		}
	}

	// Check that the probe fields are mutually exclusive
	hasCFHealthChecks := app.hasCFHealthCheckFields()
	hasK8sHealthChecks := app.hasK8sHealthCheckFields()

	switch {
	case hasCFHealthChecks && hasK8sHealthChecks:
		errs = errs.Also(&apis.FieldError{Message: "startupProbe, livenessProbe, and readinessProbe can't be used with CF health check fields"})
	case hasK8sHealthChecks:
		errs = errs.Also(kf.ValidateContainerProbe(app.StartupProbe).ViaField("startupProbe"))
		errs = errs.Also(kf.ValidateContainerProbe(app.LivenessProbe).ViaField("livenessProbe"))
		errs = errs.Also(kf.ValidateContainerProbe(app.ReadinessProbe).ViaField("readinessProbe"))
	case hasCFHealthChecks:
		// NOTE: https://docs.cloudfoundry.org/devguide/deploy-apps/healthchecks.html#health_check_timeout
		// says that officially the max timeout is 180, but checking that would likely be a
		// major breaking change with Kf because Kf was built before K8s supported startup probes so
		// longer timeouts were necessary to ensure apps would start.

		if timeout := app.HealthCheckTimeout; timeout < 0 {
			errs = errs.Also(apis.ErrInvalidValue(timeout, "timeout", "health check timeout can't be negative"))
		}

		if timeout := app.HealthCheckInvocationTimeout; timeout < 0 {
			errs = errs.Also(apis.ErrInvalidValue(timeout, "health-check-invocation-timeout", "health check timeout can't be negative"))
		}

		if app.HealthCheckType != "http" && app.HealthCheckHTTPEndpoint != "" {
			errs = errs.Also(apis.ErrInvalidValue(
				app.HealthCheckHTTPEndpoint,
				"health-check-http-endpoint",
				`field can only be set if health-check-type is "http"`))
		}

		allowedHealthCheckTypes := sets.NewString("http", "port", "", "process", "none")
		if !allowedHealthCheckTypes.Has(app.HealthCheckType) {
			errs = errs.Also(apis.ErrInvalidValue(
				app.HealthCheckType,
				"health-check-type",
				fmt.Sprintf("valid values are: %q", allowedHealthCheckTypes.List()),
			))
		}
	}

	return
}

// Validate implements apis.Validatable
func (a AppPortList) Validate(ctx context.Context) (errs *apis.FieldError) {
	seen := sets.NewInt()

	for i, port := range a {
		errs = errs.Also(port.Validate(ctx).ViaIndex(i))

		// ensure there are no duplicate port entries
		portInt := int(port.Port)
		if seen.Has(portInt) {
			errs = errs.Also(kfapis.ErrDuplicateValue(portInt, "port").ViaIndex(i))
		}
		seen.Insert(portInt)
	}

	return
}

// Validate implements apis.Validatable
func (a *AppPort) Validate(ctx context.Context) (errs *apis.FieldError) {
	// Validate port number
	errs = errs.Also(kfapis.ValidatePortNumberBounds(a.Port, "port"))

	// Validate protocol
	validProtocols := sets.NewString(protocolHTTP, protocolTCP, protocolHTTP2)
	if !validProtocols.Has(a.Protocol) {
		msg := fmt.Sprintf("must be one of: %v", validProtocols.List())
		errs = errs.Also(apis.ErrInvalidValue(msg, "protocol"))
	}

	return
}

// Validate implements apis.Validatable
func (a *ApplicationMetadata) Validate(ctx context.Context) (errs *apis.FieldError) {

	for _, err := range apivalidation.ValidateAnnotations(a.Annotations, field.NewPath("annotations")) {
		errs = errs.Also(&apis.FieldError{
			Message: err.ErrorBody(),
			Paths:   []string{err.Field},
		})
	}

	for _, err := range v1validation.ValidateLabels(a.Labels, field.NewPath("labels")) {
		errs = errs.Also(&apis.FieldError{
			Message: err.ErrorBody(),
			Paths:   []string{err.Field},
		})
	}

	return
}
