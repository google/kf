package servicebindings_test

import (
	"fmt"

	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	apiv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func ExampleVcapServicesMap_Add() {
	m := servicebindings.VcapServicesMap{}
	m.Add(servicebindings.VcapService{BindingName: "foo", InstanceName: "instance-a"})
	m.Add(servicebindings.VcapService{BindingName: "foo", InstanceName: "instance-b"})

	// Elements are registered by their binding name, if multiple instances exist
	// with the same name then the last one wins.
	fmt.Printf("Number of bindings: %d\n", len(m))
	fmt.Printf("Binding: %s, Instance: %s\n", m["foo"].BindingName, m["foo"].InstanceName)

	// Output: Number of bindings: 1
	// Binding: foo, Instance: instance-b
}

func ExampleNewVcapService() {
	binding := apiv1beta1.ServiceBinding{}
	binding.Spec.InstanceRef.Name = "my-instance"
	binding.Name = "my-binding"
	binding.Labels = map[string]string{
		servicebindings.BindingNameLabel: "custom-binding-name",
	}

	secret := corev1.Secret{}
	secret.Data = map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	vs := servicebindings.NewVcapService(binding, &secret)

	fmt.Printf("Name: %s\n", vs.Name)
	fmt.Printf("InstanceName: %s\n", vs.InstanceName)
	fmt.Printf("BindingName: %s\n", vs.BindingName)
	fmt.Printf("Credentials: %v\n", vs.Credentials)

	// Output: Name: my-binding
	// InstanceName: my-instance
	// BindingName: custom-binding-name
	// Credentials: map[key1:value1 key2:value2]
}
