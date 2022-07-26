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

package fakelister

import (
	"bytes"
	"text/template"
)

type Params struct {
	// Package is the package the client will be generated in.
	Package string

	// ObjectType is the type that the lister lists.
	ObjectType string

	// ObjectPackage is the package of the object that the lister lists.
	ObjectPackage string

	// ListerPackage is the package of the lister.
	ListerPackage string

	// Namespaced is if the object is a namespaced object.
	Namespaced bool
}

// Render writes the rendered template.
func (p *Params) Render() ([]byte, error) {
	buf := &bytes.Buffer{}

	templates := []*template.Template{
		listerTemplate,
	}

	for _, tmpl := range templates {
		if err := tmpl.Execute(buf, p); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
