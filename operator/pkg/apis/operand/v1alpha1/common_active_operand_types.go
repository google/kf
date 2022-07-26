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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var (
	// OwnerRefsInjected is set when an owner reference is injected.
	OwnerRefsInjected apis.ConditionType = "OwnerRefInjected"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LiveRef is a reference to an object that this ActiveOperand keeps alive.
type LiveRef struct {
	Group string `json:"group"`
	// Version and Resource are deprecated.
	// If they are set, they will be used.
	// If unset, Kind must be set.
	Version   string `json:"version"`
	Resource  string `json:"resource"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// GroupKind returns the GroupKind encoded in the ref
func (l LiveRef) GroupKind() *schema.GroupKind {
	if l.Kind == "" {
		return nil
	}
	return &schema.GroupKind{Group: l.Group, Kind: l.Kind}
}

// GroupVersionResource returns GroupVersionResource of LiveRef resource,
// or nil.
func (l LiveRef) GroupVersionResource() *schema.GroupVersionResource {
	if l.Version == "" || l.Resource == "" {
		return nil
	}
	return &schema.GroupVersionResource{
		Group:    l.Group,
		Version:  l.Version,
		Resource: l.Resource,
	}
}
