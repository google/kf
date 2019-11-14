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

package describe

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Unstructured prints information about an unstructured Kubernetes
// object.
func Unstructured(w io.Writer, resource *unstructured.Unstructured) error {
	return UnstructuredMap(w, resource.UnstructuredContent())
}

// UnstructuredMap writes information about a JSON style unstructured
// object.
func UnstructuredMap(w io.Writer, obj map[string]interface{}) error {
	return TabbedWriter(w, func(w io.Writer) error {
		var keys []string
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, rawKey := range keys {
			rawVal := obj[rawKey]
			key := JSONKeyToTitleCase(rawKey)

			switch val := rawVal.(type) {
			case map[string]interface{}:
				if err := SectionWriter(w, key, func(w io.Writer) error {
					return UnstructuredMap(w, val)
				}); err != nil {
					return err
				}
			case []interface{}:
				if err := SectionWriter(w, key, func(w io.Writer) error {
					return UnstructuredArray(w, val)
				}); err != nil {
					return err
				}
			default:
				if _, err := fmt.Fprintf(w, "%s:\t%v\n", key, val); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// UnstructuredArray formats and writes an array to the output.
func UnstructuredArray(w io.Writer, anArray []interface{}) error {
	for _, rawVal := range anArray {
		switch val := rawVal.(type) {
		case map[string]interface{}:
			if err := UnstructuredMap(w, val); err != nil {
				return err
			}
		case []interface{}:
			if err := UnstructuredArray(w, val); err != nil {
				return err
			}
		default:
			if _, err := fmt.Fprintf(w, "%v\n", val); err != nil {
				return err
			}
		}
	}

	return nil
}

// JSONKeyToTitleCase converts a JSON key to a human friendly casing.
func JSONKeyToTitleCase(name string) string {
	out := ""

	var last rune
	for _, curr := range name {
		insertBreak := false
		switch {
		case unicode.IsLower(curr):
			// lower only indicates a new word if from a digit
			// e.g. CSS3andJS
			insertBreak = unicode.IsNumber(last)
		case unicode.IsUpper(curr):
			// upper case only gets broken if last was not upper
			insertBreak = !unicode.IsUpper(last)
		case unicode.IsDigit(curr):
			// digits get broken from lower but not from other digits or upper
			// e.g. HTML5
			insertBreak = unicode.IsLower(last)
		default:
			return name
		}

		// First character is always capitalized and doesn't
		// get a break.
		if last == rune(0) {
			insertBreak = false
			curr = unicode.ToUpper(curr)
		}

		if insertBreak {
			out += " "
			curr = unicode.ToUpper(curr)
		}
		out += string(curr)
		last = curr
	}

	// Replace acronyms for nicer output
	parts := strings.Split(out, " ")
	for i, part := range parts {
		for _, acronym := range []string{"API", "URL", "UID", "OSB", "GUID"} {
			if strings.ToUpper(part) == acronym {
				parts[i] = acronym
			}
		}
	}

	return strings.Join(parts, " ")
}
