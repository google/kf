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

package sanitizer_test

import (
	"bytes"
	"fmt"
	"kf-operator/pkg/sanitizer"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

func TestFuncMap(t *testing.T) {
	testCases := []struct {
		name         string
		template     string
		replacements map[string]interface{}
		want         string
	}{
		{
			name:     "sanitize quotes strings",
			template: "test: {{.InputVariable | sanitize}}",
			replacements: map[string]interface{}{
				"InputVariable": "Output",
			},
			want: "test: \"Output\"",
		},
		{
			name:     "sanitize prevents newline injection",
			template: "test: {{.InputVariable | sanitize}}",
			replacements: map[string]interface{}{
				"InputVariable": "hello \n---Hack",
			},
			want: "test: \"hello \\n---Hack\"",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			tmpl, err := makeTemplate(test.template)
			if err != nil {
				t.Fatalf("Error parsing template: %v", test.template)
			}

			got, err := templateOutput(tmpl, test.replacements)
			if err != nil {
				t.Fatalf("Error executing template: %v", test.template)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ExecuteTemplate (-want, +got) = %v", diff)
			}
		})
	}
}

func templateOutput(tmpl *template.Template, replacements map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, replacements)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return buf.String(), nil
}

func makeTemplate(templateString string) (*template.Template, error) {
	tmpl, err := template.New("test").Funcs(sanitizer.FuncMap()).Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("error creating template, %v", err)
	}
	goTmplOpts := []string{"missingkey=error"}

	for _, goTmplOpt := range goTmplOpts {
		tmpl.Option(goTmplOpt)
	}
	return tmpl, nil
}
