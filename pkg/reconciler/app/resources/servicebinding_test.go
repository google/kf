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

package resources

import (
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func ExampleMakeServiceBindingLabels() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	binding := &v1alpha1.AppSpecServiceBinding{
		InstanceRef: servicecatalogv1beta1.LocalObjectReference{
			Name: "my-service",
		},
		BindingName: "cool-binding",
	}
	app.Spec.ServiceBindings = []v1alpha1.AppSpecServiceBinding{*binding}

	labels := MakeServiceBindingLabels(app, binding.BindingName)
	for k, v := range labels {
		fmt.Println(k, ":", v)
	}

	// Output: app.kubernetes.io/managed-by : kf
	// app.kubernetes.io/name : my-app
	// app.kubernetes.io/component : servicebinding
	// kf-app-name : my-app
	// kf-binding-name : cool-binding
}

func ExampleMakeServiceBindingName() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	binding := &v1alpha1.AppSpecServiceBinding{
		InstanceRef: servicecatalogv1beta1.LocalObjectReference{
			Name: "my-service",
		},
	}
	app.Spec.ServiceBindings = []v1alpha1.AppSpecServiceBinding{*binding}

	fmt.Println(MakeServiceBindingName(app, binding))

	// Output: kf-binding-my-app-my-service
}
