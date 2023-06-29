// Copyright 2023 Google LLC
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

// This file was generated with functions.go, DO NOT EDIT IT.

package gentest

// Generator defined imports
import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

// User defined imports
import (
	cv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

////////////////////////////////////////////////////////////////////////////////
// Functional Utilities
////////////////////////////////////////////////////////////////////////////////

type ResourceInfo struct{}

// NewResourceInfo returns a new instance of ResourceInfo
func NewResourceInfo() *ResourceInfo {
	return &ResourceInfo{}
}

// Namespaced returns true if the type belongs in a namespace.
func (*ResourceInfo) Namespaced() bool {
	return true
}

// GroupVersionResource gets the GVR struct for the resource.
func (*ResourceInfo) GroupVersionResource(context.Context) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind(context.Context) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}
}

// FriendlyName gets the user-facing name of the resource.
func (*ResourceInfo) FriendlyName() string {
	return "OperatorConfig"
}

var (
	ConditionReady       = apis.ConditionType("Ready")
	ConditionInitialized = apis.ConditionType(corev1.PodInitialized)
)

// Predicate is a boolean function for a corev1.Pod.
type Predicate func(*corev1.Pod) bool

// Mutator is a function that changes corev1.Pod.
type Mutator func(*corev1.Pod) error

// ObservedGenerationMatchesGeneration is a predicate that returns true if the
// object's ObservedGeneration matches the genration of the object.
func ObservedGenerationMatchesGeneration(obj *corev1.Pod) bool {
	return obj.Generation == obj.Generation
}

