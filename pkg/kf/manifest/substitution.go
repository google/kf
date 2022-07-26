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

package manifest

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

// substitutionPattern looks for variables of the form ((varname))
// the inside of the brackes can be any character except for an
// open paren to prevent nested replacements.
const substitutionPattern = `\(\([^\(]*?\)\)`

var variableRegex = regexp.MustCompile(substitutionPattern)
var wholeStringRegex = regexp.MustCompile(`^` + substitutionPattern + `$`)

// ApplySubstitution applies variable substitution to a YAML or JSON byte array.
// Variables in the source are specified using ((VARNAME)) syntax.
func ApplySubstitution(source []byte, variables map[string]interface{}) ([]byte, error) {
	var obj map[string]interface{}
	if err := yaml.Unmarshal(source, &obj); err != nil {
		return nil, err
	}

	out, err := substitute(obj, variables)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(out)
}

func substitute(obj interface{}, variables map[string]interface{}) (interface{}, error) {
	switch typed := obj.(type) {
	case map[string]interface{}: // maps, substitute key and value
		out := make(map[string]interface{})
		for k, v := range typed {
			nk, kerr := substitute(k, variables)
			if kerr != nil {
				return nil, kerr
			}
			nv, verr := substitute(v, variables)
			if verr != nil {
				return nil, verr
			}

			out[fmt.Sprintf("%v", nk)] = nv
		}
		return out, nil

	case []interface{}: // arrays, substitute each item
		for i := range typed {
			sub, err := substitute(typed[i], variables)
			if err != nil {
				return nil, err
			}

			typed[i] = sub
		}
		return typed, nil

	case string: // strings, substitute text within
		// if we match the whole string, substitute fully including the type
		if wholeStringRegex.MatchString(typed) {
			return findSubstitution(typed, variables)
		}

		var replaceErr error
		out := variableRegex.ReplaceAllStringFunc(typed, func(key string) (sub string) {
			tmp, tmpErr := findSubstitution(key, variables)
			if tmpErr != nil {
				replaceErr = tmpErr
			}
			return fmt.Sprintf("%v", tmp)
		})

		if replaceErr != nil {
			return nil, replaceErr
		}

		return out, nil

	default: // other types can't be substituted
		return typed, nil
	}
}

func findSubstitution(key string, variables map[string]interface{}) (interface{}, error) {
	key = strings.TrimPrefix(key, "((")
	key = strings.TrimSuffix(key, "))")

	found, ok := variables[key]
	if !ok {
		return nil, fmt.Errorf("no variable found for key: ((%s))", key)
	}

	return found, nil
}
