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

package apps_test

import (
	"fmt"

	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
)

func ExampleBindService() {
	myApp := &kfv1alpha1.App{}
	apps.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		Instance:    "some-service",
		BindingName: "some-binding-name",
	})
	apps.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		Instance:    "another-service",
		BindingName: "some-binding-name",
	})
	apps.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		Instance:    "third-service",
		BindingName: "third",
	})
	apps.BindService(myApp, &kfv1alpha1.AppSpecServiceBinding{
		Instance:    "forth-service",
		BindingName: "forth",
	})
	apps.UnbindService(myApp, "third")

	for _, b := range myApp.Spec.ServiceBindings {
		fmt.Println("Instance", b.Instance, "BindingName", b.BindingName)
	}

	// Output: Instance another-service BindingName some-binding-name
	// Instance forth-service BindingName forth
}
