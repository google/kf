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
	"crypto/md5"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/kmeta"
)

const (
	// BackingResourceReady specifies the condition type for backing resource finishes reconciliation.
	BackingResourceReady = "BackingResourceReady"
	// Status for ChildNotOwned error.
	NotOwned = "NotOwned"
	// Status for ReconciliationError.
	ReconciliationError = "ReconciliationError"
	// Status for TemplateError.
	TemplateError = "TemplateError"
	// Status for CacheOutdated error.
	CacheOutdated = "CacheOutdated"
	// Status for Unknown error.
	Unknown = "Unknown"
)

// PropagateCondition copies the condition of a sub-resource (source) to a
// destination on the given manager.
// It returns true if the condition is true, otherwise false.
func PropagateCondition(manager apis.ConditionManager, destination apis.ConditionType, source *apis.Condition) bool {
	switch {
	case source == nil:
		manager.MarkUnknown(destination, Unknown, "status not yet reconciled")
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

	// MarkTemplateError marks the condition as having an error in its template.
	MarkTemplateError(err error) error

	// MarkReconciliationError marks the condition having some error during the
	// reconciliation process. Context should contain the action that failed.
	MarkReconciliationError(context string, err error) error

	// MarkUnknown sets the condition to Unknown with the given reason and message
	// format.
	MarkUnknown(reason, messageFormat string, messageA ...interface{})

	// MarkFalse sets the condition to False with the given reason and message
	// format.
	MarkFalse(reason, messageFormat string, messageA ...interface{})

	// IsPending returns whether the condition's state is final or in progress.
	IsPending() bool

	// MarkReconcilationPending marks the condition as still requiring reconciliation
	// This is a useful state for when ObservedGeneration doesn't match Generation
	// but progress should still be shown to the user.
	MarkReconciliationPending()

	// MarkSuccess marks the condition as being successfully reconciled.
	MarkSuccess()

	// TimeSinceTransition returns the time since last transition for the resource.
	// If the resource hasn't transitioned, the time is zero.
	TimeSinceTransition() time.Duration

	// String converts the condition to a standard human-readable representation.
	String() string

	// ErrorIfTimeout returns an error if the TimeSinceTransition() is greater than
	// the timeout, otherwise it returns nil.
	ErrorIfTimeout(timeout time.Duration) error
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

	ci.manager.MarkFalse(ci.destination, NotOwned, msg)

	return errors.New(msg)
}

// MarkTemplateError marks the conditoin as having an error in its template.
func (ci *conditionImpl) MarkTemplateError(err error) error {
	msg := fmt.Sprintf("Couldn't populate the %s template: %s", ci.childType, err)

	ci.manager.MarkFalse(ci.destination, TemplateError, msg)

	return errors.New(msg)
}

// MarkReconciliationError marks the condition having some error during the
// reconciliation process. Context should contain the action that failed.
func (ci *conditionImpl) MarkReconciliationError(action string, err error) error {
	msg := fmt.Sprintf("Error occurred while %s %s: %s", action, ci.childType, err)

	switch {
	case apierrs.IsConflict(err) || apierrs.IsAlreadyExists(err):
		// Both of these are 409 errors returned by different implementations of
		// Kubernetes controllers.
		ci.manager.MarkUnknown(ci.destination, CacheOutdated, msg)

		// In the future, additional retryable errors can be added here if
		// Kubernetes starts returning other failures.

	default:
		ci.manager.MarkFalse(ci.destination, ReconciliationError, msg)
	}

	return errors.New(msg)
}

// MarkUnknown sets the condition to Unknown with the given reason and message
// format.
func (ci *conditionImpl) MarkUnknown(reason, messageFormat string, messageA ...interface{}) {
	ci.manager.MarkUnknown(ci.destination, reason, messageFormat, messageA...)
}

// MarkError sets the condition to False with the given reason and message
// format.
func (ci *conditionImpl) MarkFalse(reason, messageFormat string, messageA ...interface{}) {
	ci.manager.MarkFalse(ci.destination, reason, messageFormat, messageA...)
}

// MarkReconcilationPending marks the condition as still requiring reconciliation
// This is a useful state for when ObservedGeneration doesn't match Generation
// but progress should still be shown to the user.
func (ci *conditionImpl) MarkReconciliationPending() {
	ci.manager.MarkUnknown(ci.destination, "Reconciling", "The resource is still reconciling.")
}

// MarkSuccess marks the source as being successfully reconciled.
func (ci *conditionImpl) MarkSuccess() {
	ci.manager.MarkTrue(ci.destination)
}

// IsPending returns whether the condition's state is final or in progress.
func (ci *conditionImpl) IsPending() bool {
	cond := ci.manager.GetCondition(ci.destination)
	return cond == nil || cond.Status == corev1.ConditionUnknown
}

// TimeSinceTransition implements SingleConditionManager.TimeSinceTransition.
func (ci *conditionImpl) TimeSinceTransition() time.Duration {
	cond := ci.manager.GetCondition(ci.destination)
	if cond == nil || cond.LastTransitionTime.Inner.IsZero() {
		return 0 * time.Second
	}

	return time.Since(cond.LastTransitionTime.Inner.Time)
}

// ErrorIfTimeout returns an error if the TimeSinceTransition() is greater than
// the timeout, otherwise it returns nil.
func (ci *conditionImpl) ErrorIfTimeout(timeout time.Duration) error {
	if ci.TimeSinceTransition() <= timeout {
		return nil
	}

	return fmt.Errorf(
		"timed out, no progress was made in %d seconds, previous status: %q",
		int64(timeout/time.Second),
		ci.String(),
	)
}

// String implements SingleConditionManager.String.
func (ci *conditionImpl) String() string {
	condition := ci.manager.GetCondition(ci.destination)
	if condition == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"condition: %s status: %q reason: %q message: %q",
		condition.Type,
		condition.Status,
		condition.Reason,
		condition.Message,
	)
}

