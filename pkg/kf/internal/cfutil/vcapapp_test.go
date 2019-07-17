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

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/internal/cfutil"
)

func ExampleCreateVcapApplication() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	app.Namespace = "my-ns"

	env, err := cfutil.CreateVcapApplication(app)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", env.Name, "Value:", env.Value)

	// Output: Name: VCAP_APPLICATION Value: {"application_name":"my-app","name":"my-app","space_name":"my-ns"}
}
