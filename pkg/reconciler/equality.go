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
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/equality"
)

// SemanticEqualityBuilder assists in building deep equality checks for
// reconcilers.
type SemanticEqualityBuilder struct {
	logger        *zap.SugaredLogger
	semanticEqual bool
}

// NewSemanticEqualityBuilder creates a SemanticEqualityBuilder.
func NewSemanticEqualityBuilder(logger *zap.SugaredLogger, typeName string) *SemanticEqualityBuilder {
	return &SemanticEqualityBuilder{
		logger:        logger.With(zap.String("type", typeName)),
		semanticEqual: true,
	}
}

// Append adds a semantic deep equality check of the given field to the builder.
func (s *SemanticEqualityBuilder) Append(field string, desired, actual interface{}) *SemanticEqualityBuilder {
	p1, ok1 := desired.(proto.Message)
	p2, ok2 := actual.(proto.Message)
	if ok1 && ok2 {
		if proto.Equal(p1, p2) {
			return s
		}
	}

	// DeepEqual doesn't work on proto.Message because of potentially unexported fields
	if !ok1 && !ok2 {
		if equality.Semantic.DeepEqual(desired, actual) {
			return s
		}
	}

	s.semanticEqual = false

	// We don't use Knative's diff here because the underlying diff it calls
	// says not to rely on the output format and slightly modifies outputs so
	// repeated tests against it fail.

	s.logger.With(
		zap.Reflect("desired", desired),
		zap.Reflect("actual", actual),
		zap.String("field", field),
	).Debug("detected diff")
	return s
}

// IsSemanticallyEqual returns true if the fields that have been checked are
// all semantically equal.
func (s *SemanticEqualityBuilder) IsSemanticallyEqual() bool {
	return s.semanticEqual
}
