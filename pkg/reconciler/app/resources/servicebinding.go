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
	"github.com/knative/serving/pkg/resources"
	svccatv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeServiceBindingLabels creates labels that can be used to tie a source to a build.
func MakeServiceBindingLabels(app *v1alpha1.App) map[string]string {
	return app.ComponentLabels("servicebinding")
}

func MakeServiceBindingName(app *v1alpha1.App, binding *v1alpha1.AppSpecServiceBinding) string {

	name := binding.BindingName
	if name == "" {
		name = binding.InstanceRef.Name
	}

	return fmt.Sprintf("kf-binding-%s-%s", app.Name, name)
}

func MakeServiceBindings(app *v1alpha1.App) ([]*svccatv1beta1.ServiceBinding, error) {
	var bindings []*svccatv1beta1.ServiceBinding
	for _, binding := range app.Spec.ServiceBindings {

		serviceBinding, err := MakeServiceBinding(app, &binding)
		if err != nil {
			return nil, err
		}

		bindings = append(bindings, serviceBinding)
	}
	return bindings, nil

}

func MakeServiceBinding(app *v1alpha1.App, binding *v1alpha1.AppSpecServiceBinding) (*svccatv1beta1.ServiceBinding, error) {

	return &svccatv1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeServiceBindingName(app, binding),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: resources.UnionMaps(app.GetLabels(), MakeServiceBindingLabels(app)),
		},
		Spec: svccatv1beta1.ServiceBindingSpec{
			InstanceRef: binding.InstanceRef,
		},
	}, nil
}
