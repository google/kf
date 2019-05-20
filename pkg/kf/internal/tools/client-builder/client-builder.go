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

// +build ignore

package main

import "text/template"

type Client struct {
	Package string `yaml:"package"`

	Client string `yaml:"client"`

	// Imports is a map of <lib>:<name> pairs
	Imports map[string]string `yaml:"imports"`

	Kubernetes struct {
		GoType  string `yaml:"gotype"`
		Group   string `yaml:"group"`
		Version string `yaml:"version"`
		Kind    string `yaml:"kind"`
	} `yaml:"kubernetes"`
}

var fileTemplate = template.Must(template.New("").Funcs(template.FuncMap{}).Parse(`
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

// This file was generated with client-builder.go, DO NOT EDIT IT.

package {{.Package}}
{{ if .Imports }}
import ({{ range $lib, $name := .Imports }}{{ printf "\n\t%s %q" $name $lib }}{{ end }}{{printf "\n"}})
{{ end }}

// {{.Kubernetes.Kind}}Predicate is a predicate function used for filtering.
type {{.Kubernetes.Kind}}Predicate func(*{{.Kubernetes.GoType}}) bool

// {{.Kubernetes.Kind}}Validator returns an error if the provided {{.Kubernetes.Kind}} is invalid.
type {{.Kubernetes.Kind}}Validator func(*{{.Kubernetes.GoType}}) error


type internalClient struct {

}


// Create creates a Kubernetes {{.Kubernetes.Kind}} with the given values.
func (c *{{.Client}}) Create(name string, options ...CreateOption) error {
	config := CreateOptionDefaults().Extend(options).toConfig()

	secret := &corev1.Secret{StringData: config.StringData, Data: config.Data}

	secret.Name = name
	secret.Kind = "Secret"
	secret.Namespace = config.Namespace
	secret.APIVersion = "v1"
	secret.Labels = config.Labels

	if _, err := c.kclient.CoreV1().Secrets(config.Namespace).Create(secret); err != nil {
		return err
	}

	return nil
}

`))
