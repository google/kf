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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	_ duckv1.KRShaped = (*Operand)(nil)
	// LatestActiveOperandReady is set when latest active operand becomes ready.
	LatestActiveOperandReady apis.ConditionType = "LatestActiveOperandReady"
	// OperandInstalled is set when operand is installed.
	OperandInstalled apis.ConditionType = "OperandInstalled"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperandSpec keeps track of ownerrefs for one immutable version of the Operand
// (determined by calculating a hash of the GVKs in the operand).
type OperandSpec struct {
	// The desired SteadyState. This is used to stamp out
	// an ActiveOperand, which is named by a checksum of this
	// Operands name and spec.
	SteadyState []unstructured.Unstructured `json:"steadyState,omitempty"`

	// Postinstall jobs and other resources to apply after the
	// SteadyState resources are ready.
	PostInstall []unstructured.Unstructured `json:"postInstall,omitempty"`

	CheckDeploymentHealth bool `json:"checkDeploymentHealth"`
}

// OperandStatus defines the observed state of Operand
type OperandStatus struct {
	duckv1.Status `json:",inline"`

	LatestReadyActiveOperand string `json:"latestReadyActiveOperand,omitempty"`

	LatestCreatedActiveOperand string `json:"latestCreatedActiveOperand,omitempty"`

	// Set to the operand's generation after successful
	// SteadyState install. Unset if the currently applied
	// SteadyState is not yet ready.
	InstalledSteadyStateGeneration int64 `json:"latestInstalledSteadyStateGeneration,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Operand is the Schema for the Operands API
type Operand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status OperandStatus `json:"status,omitempty"`
	Spec   OperandSpec   `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperandList contains a list of Operand
type OperandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Operand `json:"items"`
}
