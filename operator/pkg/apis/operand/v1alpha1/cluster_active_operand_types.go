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
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// Verify that ClusterActiveOperand adheres to the appropriate interfaces.
var (
	// Check that we can create OwnerReferences to a ClusterActiveOperand.
	_ kmeta.OwnerRefable = (*ClusterActiveOperand)(nil)

	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*ClusterActiveOperand)(nil)
	// Not present for the namespaced forms.
	NamespaceDelegatesReady apis.ConditionType = "NamespaceDelegatesReady"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterActiveOperandSpec keeps track of ownerrefs for one immutable version of the Operand
// (determined by calculating a hash of the GVKs in the operand).
type ClusterActiveOperandSpec struct {
	// References to resources that should have k8s GC prevented by this
	// ActiveOperand.
	Live []LiveRef `json:"live"`
}

// DelegateRef is a reference from the cluster-scoped ClusterActiveOperand
// to the namespaced equivalent(s) for each namespace referenced.
type DelegateRef struct {
	Namespace string `json:"namespace"`
}

// ClusterActiveOperandStatus defines the observed state of ClusterActiveOperand
type ClusterActiveOperandStatus struct {
	duckv1.Status `json:",inline"`

	Delegates []DelegateRef `json:"delegates"`

	// ClusterLive are the references this cluster scoped operand is responsible for.
	ClusterLive []LiveRef `json: "clusterlive"`
}

// +genclient
// +genclient:nonNamespaced
// +genreconciler:krshapedlogic=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterActiveOperand is the Schema for the ClusterActiveOperands API
type ClusterActiveOperand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterActiveOperandSpec   `json:"spec"`
	Status ClusterActiveOperandStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterActiveOperandList contains a list of ClusterActiveOperand
type ClusterActiveOperandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterActiveOperand `json:"items"`
}
