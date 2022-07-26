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
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// PackageChecksumType is the name of the sha256 type.
	PackageChecksumSHA256Type = "sha256"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourcePackage is responsible for storing the metadata about the source code
// bits.
type SourcePackage struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec SourcePackageSpec `json:"spec,omitempty"`

	// +optional
	Status SourcePackageStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*SourcePackage)(nil)
var _ apis.Defaultable = (*SourcePackage)(nil)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourcePackageList is a list of SourcePackage resources.
type SourcePackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SourcePackage `json:"items"`
}

// SourcePackageSpec is the desired configuration for a SourcePackage.
type SourcePackageSpec struct {
	// Checksum has the checksum information.
	Checksum SourcePackageChecksum `json:"checksum,omitempty"`

	// Size has the number of bytes of the package.
	Size uint64 `json:"size,omitempty"`
}

// SourcePackageChecksum has the checksum information for the SourcePackage
// bits.
type SourcePackageChecksum struct {
	// Type is the type of checksum used.
	// The allowed values are (more might be added in the future):
	// * sha256
	Type string `json:"type,omitempty"`

	// Value is the hex encoded checksum of the package bits.
	Value string `json:"value,omitempty"`
}

// SourcePackageStatus is the current configuration and running state for a
// SourcePackage.
type SourcePackageStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// Image is the fully qualified image name that has stored the underlying
	// data.
	Image string `json:"image,omitempty"`

	// Checksum has the checksum information.
	Checksum SourcePackageChecksum `json:"checksum,omitempty"`

	// Size has the number of bytes of the package.
	Size uint64 `json:"size,omitempty"`
}