// ExtractConditions converts the native condition types into an apis.Condition
// array with the Type, Status, Reason, and Message fields intact.
func ExtractConditions(obj *corev1.Pod) (extracted []apis.Condition) {
	for _, cond := range obj.Status.Conditions {
		// Only copy the following four fields to be compatible with
		// recommended Kubernetes fields.
		extracted = append(extracted, apis.Condition{
			Type:    apis.ConditionType(cond.Type),
			Status:  corev1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return
}

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

// Client is the interface for interacting with corev1.Pod types as OperatorConfig CF style objects.
type Client interface {
	Create(ctx context.Context, namespace string, obj *corev1.Pod) (*corev1.Pod, error)
	Transform(ctx context.Context, namespace string, name string, transformer Mutator) (*corev1.Pod, error)
	Get(ctx context.Context, namespace string, name string) (*corev1.Pod, error)
	Delete(ctx context.Context, namespace string, name string) error
	List(ctx context.Context, namespace string) ([]corev1.Pod, error)
	Upsert(ctx context.Context, namespace string, newObj *corev1.Pod, merge Merger) (*corev1.Pod, error)
	WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*corev1.Pod, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (*corev1.Pod, error)
	WaitForConditionReadyTrue(ctx context.Context, namespace string, name string, interval time.Duration) (*corev1.Pod, error)
	WaitForConditionInitializedTrue(ctx context.Context, namespace string, name string, interval time.Duration) (*corev1.Pod, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient cv1.PodsGetter
}

// Create inserts the given corev1.Pod into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(ctx context.Context, namespace string, obj *corev1.Pod) (*corev1.Pod, error) {
	return core.kclient.Pods(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls. The transform will be retried as long as the resource is in
// conflict.
func (core *coreClient) Transform(ctx context.Context, namespace string, name string, mutator Mutator) (*corev1.Pod, error) {
	for {
		obj, err := core.Get(ctx, namespace, name)
		if err != nil {
			return nil, err
		}

		if err := mutator(obj); err != nil {
			return nil, err
		}

		result, err := core.kclient.Pods(namespace).Update(ctx, obj, metav1.UpdateOptions{})

		if apierrors.IsConflict(err) {
			continue
		}
		return result, err
	}
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(ctx context.Context, namespace string, name string) (*corev1.Pod, error) {
	res, err := core.kclient.Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
// The object is only deleted once all of the objects it owns are deleted.
func (core *coreClient) Delete(ctx context.Context, namespace string, name string) error {
	foreground := metav1.DeletePropagationForeground
	if err := core.kclient.Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &foreground}); err != nil {
		return fmt.Errorf("couldn't delete the OperatorConfig with the name %q: %v", name, err)
	}

	return nil
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	res, err := core.kclient.Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list OperatorConfigs: %v", err)
	}

	return res.Items, nil
}

// Merger is a type to merge an existing value with a new one.
type Merger func(newObj, oldObj *corev1.Pod) *corev1.Pod

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
// If the update results in a conflict error, then it is retried with the new
// object. Meaning, the merge function is invoked again.
func (core *coreClient) Upsert(ctx context.Context, namespace string, newObj *corev1.Pod, merge Merger) (*corev1.Pod, error) {
	for ctx.Err() == nil {
		// kclient must be used so the error code can be validated by apierrors
		oldObj, err := core.kclient.Pods(namespace).Get(ctx, newObj.Name, metav1.GetOptions{})

		switch {
		case apierrors.IsNotFound(err):
			return core.Create(ctx, namespace, newObj)
		case err != nil:
			return nil, err
		}

		updated, err := core.kclient.Pods(namespace).Update(ctx, merge(newObj, oldObj), metav1.UpdateOptions{})
		switch {
		case apierrors.IsConflict(err):
			continue
		case err != nil:
			return nil, err
		}

		return updated, nil
	}

	return nil, ctx.Err()
}

// WaitFor polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*corev1.Pod, error) {
	return core.waitForE(ctx, namespace, name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by waitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(ctx context.Context, instance *corev1.Pod, apiErr error) (done bool, err error)

// ConditionReporter reports on changes to conditions while waiting.
type ConditionReporter func(message string)
type conditionReporterKey struct{}

// WithConditionReporter adds a callback to condition waits.
func WithConditionReporter(ctx context.Context, reporter ConditionReporter) context.Context {
	return context.WithValue(ctx, conditionReporterKey{}, reporter)
}

func maybeGetConditionReporter(ctx context.Context) ConditionReporter {
	if v := ctx.Value(conditionReporterKey{}); v != nil {
		return v.(ConditionReporter)
	}

	return nil
}

// waitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) waitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (instance *corev1.Pod, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if done, err = condition(ctx, instance, err); done {
			return
		}

		select {
		case <-tick:
			// repeat instance check
		case <-ctx.Done():
			return nil, errors.New("waiting for OperatorConfig timed out")
		}
	}
}

// ConditionDeleted is a ConditionFuncE that succeeds if the error returned by
// the cluster was a not found error.
func ConditionDeleted(ctx context.Context, _ *corev1.Pod, apiErr error) (bool, error) {
	if apiErr != nil {
		if apierrors.IsNotFound(apiErr) {
			apiErr = nil
		}

		return true, apiErr
	}

	return false, nil
}

// wrapPredicate converts a predicate to a ConditionFuncE that fails if the
// error is not nil or if the Status has a False condition.
func wrapPredicate(condition Predicate) ConditionFuncE {
	return func(ctx context.Context, obj *corev1.Pod, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		if ObservedGenerationMatchesGeneration(obj) {
			for _, cond := range ExtractConditions(obj) {
				if cond.Status == corev1.ConditionFalse {
					return true, fmt.Errorf("Reason: %q, Message: %q", cond.Reason, cond.Message)
				}
			}
		}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines waitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (instance *corev1.Pod, err error) {
	return core.waitForE(ctx, namespace, name, interval, ConditionDeleted)
}

func checkConditionTrue(ctx context.Context, obj *corev1.Pod, err error, condition apis.ConditionType) (bool, error) {
	conditionReporter := func(_ string) {}
	if reporter := maybeGetConditionReporter(ctx); reporter != nil {
		conditionReporter = reporter
	}

	if err != nil {
		return true, err
	}

	// don't propagate old statuses
	if !ObservedGenerationMatchesGeneration(obj) {
		conditionReporter("Waiting for object to be reconciled (generation out of sync)")
		return false, nil
	}

	for _, cond := range ExtractConditions(obj) {
		if cond.Type == condition {
			switch {
			case cond.IsTrue():
				return true, nil

			case cond.IsUnknown():
				conditionReporter(fmt.Sprintf("Last Transition Time: %s Reason: %q Message: %s", cond.LastTransitionTime.Inner, cond.Reason, cond.Message))
				return false, nil

			default:
				// return true and a failure assuming IsFalse and other statuses can't be
				// recovered from because they violate the K8s spec
				return true, fmt.Errorf("checking %s failed, status: %s message: %s reason: %s", cond.Type, cond.Status, cond.Message, cond.Reason)
			}
		}
	}

	conditionReporter(fmt.Sprintf("Condition %q not found", condition))

	return false, nil
}

// ConditionReadyTrue is a ConditionFuncE that waits for Condition{Ready  Ready} to
// become true and fails with an error if the condition becomes false.
func ConditionReadyTrue(ctx context.Context, obj *corev1.Pod, err error) (bool, error) {
	return checkConditionTrue(ctx, obj, err, ConditionReady)
}

// WaitForConditionReadyTrue is a utility function that combines waitForE with ConditionReadyTrue.
func (core *coreClient) WaitForConditionReadyTrue(ctx context.Context, namespace string, name string, interval time.Duration) (instance *corev1.Pod, err error) {
	return core.waitForE(ctx, namespace, name, interval, ConditionReadyTrue)
}

// ConditionInitializedTrue is a ConditionFuncE that waits for Condition{Initialized corev1.PodInitialized } to
// become true and fails with an error if the condition becomes false.
func ConditionInitializedTrue(ctx context.Context, obj *corev1.Pod, err error) (bool, error) {
	return checkConditionTrue(ctx, obj, err, ConditionInitialized)
}

// WaitForConditionInitializedTrue is a utility function that combines waitForE with ConditionInitializedTrue.
func (core *coreClient) WaitForConditionInitializedTrue(ctx context.Context, namespace string, name string, interval time.Duration) (instance *corev1.Pod, err error) {
	return core.waitForE(ctx, namespace, name, interval, ConditionInitializedTrue)
}
