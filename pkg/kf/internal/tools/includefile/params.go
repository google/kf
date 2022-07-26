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

package includefile

import (
	"bytes"
	"text/template"
)

type Params struct {
	// Package is the package the client will be generated in.
	Package string

	// Variable is the name of the variable that will include the file.
	Variable string

	// File is the name of the file that was included
	File string

	// Contents is the contents of the file to include.
	Contents []byte
}

func (f *Params) Render() ([]byte, error) {
	buf := &bytes.Buffer{}

	templates := []*template.Template{
		variableTemplate,
	}

	for _, tmpl := range templates {
		if err := tmpl.Execute(buf, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
