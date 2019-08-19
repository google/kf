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

package completion

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	// AppCompletion is the type for completing apps
	AppCompletion = "apps"

	// SourceCompletion is the type for completing sources
	SourceCompletion = "sources"

	// SpaceCompletion is the type for completing spaces
	SpaceCompletion = "spaces"
)

var namespacedTypes = map[string]schema.GroupVersionResource{
	AppCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "apps",
	},

	SourceCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "sources",
	},
}

var globalTypes = map[string]schema.GroupVersionResource{
	SpaceCompletion: {
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "spaces",
	},
}

// KnownGenericTypes returns the keys for all registered generic types.
func KnownGenericTypes() (out []string) {
	for k := range namespacedTypes {
		out = append(out, k)
	}

	for k := range globalTypes {
		out = append(out, k)
	}

	// make ordering deterministic
	sort.Strings(out)

	return
}

func getResourceInterface(client dynamic.Interface, k8sType, ns string) (dynamic.ResourceInterface, error) {
	if resource, ok := namespacedTypes[k8sType]; ok {
		return client.Resource(resource).Namespace(ns), nil
	}

	if resource, ok := globalTypes[k8sType]; ok {
		return client.Resource(resource), nil
	}

	return nil, fmt.Errorf("unknown type: %s", k8sType)
}
