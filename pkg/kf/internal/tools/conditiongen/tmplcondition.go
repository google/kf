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

package conditiongen

import (
	"text/template"

	"github.com/google/kf/v2/pkg/kf/internal/tools/generator"
)

var conditionTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{genlicense}}

{{gennotice "conditiongen/generator.go"}}

package {{.Package}}

import (
  "knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

{{ $params := . }}

// ConditionType represents a Service condition value
const (
	{{ if .LivingConditionSet }}
	// {{.ConditionPrefix}}ConditionReady is set when the CRD is configured and is usable.
	{{.ConditionPrefix}}ConditionReady = apis.ConditionReady
	{{ else }}
	// {{.ConditionPrefix}}ConditionSucceeded is set when the CRD is completed.
	{{.ConditionPrefix}}ConditionSucceeded = apis.ConditionSucceeded
	{{ end }}

	{{ range .Conditions }}
	// {{ $params.ConditionPrefix}}Condition{{.}}Ready is set when the child
	// resource(s) {{.}} is/are ready.
	{{ $params.ConditionPrefix}}Condition{{.}}Ready apis.ConditionType = "{{.}}Ready"
	{{ end }}
)

func (status *{{$params.StatusType}}) manage() apis.ConditionManager {
	return apis.New{{ if .LivingConditionSet }}Living{{ else }}Batch{{ end }}ConditionSet(
    {{ range .Conditions }}{{ $params.ConditionPrefix}}Condition{{.}}Ready,
  	{{ end }}
	).Manage(status)
}

{{ if .LivingConditionSet }}
// IsReady looks at the conditions to see if they are happy.
func (status *{{$params.StatusType}}) IsReady() bool {
	return status.manage().IsHappy()
}

// PropagateTerminatingStatus updates the ready status of the resource to False
// if the resource received a delete request.
func (status *{{$params.StatusType}}) PropagateTerminatingStatus() {
	status.manage().MarkFalse({{.ConditionPrefix}}ConditionReady, "Terminating", "resource is terminating")
}
{{ else }}
// Succeeded returns if the type successfully completed.
func (status *{{$params.StatusType}}) Succeeded() bool {
	return status.manage().IsHappy()
}
{{ end }}

// GetCondition returns the condition by name.
func (status *{{$params.StatusType}}) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *{{$params.StatusType}}) InitializeConditions() {
	status.manage().InitializeConditions()
}

{{ range .Conditions }}
// {{.}}Condition gets a manager for the state of the child resource.
func (status *{{$params.StatusType}}) {{.}}Condition() SingleConditionManager {
  return NewSingleConditionManager(status.manage(), {{ $params.ConditionPrefix}}Condition{{.}}Ready, "{{.}}")
}
{{ end }}

func (status *{{$params.StatusType}}) duck() *duckv1beta1.Status {
	return &status.Status
}
`))
