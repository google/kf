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

package v1alpha1

import (
	"errors"
	"fmt"

	"knative.dev/pkg/apis"
)

// PropagateCondition copies the condition of a sub-resource (source) to a
// destination on the given manager.
// It returns true if the condition is true, otherwise false.
func PropagateCondition(manager apis.ConditionManager, destination apis.ConditionType, source *apis.Condition) bool {
	switch {
	case source == nil:
		return false
	case source.IsFalse():
		manager.MarkFalse(destination, source.Reason, source.Message)
	case source.IsTrue():
		manager.MarkTrue(destination)
	case source.IsUnknown():
		manager.MarkUnknown(destination, source.Reason, source.Message)
	}

	return source.IsTrue()
}

// SingleConditionManager provides a standard way to set conditions.
type SingleConditionManager interface {
	// MarkChildNotOwned marks the child with the given name as not being owned by
	// the app.
	MarkChildNotOwned(childName string) error

	// MarkTemplateError marks the conditoin as having an error in its template.
	MarkTemplateError(err error) error

	// MarkSourceReconciliationError marks the Source having some error during the
	// reconciliation process. Context should contain the action that failed.
	MarkReconciliationError(context string, err error) error
}

// NewSingleConditionManager sets up a manager for setting the conditions of
// a single sub-resource.
func NewSingleConditionManager(manager apis.ConditionManager, destination apis.ConditionType, childType string) SingleConditionManager {
	return &conditionImpl{
		manager:     manager,
		destination: destination,
		childType:   childType,
	}
}

type conditionImpl struct {
	manager     apis.ConditionManager
	destination apis.ConditionType
	childType   string
}

var _ SingleConditionManager = (*conditionImpl)(nil)

// MarkChildNotOwned marks the child with the given name as not being owned by
// the app.
func (ci *conditionImpl) MarkChildNotOwned(childName string) error {
	msg := fmt.Sprintf("There is an existing %s %q that we do not own.", ci.childType, childName)

	ci.manager.MarkFalse(ci.destination, "NotOwned", msg)

	return errors.New(msg)
}

// MarkTemplateError marks the conditoin as having an error in its template.
func (ci *conditionImpl) MarkTemplateError(err error) error {
	msg := fmt.Sprintf("Couldn't populate the %s template: %s", ci.childType, err)

	ci.manager.MarkFalse(ci.destination, "TemplateError", msg)

	return errors.New(msg)
}

// MarkSourceReconciliationError marks the Source having some error during the
// reconciliation process. Context should contain the action that failed.
func (ci *conditionImpl) MarkReconciliationError(action string, err error) error {
	msg := fmt.Sprintf("Error occurred while %s %s: %s", action, ci.childType, err)

	ci.manager.MarkFalse(ci.destination, "ReconciliationError", msg)

	return errors.New(msg)
}
