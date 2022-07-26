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

package reconciler

import (
	"fmt"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ExampleNewUnstructuredSemanticEqualityBuilder() {
	logger := zap.NewNop()

	a := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 1,
				"slice": []int{
					0, 1,
				},
			},
		},
	}
	b := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 2,
			},
		},
	}

	builder := NewUnstructuredSemanticEqualityBuilder(logger.Sugar(), "ExampleType")
	fmt.Println("Initially equal:", builder.IsSemanticallyEqual())

	builder.Append("spec.a", a, b)
	fmt.Println("After spec.a:", builder.IsSemanticallyEqual())

	builder.Append("spec.slice.bad", a, b)
	fmt.Println("After spec.slice.bad:", builder.IsSemanticallyEqual())

	// Output: Initially equal: true
	// After spec.a: true
	// After spec.slice.bad: false
}

func ExampleNewUnstructuredSemanticEqualityBuilder_logging() {
	logger := zap.NewExample()
	defer logger.Sync()

	a := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 1,
			},
		},
	}
	b := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 2,
			},
		},
	}

	NewUnstructuredSemanticEqualityBuilder(logger.Sugar(), "ExampleType").
		Append("spec.a", a, b).
		Append("spec.b", a, b).
		IsSemanticallyEqual()

	// Output:
	// {"level":"debug","msg":"detected diff","type":"ExampleType","desired":1,"actual":2,"field":"spec.b"}
	//
}

func ExampleNewUnstructuredSemanticEqualityBuilder_transform() {
	logger := zap.NewNop()

	desired := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 1,
			},
		},
	}
	actual := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{
				"a": 0,
				"b": 2,
			},
		},
	}

	builder := NewUnstructuredSemanticEqualityBuilder(logger.Sugar(), "ExampleType")
	fmt.Println("Initially equal:", builder.IsSemanticallyEqual())

	builder.Append("spec.a", desired, actual)
	fmt.Println("After spec.a:", builder.IsSemanticallyEqual())

	builder.Append("spec.b", desired, actual)
	fmt.Println("After spec.b:", builder.IsSemanticallyEqual())

	builder.Transform(actual)
	builder2 := NewUnstructuredSemanticEqualityBuilder(logger.Sugar(), "ExampleType")
	builder2.Append("spec.b", desired, actual)
	fmt.Println("After transform spec.b:", builder2.IsSemanticallyEqual())

	// Output: Initially equal: true
	// After spec.a: true
	// After spec.b: false
	// After transform spec.b: true
}
