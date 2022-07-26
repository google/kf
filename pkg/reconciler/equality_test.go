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
)

func ExampleNewSemanticEqualityBuilder() {
	logger := zap.NewNop()

	builder := NewSemanticEqualityBuilder(logger.Sugar(), "ExampleType")
	fmt.Println("Initially equal:", builder.IsSemanticallyEqual())

	builder.Append("some.field", "a", "a")
	fmt.Println("After some.field:", builder.IsSemanticallyEqual())

	builder.Append("bad.field", "a", 1)
	fmt.Println("After bad.field:", builder.IsSemanticallyEqual())

	// Output: Initially equal: true
	// After some.field: true
	// After bad.field: false
}

func ExampleNewSemanticEqualityBuilder_logging() {
	logger := zap.NewExample()
	defer logger.Sync()

	NewSemanticEqualityBuilder(logger.Sugar(), "ExampleType").
		Append("some.field", "a", "a").
		Append("bad.field", "a", 1).
		IsSemanticallyEqual()

	// Output:
	// {"level":"debug","msg":"detected diff","type":"ExampleType","desired":"a","actual":1,"field":"bad.field"}
	//
}
