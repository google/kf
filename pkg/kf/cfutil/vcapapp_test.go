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

package cfutil_test

import (
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
)

func ExampleCreateVcapApplication() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	app.Namespace = "my-ns"
	app.UID = "12345"
	app.Status.Routes = []v1alpha1.AppRouteStatus{{URL: "app.example.com"}}

	envMap := cfutil.CreateVcapApplication(app)

	fmt.Println("VCAP_APPLICATION values:")
	fmt.Println("application_id:", envMap["application_id"])
	fmt.Println("application_name:", envMap["application_name"])
	fmt.Println("name:", envMap["name"])
	fmt.Println("application_uris:", envMap["application_uris"])
	fmt.Println("uris:", envMap["uris"])
	fmt.Println("space_name:", envMap["space_name"])
	fmt.Println("process_id:", envMap["process_id"])
	fmt.Println("process_type:", envMap["process_type"])

	// Output: VCAP_APPLICATION values:
	// application_id: 12345
	// application_name: my-app
	// name: my-app
	// application_uris: [app.example.com]
	// uris: [app.example.com]
	// space_name: my-ns
	// process_id: 12345
	// process_type: web
}
