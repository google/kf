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

package transformations

import (
	"context"

	"github.com/hashicorp/go-multierror"
	mf "github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// AddMissingProtocol decorates Deployment-like and Service-like (as
// defined by 'has the same nested fields') objects with a protocol
// field set to TCP to bypass
// https://github.com/kubernetes-sigs/structured-merge-diff/issues/130
func AddMissingProtocol(context.Context) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		var result *multierror.Error
		if containers, found, _ := unstructured.NestedSlice(u.Object, "spec", "template", "spec", "containers"); found {
			nc := make([]interface{}, 0)
			for _, c := range containers {
				cm, y := c.(map[string]interface{})
				if !y {
					nc = append(nc, cm)
					continue
				}
				if ports, found, _ := unstructured.NestedSlice(cm, "ports"); found {
					np := make([]interface{}, 0)
					for _, p := range ports {
						if pm, y := p.(map[string]interface{}); y {
							if pm["protocol"] == nil || pm["protocol"] == "" {
								pm["protocol"] = "TCP"
							}
						}
						np = append(np, p)
					}

					result = multierror.Append(result, unstructured.SetNestedSlice(cm, np, "ports"))
				}
				nc = append(nc, cm)
			}
			result = multierror.Append(result, unstructured.SetNestedSlice(u.Object, nc, "spec", "template", "spec", "containers"))
		}
		if ports, found, _ := unstructured.NestedSlice(u.Object, "spec", "ports"); found {
			np := make([]interface{}, 0)
			for _, p := range ports {
				if ports, y := p.(map[string]interface{}); y {
					if ports["protocol"] == nil || ports["protocol"] == "" {
						ports["protocol"] = "TCP"
					}
				}
				np = append(np, p)
			}
			result = multierror.Append(result, unstructured.SetNestedSlice(u.Object, np, "spec", "ports"))
		}
		return result.ErrorOrNil()
	}
}
