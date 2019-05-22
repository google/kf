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
	"text/template"
)

type ClientParams struct {
	// Package is the package the client will be generated in.
	Package string `yaml:"package"`

	// Imports is used for imports into the client.
	Imports map[string]string `yaml:"imports"`

	// Kubernetes holds information about the backing API.
	Kubernetes struct {
		// Kind is the kind of the resource.
		Kind string `yaml:"kind"`
		// Version is the version of the resource kf supports.
		Version string `yaml:"version"`
		// Namespaced indicates whether this object is namespaced or global.
		Namespaced bool `yaml:"namespaced"`
		// Plural contains the pluralizataion of kind. If blank, default of Kind+"s"
		// is assumed.
		Plural string `yaml:"plural"`
	} `yaml:"kubernetes"`

	// CF contains information about this resource from a CF side.
	CF struct {
		// The name of the CF type.
		Name string `yaml:"name"`
	} `yaml:"cf"`

	// Type is the go type of the resource. This MUST be imported using Imports.
	Type string `yaml:"type"`

	// ClientType is the go type of the Kubernetes client. This MUST be imported using Imports.
	ClientType string `yaml:"clienttype"`
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