var (
	invalidDNSCharacters = regexp.MustCompile(`[^a-z0-9-]`)
)

// GenerateName generates a name given the parts. It is a DNS valid name.
func GenerateName(parts ...string) string {
	prefix := strings.Join(parts, "-")

	// We have to ensure that what we build will always be unique. This means
	// if we were given something-a, and something_a, they should not return
	// the same value (which would happen if we simply replaced the underscore
	// with a hyphen).
	//
	// Therefore, we can't simply replace all the non-valid DNS characters.
	// Instead, if a value was replaced, then we will add the hash of the
	// original name.
	before := prefix

	prefix = strings.ToLower(prefix)

	// Remove all non-valid characters.
	prefix = invalidDNSCharacters.ReplaceAllString(prefix, "-")

	// Trim prefix and suffix of invalid leading/trailing dashes
	prefix = strings.TrimFunc(prefix, func(r rune) bool {
		return r == '-'
	})

	var suffix string
	if prefix != before {
		// Modifications were required, therefore a hash suffix is required to
		// ensure uniquness.
		suffix = fmt.Sprintf("%x", md5.Sum([]byte(before)))
	}

	return kmeta.ChildName(prefix, suffix)
}

// IsStatusFinal returns true if the Ready or Succeeded conditions are True or
// False for a Status.
func IsStatusFinal(duck duckv1beta1.Status) bool {
	// Ready conditions are used for long running tasks while Succeeded conditions
	// are used for one time tasks so it's okay to include both.
	if cond := duck.GetCondition(apis.ConditionReady); cond != nil {
		return cond.IsTrue() || cond.IsFalse()
	}

	if cond := duck.GetCondition(apis.ConditionSucceeded); cond != nil {
		return cond.IsTrue() || cond.IsFalse()
	}

	return false
}

// UnionMaps merges the keys of all the maps. Maps are merged
// in-order and the values from later params overwrite earlier
// params.
func UnionMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// MustRequirement constructs a labels.Requirement, panicking on error.
// It should only be used to create requirements which cannot fail.
func MustRequirement(key string, op selection.Operator, val string) labels.Requirement {
	r, err := labels.NewRequirement(key, op, []string{val})
	if err != nil {
		panic(err)
	}
	return *r
}

// ManagedByKfSelector selects resources which are managed by Kf.
func ManagedByKfRequirement() labels.Requirement {
	return MustRequirement(ManagedByLabel, selection.Equals, "kf")
}

// SummarizeChildConditions converts a list of conditions into a living set
// and returns the overall status as well as that set.
//
// It is intended to be used with a list of top-level conditions for child
// resources. For example, if your CRD creates a set of Pods that need to be
// handled. Each desired child should get a status regardless of whether or not
// it currently exists.
//
// If there are no conditions to summarize, the overall status is ready.
func SummarizeChildConditions(conditions []apis.Condition) (*apis.Condition, duckv1beta1.Conditions) {
	// Gather condition types into a list
	var conditionTypes []apis.ConditionType
	for _, cond := range conditions {
		conditionTypes = append(conditionTypes, cond.Type)
	}

	duckStatus := &duckv1beta1.Status{}
	manager := apis.NewLivingConditionSet(conditionTypes...).Manage(duckStatus)
	manager.InitializeConditions()

	for _, cond := range conditions {
		PropagateCondition(manager, cond.Type, &cond)
	}

	// If there are no conditions, then the overall status is ready
	if len(conditions) == 0 {
		manager.SetCondition(apis.Condition{
			Type:   apis.ConditionReady,
			Status: corev1.ConditionTrue,
		})
	}

	return manager.GetCondition(apis.ConditionReady), duckStatus.Conditions
}

// CopyMap copies the originalMap to a new map targetMap and returns it.
func CopyMap(originalMap map[string]string) map[string]string {
	if originalMap == nil {
		return nil
	}
	targetMap := make(map[string]string)
	for key, value := range originalMap {
		targetMap[key] = value
	}
	return targetMap
}
