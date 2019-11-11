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
	"context"
	"errors"
	"fmt"
	"hash/crc64"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	cv1alpha3 "knative.dev/pkg/client/clientset/versioned/typed/istio/v1alpha3"
)

// PropagateCondition copies the condition of a sub-resource (source) to a
// destination on the given manager.
// It returns true if the condition is true, otherwise false.
func PropagateCondition(manager apis.ConditionManager, destination apis.ConditionType, source *apis.Condition) bool {
	switch {
	case source == nil:
		manager.MarkUnknown(destination, "Unknown", "source status is nil")
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

	// MarkSourceReconciliationError marks the Source having some error during the
	// reconciliation process. Context should contain the action that failed.
	MarkReconciliationError(context string, err error) error

	// IsPending returns whether the condition's state is final or in progress.
	IsPending() bool
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

	switch {
	case apierrs.IsConflict(err) || apierrs.IsAlreadyExists(err):
		// Both of these are 409 errors returned by different implementations of
		// Kubernetes controllers.
		ci.manager.MarkUnknown(ci.destination, "CacheOutdated", msg)

		// In the future, additional retryable errors can be added here if
		// Kubernetes starts returning other failures.

	default:
		ci.manager.MarkFalse(ci.destination, "ReconciliationError", msg)
	}

	return errors.New(msg)
}

// IsPending returns whether the condition's state is final or in progress.
func (ci *conditionImpl) IsPending() bool {
	cond := ci.manager.GetCondition(ci.destination)
	return cond == nil || cond.Status == corev1.ConditionUnknown
}

var (
	nonAlphaNumeric    = regexp.MustCompile(`[^a-z0-9]`)
	validDNSCharacters = regexp.MustCompile(`[^a-z0-9-_]`)
)

// GenerateName generates a name given the parts. It is a DNS valid name.
func GenerateName(parts ...string) string {
	prefix := strings.Join(parts, "-")

	// Base 36 uses the characters 0-9a-z. The maximum number of chars it
	// requires is 13 for a Uin64.
	checksum := strconv.FormatUint(
		crc64.Checksum(
			[]byte(strings.Join(parts, "")),
			crc64.MakeTable(crc64.ECMA),
		),
		36)

	prefix = strings.ToLower(prefix)

	// Remove all non-alphanumeric characters.
	prefix = validDNSCharacters.ReplaceAllString(prefix, "-")

	// First char must be alphanumeric.
	for len(prefix) > 0 && nonAlphaNumeric.Match([]byte{prefix[0]}) {
		prefix = prefix[1:]
	}

	// Subtract an extra 1 for the hyphen between the prefix and checksum.
	maxPrefixLen := 64 - 1 - len(checksum)

	if len(prefix) > maxPrefixLen {
		prefix = prefix[:maxPrefixLen]
	}

	if prefix == "" {
		return checksum
	}

	return fmt.Sprintf("%s-%s", prefix, checksum)
}

type istioClientKey struct{}

func SetupIstioClient(ctx context.Context, istioClient cv1alpha3.VirtualServicesGetter) context.Context {
	return context.WithValue(ctx, istioClientKey{}, istioClient)
}

func IstioClientFromContext(ctx context.Context) cv1alpha3.VirtualServicesGetter {
	return ctx.Value(istioClientKey{}).(cv1alpha3.VirtualServicesGetter)
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

// UnionMaps is similar to
// github.com/google/kf/third_party/knative-serving/pkg/resources however it
// takes multiple maps instead of only 2.
func UnionMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
