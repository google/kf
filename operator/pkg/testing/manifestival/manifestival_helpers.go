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

package manifestival

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Object is the type which encompasses most kubernetes types
// https://github.com/kubernetes-sigs/controller-runtime/pull/898
type Object interface {
	metav1.Object
	runtime.Object
}

// UnstructuredOption enables further configuration of a Unstructured.
type UnstructuredOption func(*unstructured.Unstructured)

var (
	isLastAppliedConfigPath = regexp.MustCompile(fmt.Sprintf(".*[Aa]nnotations.*%s.*", v1.LastAppliedConfigAnnotation))

	// IgnoreLastAppliedConfig is a cmp.Option that ignores LastAppliedConfig.
	IgnoreLastAppliedConfig = cmp.FilterPath(func(p cmp.Path) bool {
		return isLastAppliedConfigPath.MatchString(p.GoString())
	}, cmp.Ignore())

	// CommonOptions are common cmp.Options.
	CommonOptions = []cmp.Option{
		cmpopts.IgnoreMapEntries(func(k, v interface{}) bool {
			if v == nil {
				return true
			}
			if list, converted := v.([]interface{}); converted {
				return len(list) == 0
			}
			if dict, converted := v.(map[string]interface{}); converted {
				return len(dict) == 0
			}
			return false
		}), IgnoreLastAppliedConfig}
)

// Create is used in TableTest to assert creates.
func Create(obj Object, uo ...UnstructuredOption) *unstructured.Unstructured {
	return ToUnstructured(obj, uo...)
}

// SetManifestivalAnnotation updates annotation of obj.
func SetManifestivalAnnotation(obj Object) {
	setAnnotations(obj, "manifestival", "new")
}

// ToManifestivalOwnedObjs converts Objects to Objects owned by manifestival.
func ToManifestivalOwnedObjs(t *testing.T, in []Object) []Object {
	asRuntimeObj := make([]Object, len(in))
	for i, obj := range in {
		obj, ok := obj.DeepCopyObject().(Object)
		if !ok {
			t.Fatalf("Failed to copy to a new Object")
		}
		SetManifestivalAnnotation(obj)
		SetLastApplied(obj)
		asRuntimeObj[i] = obj
	}
	return asRuntimeObj
}

// SetLastApplied sets annotation LastAppliedConfig of obj.
func SetLastApplied(obj Object, uo ...UnstructuredOption) {
	// If we don't do this, we risk runaway expansion if an object is reused.
	ClearLastApplied(obj)
	json, err := ToUnstructured(obj, uo...).MarshalJSON()
	if err != nil {
		panic(err)
	}
	setAnnotations(obj, v1.LastAppliedConfigAnnotation, string(json))
}

// ClearLastApplied clears annotation LastAppliedConfig of obj.
func ClearLastApplied(obj Object) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return
	}
	delete(annotations, v1.LastAppliedConfigAnnotation)
}

// ClearTimestamp clears the creation time, which is generally defaulted in tests.
func ClearTimestamp(u *unstructured.Unstructured) {
	u.SetCreationTimestamp(metav1.Time{})
}

// DeleteStatus clears the status
func DeleteStatus(u *unstructured.Unstructured) {
	delete(u.Object, "status")
}

// ToUnstructured creates Unstructured and then applies UnstructuredOptions to it.
func ToUnstructured(obj runtime.Object, uo ...UnstructuredOption) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	if err := scheme.Scheme.Convert(obj, u, nil); err != nil {
		panic(err)
	}

	for _, opt := range uo {
		opt(u)
	}

	return u
}

func setAnnotations(obj metav1.Object, key, value string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		obj.SetAnnotations(make(map[string]string))
	}
	obj.GetAnnotations()[key] = value
}
