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

// go:generate ../option-builder.go

package clientgen

import (
	"bytes"
	"fmt"
	"text/template"
)

type Condition struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
	Val  string `json:"value"`
}

func (c *Condition) Definition() string {
	if c.Ref != "" {
		return c.Ref
	}

	return fmt.Sprintf("%q", c.Val)
}

func (c *Condition) ConditionName() string {
	return fmt.Sprintf("Condition%s", c.Name)
}

func (c *Condition) PredicateName() string {
	return fmt.Sprintf("%sTrue", c.ConditionName())
}

func (c *Condition) WaitForName() string {
	return fmt.Sprintf("WaitFor%s", c.PredicateName())
}

type ClientParams struct {
	// Package is the package the client will be generated in.
	Package string `json:"package"`

	// Imports is used for imports into the client.
	Imports map[string]string `json:"imports"`

	// Kubernetes holds information about the backing API.
	Kubernetes struct {
		// Group is the group of the k8s resource.
		Group string `json:"group"`
		// Kind is the kind of the resource.
		Kind string `json:"kind"`
		// Version is the version of the resource kf supports.
		Version string `json:"version"`
		// Namespaced indicates whether this object is namespaced or global.
		Namespaced bool `json:"namespaced"`
		// Plural contains the pluralizataion of kind. If blank, default of Kind+"s"
		// is assumed.
		Plural string `json:"plural"`
		// ObservedGenerationField contains the name of the field used for
		// ObservedGenerations (if the resource suuports it).
		ObservedGenerationFieldPath string `json:"observedGenerationFieldPath"`
		// ConditionField contains the name of the field used for
		// conditions.
		ConditionsFieldPath string `json:"conditionsFieldPath"`
		// Conditions holds the names of interesting conditions that will be turned
		// into predicates.
		Conditions []Condition `json:"conditions"`
	} `json:"kubernetes"`

	// CF contains information about this resource from a CF side.
	CF struct {
		// The name of the CF type.
		Name string `json:"name"`
	} `json:"cf"`

	// Type is the Go type of the resource. This MUST be imported using Imports.
	Type string `json:"type"`

	// ClientType is the Go type of the Kubernetes client. This MUST be imported using Imports.
	ClientType string `json:"clientType"`
}

// SupportsObservedGeneration returns true if the type supports
// ObservedGeneration fields.
func (f *ClientParams) SupportsObservedGeneration() bool {
	return f.Kubernetes.ObservedGenerationFieldPath != ""
}

// SupportsConditions returns true if the type supports condition fields.
func (f *ClientParams) SupportsConditions() bool {
	return f.Kubernetes.ConditionsFieldPath != ""
}

func (f *ClientParams) Render() ([]byte, error) {
	buf := &bytes.Buffer{}

	if f.Kubernetes.Plural == "" {
		f.Kubernetes.Plural = f.Kubernetes.Kind + "s"
	}

	templates := []*template.Template{
		headerTemplate,
		functionalUtilTemplate,
		clientTemplate,
	}

	for _, tmpl := range templates {
		if err := tmpl.Execute(buf, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
