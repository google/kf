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

package clientgen

import (
	"text/template"

	"github.com/google/kf/v2/pkg/kf/internal/tools/generator"
)

var functionalUtilTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`

////////////////////////////////////////////////////////////////////////////////
// Functional Utilities
////////////////////////////////////////////////////////////////////////////////

type ResourceInfo struct{}

// NewResourceInfo returns a new instance of ResourceInfo
func NewResourceInfo() *ResourceInfo {
	return &ResourceInfo{}
}

// Namespaced returns true if the type belongs in a namespace.
func (*ResourceInfo) Namespaced() bool {
	return {{.Kubernetes.Namespaced}}
}

// GroupVersionResource gets the GVR struct for the resource.
func (*ResourceInfo) GroupVersionResource(context.Context) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "{{ .Kubernetes.Group }}",
		Version:  "{{ .Kubernetes.Version }}",
		Resource: "{{ .Kubernetes.Plural | lower }}",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind(context.Context) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "{{ .Kubernetes.Group }}",
		Version: "{{ .Kubernetes.Version }}",
		Kind:    "{{ .Kubernetes.Kind }}",
	}
}

// FriendlyName gets the user-facing name of the resource.
func (*ResourceInfo) FriendlyName() string {
	return "{{.CF.Name}}"
}

{{ if and .SupportsConditions .Kubernetes.Conditions }}
var (
	{{ range .Kubernetes.Conditions }}
	{{.ConditionName}} = apis.ConditionType({{.Definition}}){{ end }}
)
{{ end }}

// Predicate is a boolean function for a {{.Type}}.
type Predicate func(*{{.Type}}) bool

// Mutator is a function that changes {{.Type}}.
type Mutator func(*{{.Type}}) error

{{ if .SupportsObservedGeneration }}
// ObservedGenerationMatchesGeneration is a predicate that returns true if the
// object's ObservedGeneration matches the genration of the object.
func ObservedGenerationMatchesGeneration(obj *{{.Type}}) bool {
	return obj.Generation == obj.{{.Kubernetes.ObservedGenerationFieldPath}}
}
{{ end }}

{{ if .SupportsConditions }}
// ExtractConditions converts the native condition types into an apis.Condition
// array with the Type, Status, Reason, and Message fields intact.
func ExtractConditions(obj *{{.Type}}) (extracted []apis.Condition) {
	for _, cond := range obj.{{.Kubernetes.ConditionsFieldPath}} {
		// Only copy the following four fields to be compatible with
		// recommended Kubernetes fields.
		extracted = append(extracted, apis.Condition{
			Type:    apis.ConditionType(cond.Type),
			Status:  corev1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return
}
{{ end }}
`))
