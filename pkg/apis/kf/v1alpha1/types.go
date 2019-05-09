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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommandSet is a list of CommandSet.
type CommandSet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ContainerRegistry string        `json:"containerRegistry"`
	Spec              []CommandSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommandSetList is a list of CommandSet.
type CommandSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CommandSet `json:"items"`
}

// CommandSpec stores information about a command used by the CLI.
type CommandSpec struct {
	// Name of command.
	Name string `json:"name"`

	// Use of the command.
	Use string `json:"use"`

	// +optional
	// Long is the long description of the command.
	Long string `json:"long,omitempty"`

	// +optional
	// Short is the short description of the command.
	Short string `json:"short,omitempty"`

	// +optional
	// Flags are passed to the command.
	Flags []Flag `json:"flags,omitempty"`

	// The build-template to use.
	BuildTemplate string `json:"buildTemplate,omitempty"`

	// Upload the current directory. This is used for uploading source code
	// and the like.
	UploadDir bool `json:"uploadDir,omitempty"`
}

// Flag is a flag for a Command.
type Flag struct {
	// Type is the type of the argument (e.g., string, integer).
	Type string `json:"type,omitempty"`

	// +optional
	// Default is the default value of the argument.
	Default string `json:"default,omitempty"`

	// Long is the full flag name.
	Long string `json:"long,omitempty"`

	// +optional
	// Short is a one character flag name.
	Short string `json:"short,omitempty"`

	// +optional
	// Description is the description of the flag.
	Description string `json:"description,omitempty"`
}
