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
	"encoding/json"
	"fmt"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	val1 := json.RawMessage(`"value1"`)
	val2 := json.RawMessage(`"value2"`)
	credentialsSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-creds-secret",
		},
		Data: map[string][]byte{
			"key1": val1,
			"key2": val2,
		},
	}

	binding := v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-binding-generated-name",
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "my-app",
				},
			},
			InstanceRef: corev1.LocalObjectReference{
				Name: "my-instance",
			},
			BindingNameOverride: "custom-binding-name",
		},
		Status: v1alpha1.ServiceInstanceBindingStatus{
			BindingName: "custom-binding-name",
			CredentialsSecretRef: corev1.LocalObjectReference{
				Name: "my-creds-secret",
			},
			ServiceFields: v1alpha1.ServiceFields{
				Tags:      []string{"mysql"},
				ClassName: "my-service",
				PlanName:  "my-service-plan",
			},
		},
	}

	vs := cfutil.NewVcapService(binding, credentialsSecret)

	fmt.Printf("Name: %s\n", vs.Name)
	fmt.Printf("InstanceName: %s\n", vs.InstanceName)
	fmt.Printf("BindingName: %s\n", *vs.BindingName)
	fmt.Printf("Credentials: %s\n", vs.Credentials)
	fmt.Printf("Service: %v\n", vs.Label)
	fmt.Printf("Plan: %v\n", vs.Plan)
	fmt.Printf("Tags: %v\n", vs.Tags)

	// Output: Name: custom-binding-name
	// InstanceName: my-instance
	// BindingName: custom-binding-name
	// Credentials: map[key1:"value1" key2:"value2"]
	// Service: my-service
	// Plan: my-service-plan
	// Tags: [mysql]
}
