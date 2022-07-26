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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterServiceBrokerValidationCallback validates that an existing ServiceBroker is not part of a ClusterServiceInstance.
// It is intended to be used as a callback on a delete request.
func ClusterServiceBrokerValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	clusterServiceBroker := &v1alpha1.ClusterServiceBroker{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, clusterServiceBroker); err != nil {
		return err
	}
	serviceInstanceInformer := ctx.Value(ServiceInstanceInformerKey{}).(kfinformer.ServiceInstanceInformer)
	serviceInstanceLister := serviceInstanceInformer.Lister()
	if err := validateClusterServiceBroker(serviceInstanceLister, clusterServiceBroker); err != nil {
		return err
	}
	return nil
}

// ServiceBrokerValidationCallback validates that an existing ServiceBroker is not part of a ServiceInstance.
// It is intended to be used as a callback on a delete request.
func ServiceBrokerValidationCallback(ctx context.Context, unstructured *unstructured.Unstructured) error {
	serviceBroker := &v1alpha1.ServiceBroker{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, serviceBroker); err != nil {
		return err
	}
	serviceInstanceInformer := ctx.Value(ServiceInstanceInformerKey{}).(kfinformer.ServiceInstanceInformer)
	serviceInstanceLister := serviceInstanceInformer.Lister()
	if err := validateServiceBroker(serviceInstanceLister, serviceBroker); err != nil {
		return err
	}
	return nil
}

func validateServiceBroker(serviceInstanceLister kflisters.ServiceInstanceLister, serviceBroker *v1alpha1.ServiceBroker) error {
	serviceInstances, err := serviceInstanceLister.ServiceInstances(serviceBroker.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	return validateServiceInstanceNotExists(serviceInstances, serviceBroker.Name)
}

func validateClusterServiceBroker(serviceInstanceLister kflisters.ServiceInstanceLister, clusterServiceBroker *v1alpha1.ClusterServiceBroker) error {
	serviceInstances, err := serviceInstanceLister.List(labels.Everything())
	if err != nil {
		return err
	}
	return validateServiceInstanceNotExists(serviceInstances, clusterServiceBroker.Name)
}

func validateServiceInstanceNotExists(serviceInstances []*v1alpha1.ServiceInstance, targetBrokerName string) error {
	for _, instance := range serviceInstances {
		var brokerName = ""
		switch {
		case instance.Spec.ServiceType.Brokered != nil:
			brokerName = instance.Spec.ServiceType.Brokered.Broker
		case instance.Spec.ServiceType.OSB != nil:
			brokerName = instance.Spec.ServiceType.OSB.BrokerName
		}
		if len(brokerName) > 0 {
			if brokerName == targetBrokerName {
				return fmt.Errorf("ServiceInstance %q at Space %q still exists for broker %q", instance.Name, instance.Namespace, targetBrokerName)
			}
		}
	}
	return nil
}
