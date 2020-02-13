// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cfutil

import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	corev1 "k8s.io/api/core/v1"
)

const (
	// VcapApplicationEnvVarName is the environment variable expected by
	// applications looking for CF style app environment info.
	VcapApplicationEnvVarName = "VCAP_APPLICATION"
)

// CreateVcapApplication creates a VCAP_APPLICATION style environment variable
// based on the values on the given service.
func CreateVcapApplication(app *v1alpha1.App) (corev1.EnvVar, error) {
	// You can find a list of values here:
	// https://docs.run.pivotal.io/devguide/deploy-apps/environment-variable.html

	// XXX: The values here are incomplete but are currently the best we can do.
	values := map[string]interface{}{
		// application_name The name assigned to the app when it was pushed.
		"application_name": app.Name,
		// name Identical to application_name.
		"name": app.Name,
		// space_name	Human-readable name of the space where the app is deployed.
		"space_name": app.Namespace,
	}

	return envutil.NewJSONEnvVar(VcapApplicationEnvVarName, values)
}
