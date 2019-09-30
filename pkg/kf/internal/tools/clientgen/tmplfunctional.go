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

	"github.com/google/kf/pkg/kf/internal/tools/generator"
)

var functionalUtilTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`

////////////////////////////////////////////////////////////////////////////////
// Functional Utilities
////////////////////////////////////////////////////////////////////////////////

const (
	// Kind contains the kind for the backing Kubernetes API.
	Kind = "{{.Kubernetes.Kind}}"

	// APIVersion contains the version for the backing Kubernetes API.
	APIVersion = "{{.Kubernetes.Version}}"
)

{{ if .SupportsConditions }}
var (
	{{ range .Kubernetes.Conditions }}
	{{.ConditionName}} = apis.ConditionType({{.Definition}}){{ end }}
)
{{ end }}

// Predicate is a boolean function for a {{.Type}}.
type Predicate func(*{{.Type}}) bool

// Mutator is a function that changes {{.Type}}.
type Mutator func(*{{.Type}}) error

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator Mutator) Mutator {
	return func(mutable *{{.Type}}) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two {{.Type}}s and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *{{.Type}}) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "{{.CF.Name}} Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}

// List represents a collection of {{.Type}}.
type List []{{.Type}}

// Filter returns a new list items for which the predicates fails removed.
func (list List) Filter(filter Predicate) (out List) {
	for _, v := range list {
		if filter(&v) {
			out = append(out, v)
		}
	}

	return
}


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
		// recommended Kuberntes fields.
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
