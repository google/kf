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

package reconcilerutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// HasFinalizer checks if an object has a finalizer.
func HasFinalizer(obj metav1.Object, finalizer string) bool {
	return sets.NewString(obj.GetFinalizers()...).Has(finalizer)
}

// NOTE: Some finalizer functions are aliased from controllerutil so we can
// update/modify if needed to continue working in conjunction.

// AddFinalizer adds a finalizer to the given object.
var AddFinalizer = controllerutil.AddFinalizer

// RemoveFinalizer removes a finalizer from the given object.
var RemoveFinalizer = controllerutil.RemoveFinalizer
