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

	mf "github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// AddAnnotation add annotation to specified resource
func AddAnnotation(context context.Context, kind string, name string, annotations map[string]string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		if len(kind) == 0 && len(name) == 0 {
			return nil
		}

		if k := u.GetKind(); len(kind) == 0 || k == kind {
			if n := u.GetName(); len(name) == 0 || n == name {
				u.SetAnnotations(annotations)
			}
		}
		return nil
	}
}
