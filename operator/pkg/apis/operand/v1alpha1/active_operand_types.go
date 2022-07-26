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
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	_ duckv1.KRShaped = (*ActiveOperand)(nil)
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ActiveOperandSpec keeps track of ownerrefs for one immutable version of the Operand
// (determined by calculating a hash of the GVKs in the operand).
type ActiveOperandSpec struct {
	// References to resources that should have k8s GC prevented by this
	// ActiveOperand.
	Live []LiveRef `json:"live"`
}

// ActiveOperandStatus defines the observed state of ActiveOperand
type ActiveOperandStatus struct {
	duckv1.Status `json:",inline"`
}

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveOperand is the Schema for the ActiveOperands API
type ActiveOperand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveOperandSpec   `json:"spec"`
	Status ActiveOperandStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveOperandList contains a list of ActiveOperand
type ActiveOperandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveOperand `json:"items"`
}
