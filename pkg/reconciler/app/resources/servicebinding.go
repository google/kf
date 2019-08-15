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
	"encoding/json"
	"fmt"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/resources"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalog "github.com/poy/service-catalog/pkg/svcat/service-catalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"knative.dev/pkg/kmeta"
)

const (
	// AppNameLabel is the label used on bindings to define which app the binding belongs to.
	AppNameLabel = "kf-app-name"
	// BindingNameLabel is the label used on bindings to define what VCAP name the secret should be rooted under.
	BindingNameLabel = "kf-binding-name"
)

// MakeServiceBindingLabels creates labels that can be used to tie a source to a build.
func MakeServiceBindingLabels(app *v1alpha1.App, bindingName string) map[string]string {
	return resources.UnionMaps(app.ComponentLabels("servicebinding"), map[string]string{
		AppNameLabel:     app.Name,
		BindingNameLabel: bindingName,
	})
}

func MakeServiceBindingName(app *v1alpha1.App, binding *v1alpha1.AppSpecServiceBinding) string {
	return fmt.Sprintf("kf-binding-%s-%s", app.Name, binding.InstanceRef.Name)
}

// MakeServiceBindingAppSelector creates a labels.Selector for listing all the
// Service Bindings for the given App.
func MakeServiceBindingAppSelector(app *v1alpha1.App) labels.Selector {
	return labels.NewSelector().Add(
		mustRequirement(AppNameLabel, selection.Equals, app.Name),
	)
}

func MakeServiceBindings(app *v1alpha1.App) ([]servicecatalogv1beta1.ServiceBinding, error) {
	var bindings []servicecatalogv1beta1.ServiceBinding
	for _, binding := range app.Spec.ServiceBindings {
		serviceBinding, err := MakeServiceBinding(app, &binding)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, *serviceBinding)
	}
	return bindings, nil

}

func MakeServiceBinding(app *v1alpha1.App, binding *v1alpha1.AppSpecServiceBinding) (*servicecatalogv1beta1.ServiceBinding, error) {
	var params interface{}
	err := json.Unmarshal(binding.Parameters, &params)
	if err != nil {
		return nil, err
	}

	bindingName := binding.BindingName
	if bindingName == "" {
		bindingName = binding.InstanceRef.Name
	}

	return &servicecatalogv1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeServiceBindingName(app, binding),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: resources.UnionMaps(app.GetLabels(), MakeServiceBindingLabels(app, bindingName)),
		},
		Spec: servicecatalogv1beta1.ServiceBindingSpec{
			InstanceRef: binding.InstanceRef,
			Parameters:  servicecatalog.BuildParameters(params),
		},
	}, nil
}
