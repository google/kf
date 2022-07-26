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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// GroupName is the group name. This is used for CRDs.
	GroupName = "operand.run.cloud.google.com"

	// SchemaVersion is the Version of the schema. This is used for CRDs.
	SchemaVersion = "v1alpha1"
)

var (
	// SchemaGroupVersion is the overall version for the surface.
	SchemaGroupVersion = schema.GroupVersion{
		Group:   GroupName,
		Version: SchemaVersion,
	}
)

// Kind takes an unqualified kind and returns a GroupVersionKind
func Kind(kind string) schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(kind)
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package v1alpha1
// scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&ActiveOperand{},
		&ActiveOperandList{},
		&ClusterActiveOperand{},
		&ClusterActiveOperandList{},
		&Operand{},
		&OperandList{})
	metav1.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: SchemaVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds types to the scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
