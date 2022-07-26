// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kfvalidation

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfinformer "github.com/google/kf/v2/pkg/client/kf/informers/externalversions/kf/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// ServiceInstanceValidationCallback validates that an existing ServiceInstance is not part of a binding.
// It is intended to be used as a callback on a delete request.
func ServiceInstanceValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	serviceinstance := &v1alpha1.ServiceInstance{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, serviceinstance); err != nil {
		return err
	}

	serviceBindingInformer := ctx.Value(ServiceInstanceBindingInformerKey{}).(kfinformer.ServiceInstanceBindingInformer)
	serviceInstanceBindingLister := serviceBindingInformer.Lister()
	bindings, err := serviceInstanceBindingLister.ServiceInstanceBindings(serviceinstance.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}

	// matchedBindings holds the names of the apps the service instance is bound to.
	matchedBindings := sets.NewString()
	for _, binding := range bindings {
		if binding.Spec.InstanceRef.Name == serviceinstance.Name {
			matchedBindings.Insert(binding.Spec.BindingType.App.Name)
		}
	}

	if len(matchedBindings) > 0 {
		return fmt.Errorf("ServiceInstance %q cannot be deleted while it is part of a binding. The service is bound to the App(s): %s",
			serviceinstance.Name, strings.Join(matchedBindings.List(), ", "))
	}

	return nil
}
