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

	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/cfutil"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func ExampleVcapServicesMap_Add() {
	m := cfutil.VcapServicesMap{}
	m.Add(cfutil.VcapService{InstanceName: "instance-a", Label: "foo"})
	m.Add(cfutil.VcapService{InstanceName: "instance-b", Label: "foo"})

	// Elements are registered by their Label.
	fmt.Printf("Number of bindings: %d\n", len(m))
	fmt.Printf("Binding 0: %s, Instance: %s\n", m["foo"][0].Label, m["foo"][0].InstanceName)
	fmt.Printf("Binding 1: %s, Instance: %s\n", m["foo"][1].Label, m["foo"][1].InstanceName)

	// Output: Number of bindings: 1
	// Binding 0: foo, Instance: instance-a
	// Binding 1: foo, Instance: instance-b
}

func ExampleNewVcapService() {
	instance := apiv1beta1.ServiceInstance{}
	instance.Name = "my-instance"
	instance.Spec.ServiceClassExternalName = "my-service"
	instance.Spec.ServicePlanExternalName = "my-service-plan"

	binding := apiv1beta1.ServiceBinding{}
	binding.Spec.InstanceRef.Name = "my-instance"
	binding.Name = "my-binding"
	binding.Labels = map[string]string{
		kfv1alpha1.ComponentLabel: "custom-binding-name",
	}

	secret := corev1.Secret{}
	secret.Data = map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	class := apiv1beta1.CommonServiceClassSpec{
		Tags: []string{"mysql"},
	}

	vs := cfutil.NewVcapService(class, instance, binding, &secret)

	fmt.Printf("Name: %s\n", vs.Name)
	fmt.Printf("InstanceName: %s\n", vs.InstanceName)
	fmt.Printf("BindingName: %s\n", vs.BindingName)
	fmt.Printf("Credentials: %v\n", vs.Credentials)
	fmt.Printf("Service: %v\n", vs.Label)
	fmt.Printf("Plan: %v\n", vs.Plan)
	fmt.Printf("Tags: %v\n", vs.Tags)

	// Output: Name: custom-binding-name
	// InstanceName: my-instance
	// BindingName: custom-binding-name
	// Credentials: map[key1:value1 key2:value2]
	// Service: my-service
	// Plan: my-service-plan
	// Tags: [mysql]
}
