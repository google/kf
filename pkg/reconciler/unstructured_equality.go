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
	"strings"

	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UnstructuredSemanticEqualityBuilder assists in building deep equality
// checks for reconcilers.
type UnstructuredSemanticEqualityBuilder struct {
	logger        *zap.SugaredLogger
	semanticEqual bool
	transformers  []func(u *unstructured.Unstructured)
}

// NewUnstructuredSemanticEqualityBuilder creates a
// UnstructuredSemanticEqualityBuilder.
func NewUnstructuredSemanticEqualityBuilder(logger *zap.SugaredLogger, typeName string) *UnstructuredSemanticEqualityBuilder {
	return &UnstructuredSemanticEqualityBuilder{
		logger:        logger.With(zap.String("type", typeName)),
		semanticEqual: true,
	}
}

// Append adds a semantic deep equality check of the given field to the builder.
func (s *UnstructuredSemanticEqualityBuilder) Append(
	fieldStr string,
	desired *unstructured.Unstructured,
	actual *unstructured.Unstructured,
) *UnstructuredSemanticEqualityBuilder {
	fields := strings.Split(fieldStr, ".")

	desiredObj, _, err := unstructured.NestedFieldNoCopy(
		desired.Object,
		fields...,
	)
	if err != nil {
		// Errors will imply we are NOT equal, and will be logged.
		s.logger.Warn(err.Error())
		s.semanticEqual = false
		return s
	}
	s.transformers = append(s.transformers, func(u *unstructured.Unstructured) {
		dynamicutils.SetNestedFieldNoCopy(u.Object, desiredObj, fields...)
	})

	actualObj, _, err := unstructured.NestedFieldNoCopy(
		actual.Object,
		fields...,
	)
	if err != nil {
		// Errors will imply we are NOT equal, and will be logged.
		s.logger.Warn(err.Error())
		s.semanticEqual = false
		return s
	}

	if equality.Semantic.DeepEqual(desiredObj, actualObj) {
		return s
	}

	s.semanticEqual = false

	// We don't use Knative's diff here because the underlying diff it calls
	// says not to rely on the output format and slightly modifies outputs so
	// repeated tests against it fail.

	s.logger.With(
		zap.Reflect("desired", desiredObj),
		zap.Reflect("actual", actualObj),
		zap.String("field", strings.Join(fields, ".")),
	).Debug("detected diff")
	return s
}

// IsSemanticallyEqual returns true if the fields that have been checked are
// all semantically equal.
func (s *UnstructuredSemanticEqualityBuilder) IsSemanticallyEqual() bool {
	return s.semanticEqual
}

// Transform will perform all the transformations on the Unstructured object
// to make the actual object match the desired one.
func (s *UnstructuredSemanticEqualityBuilder) Transform(existing *unstructured.Unstructured) {
	for _, transform := range s.transformers {
		transform(existing)
	}
}
