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
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
)

const (
	// VcapApplicationEnvVarName is the environment variable expected by
	// applications looking for CF style app environment info.
	VcapApplicationEnvVarName = "VCAP_APPLICATION"
)

// CreateVcapApplication creates values for the VCAP_APPLICATION environment variable
// based on values on the app. These values are merged with the pod values determined at runtime.
func CreateVcapApplication(app *v1alpha1.App) map[string]interface{} {
	// You can find a list of values here:
	// https://docs.run.pivotal.io/devguide/deploy-apps/environment-variable.html

	urls := []string{}
	for _, r := range app.Status.Routes {
		urls = append(urls, r.URL)
	}

	values := map[string]interface{}{
		// application_id The GUID identifying the app.
		"application_id": app.UID,
		// application_name The name assigned to the app when it was pushed.
		"application_name": app.Name,
		// application_uris The URIs assigned to the app.
		"application_uris": urls,
		// name Identical to application_name.
		"name": app.Name,
		// process_id The UID identifying the process. Only present in running app containers.
		"process_id": app.UID,
		// process_type The type of process. Only present in running app containers.
		"process_type": "web",
		// space_name Human-readable name of the space where the app is deployed.
		"space_name": app.Namespace,
		// uris Identical to application_uris.
		"uris": urls,
	}

	return values
}
