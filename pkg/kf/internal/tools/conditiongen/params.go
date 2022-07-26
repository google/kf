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
	"bytes"
	"text/template"
)

type Params struct {
	// Package is the package the client will be generated in.
	Package string

	// ConditionPrefix is the prefix conditions will start with.
	ConditionPrefix string

	// StatusType is the type that contains the duck.Status
	StatusType string

	// Conditions is the list of conditions to generate (along with Ready).
	Conditions []string

	// LivingConditionSet is true if the conditions will be summarized by the
	// "Ready" condition. If false, they will be summarized by the "Succeeded"
	// condition. Ready should be used for continuous tasks whereas Succeeded
	// should be used for one-off jobs.
	LivingConditionSet bool
}

func (f *Params) Render() ([]byte, error) {
	buf := &bytes.Buffer{}

	templates := []*template.Template{
		conditionTemplate,
	}

	for _, tmpl := range templates {
		if err := tmpl.Execute(buf, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
