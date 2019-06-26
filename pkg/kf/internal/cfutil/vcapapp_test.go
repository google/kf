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

	"github.com/google/kf/pkg/kf/internal/cfutil"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

func ExampleCreateVcapApplication() {
	svc := &serving.Service{}
	svc.Name = "my-app"
	svc.Namespace = "my-ns"

	env, err := cfutil.CreateVcapApplication(svc)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", env.Name, "Value:", env.Value)

	// Output: Name: VCAP_APPLICATION Value: {"application_name":"my-app","name":"my-app","space_name":"my-ns"}
}
