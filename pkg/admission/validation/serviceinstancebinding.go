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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfinformer "github.com/google/kf/v2/pkg/client/kf/informers/externalversions/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceInstanceBindingValidationCallback validates that the App and ServiceInstance referenced in a ServiceInstanceBinding exist.
// It is intended to be used as a callback on create and update requests.
func ServiceInstanceBindingValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	serviceinstancebinding := &v1alpha1.ServiceInstanceBinding{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, serviceinstancebinding); err != nil {
		return err
	}

	// Do not check App /Service Instance existence if binding is being deleted. It allows removing the finalizers on the binding.
	if serviceinstancebinding.DeletionTimestamp != nil {
		return nil
	}

	if serviceinstancebinding.IsAppBinding() {
		appInformer := ctx.Value(AppInformerKey{}).(kfinformer.AppInformer)
		appLister := appInformer.Lister()
		if err := validateAppExists(appLister, serviceinstancebinding); err != nil {
			return err
		}
	}

	serviceInstanceInformer := ctx.Value(ServiceInstanceInformerKey{}).(kfinformer.ServiceInstanceInformer)
	serviceInstanceLister := serviceInstanceInformer.Lister()
	if err := validateServiceInstanceExists(serviceInstanceLister, serviceinstancebinding); err != nil {
		return err
	}
	return nil
}

func validateAppExists(appLister kflisters.AppLister, serviceinstancebinding *v1alpha1.ServiceInstanceBinding) error {
	_, err := appLister.Apps(serviceinstancebinding.Namespace).Get(serviceinstancebinding.Spec.App.Name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("App %q does not exist. The binding cannot be created", serviceinstancebinding.Spec.App.Name)
	}
	return err
}

func validateServiceInstanceExists(serviceInstanceLister kflisters.ServiceInstanceLister, serviceinstancebinding *v1alpha1.ServiceInstanceBinding) error {
	_, err := serviceInstanceLister.ServiceInstances(serviceinstancebinding.Namespace).Get(serviceinstancebinding.Spec.InstanceRef.Name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("ServiceInstance %q does not exist. The binding cannot be created", serviceinstancebinding.Spec.InstanceRef.Name)
	}
	return err
}
