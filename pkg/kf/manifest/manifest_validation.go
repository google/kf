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

	"knative.dev/pkg/apis"
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

	return
}
