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

package operand

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// Return values for GetState()

	// Error is the state that an error occurred while checking state.
	Error = "ERROR"
	// Installed is the state that the Operand is installed succesfully.
	Installed = "INSTALLED"
	// PendingChanges is the state that the Operator would like to continue making changes..
	PendingChanges = "PENDING_CHANGES"
)

// ResourceReconciler reconciles the operand's underlying resources.
type ResourceReconciler interface {
	// Apply applies the given resources.
	Apply(context.Context, []unstructured.Unstructured) error
	// GetState returns the current state of the given resources.
	//
	// Resulting string is guaranteed to be one of the constants
	// defined in this file or in the file for the particular
	// reconciler type.
	GetState(context.Context, []unstructured.Unstructured) (string, error)
}
