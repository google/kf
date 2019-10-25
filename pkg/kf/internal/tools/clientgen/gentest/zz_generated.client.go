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

// This file was generated with functions.go, DO NOT EDIT IT.

package gentest

// Generator defined imports
import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmp"
)

// User defined imports
import (
	v1 "k8s.io/api/core/v1"
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
func (*ResourceInfo) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind() schema.GroupVersionKind {
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
	ConditionInitialized = apis.ConditionType(v1.PodInitialized)
)

// Predicate is a boolean function for a v1.Pod.
type Predicate func(*v1.Pod) bool

// Mutator is a function that changes v1.Pod.
type Mutator func(*v1.Pod) error

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator Mutator) Mutator {
	return func(mutable *v1.Pod) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two v1.Pods and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1.Pod) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "OperatorConfig Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}

// List represents a collection of v1.Pod.
type List []v1.Pod

// Filter returns a new list items for which the predicates fails removed.
func (list List) Filter(filter Predicate) (out List) {
	for _, v := range list {
		if filter(&v) {
			out = append(out, v)
		}
	}

	return
}

// ObservedGenerationMatchesGeneration is a predicate that returns true if the
// object's ObservedGeneration matches the genration of the object.
func ObservedGenerationMatchesGeneration(obj *v1.Pod) bool {
	return obj.Generation == obj.Generation
}

// ExtractConditions converts the native condition types into an apis.Condition
// array with the Type, Status, Reason, and Message fields intact.
func ExtractConditions(obj *v1.Pod) (extracted []apis.Condition) {
	for _, cond := range obj.Status.Conditions {
		// Only copy the following four fields to be compatible with
		// recommended Kuberntes fields.
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

// Client is the interface for interacting with v1.Pod types as OperatorConfig CF style objects.
type Client interface {
	Create(namespace string, obj *v1.Pod, opts ...CreateOption) (*v1.Pod, error)
	Update(namespace string, obj *v1.Pod, opts ...UpdateOption) (*v1.Pod, error)
	Transform(namespace string, name string, transformer Mutator) (*v1.Pod, error)
	Get(namespace string, name string, opts ...GetOption) (*v1.Pod, error)
	Delete(namespace string, name string, opts ...DeleteOption) error
	List(namespace string, opts ...ListOption) ([]v1.Pod, error)
	Upsert(namespace string, newObj *v1.Pod, merge Merger) (*v1.Pod, error)
	WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1.Pod, error)
	WaitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (*v1.Pod, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (*v1.Pod, error)
	WaitForConditionReadyTrue(ctx context.Context, namespace string, name string, interval time.Duration) (*v1.Pod, error)
	WaitForConditionInitializedTrue(ctx context.Context, namespace string, name string, interval time.Duration) (*v1.Pod, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient      cv1.PodsGetter
	upsertMutate Mutator
}

func (core *coreClient) preprocessUpsert(obj *v1.Pod) error {
	if core.upsertMutate == nil {
		return nil
	}

	return core.upsertMutate(obj)
}

// Create inserts the given v1.Pod into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(namespace string, obj *v1.Pod, opts ...CreateOption) (*v1.Pod, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Pods(namespace).Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update(namespace string, obj *v1.Pod, opts ...UpdateOption) (*v1.Pod, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Pods(namespace).Update(obj)
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls.
func (core *coreClient) Transform(namespace string, name string, mutator Mutator) (*v1.Pod, error) {
	obj, err := core.Get(namespace, name)
	if err != nil {
		return nil, err
	}

	if err := mutator(obj); err != nil {
		return nil, err
	}

	return core.Update(namespace, obj)
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(namespace string, name string, opts ...GetOption) (*v1.Pod, error) {
	res, err := core.kclient.Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the OperatorConfig with the name %q: %v", name, err)
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete(namespace string, name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.Pods(namespace).Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the OperatorConfig with the name %q: %v", name, err)
	}

	return nil
}

func (cfg deleteConfig) ToDeleteOptions() *metav1.DeleteOptions {
	resp := metav1.DeleteOptions{}

	if cfg.ForegroundDeletion {
		propigationPolicy := metav1.DeletePropagationForeground
		resp.PropagationPolicy = &propigationPolicy
	}

	return &resp
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List(namespace string, opts ...ListOption) ([]v1.Pod, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.Pods(namespace).List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list OperatorConfigs: %v", err)
	}

	if cfg.filter == nil {
		return res.Items, nil
	}

	return List(res.Items).Filter(cfg.filter), nil
}

func (cfg listConfig) ToListOptions() (resp metav1.ListOptions) {
	if cfg.fieldSelector != nil {
		resp.FieldSelector = metav1.FormatLabelSelector(metav1.SetAsLabelSelector(cfg.fieldSelector))
	}

	return
}

// Merger is a type to merge an existing value with a new one.
type Merger func(newObj, oldObj *v1.Pod) *v1.Pod

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert(namespace string, newObj *v1.Pod, merge Merger) (*v1.Pod, error) {
	// NOTE: the field selector may be ignored by some Kubernetes resources
	// so we double check down below.
	existing, err := core.List(namespace, WithListFieldSelector(map[string]string{"metadata.name": newObj.Name}))
	if err != nil {
		return nil, err
	}

	for _, oldObj := range existing {
		if oldObj.Name == newObj.Name {
			return core.Update(namespace, merge(newObj, &oldObj))
		}
	}

	return core.Create(namespace, newObj)
}

// WaitFor is a convenience wrapper for WaitForE that fails if the error
// passed is non-nil. It allows the use of Predicates instead of ConditionFuncE.
func (core *coreClient) WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1.Pod, error) {
	return core.WaitForE(ctx, namespace, name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by WaitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(instance *v1.Pod, apiErr error) (done bool, err error)

// WaitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) WaitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (instance *v1.Pod, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.Pods(namespace).Get(name, metav1.GetOptions{})
		if done, err = condition(instance, err); done {
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
func ConditionDeleted(_ *v1.Pod, apiErr error) (bool, error) {
	if apiErr != nil {
		if apierrors.IsNotFound(apiErr) {
			apiErr = nil
		}

		return true, apiErr
	}

	return false, nil
}

// wrapPredicate converts a predicate to a ConditionFuncE that fails if the
// error is not nil
func wrapPredicate(condition Predicate) ConditionFuncE {
	return func(obj *v1.Pod, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines WaitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1.Pod, err error) {
	return core.WaitForE(ctx, namespace, name, interval, ConditionDeleted)
}

func checkConditionTrue(obj *v1.Pod, err error, condition apis.ConditionType) (bool, error) {
	if err != nil {
		return true, err
	}

	// don't propagate old statuses
	if !ObservedGenerationMatchesGeneration(obj) {
		return false, nil
	}

	for _, cond := range ExtractConditions(obj) {
		if cond.Type == condition {
			switch {
			case cond.IsTrue():
				return true, nil

			case cond.IsUnknown():
				return false, nil

			default:
				// return true and a failure assuming IsFalse and other statuses can't be
				// recovered from because they violate the K8s spec
				return true, fmt.Errorf("checking %s failed, status: %s message: %s reason: %s", cond.Type, cond.Status, cond.Message, cond.Reason)
			}
		}
	}

	return false, nil
}

// ConditionReadyTrue is a ConditionFuncE that waits for Condition{Ready  Ready} to
// become true and fails with an error if the condition becomes false.
func ConditionReadyTrue(obj *v1.Pod, err error) (bool, error) {
	return checkConditionTrue(obj, err, ConditionReady)
}

// WaitForConditionReadyTrue is a utility function that combines WaitForE with ConditionReadyTrue.
func (core *coreClient) WaitForConditionReadyTrue(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1.Pod, err error) {
	return core.WaitForE(ctx, namespace, name, interval, ConditionReadyTrue)
}

// ConditionInitializedTrue is a ConditionFuncE that waits for Condition{Initialized v1.PodInitialized } to
// become true and fails with an error if the condition becomes false.
func ConditionInitializedTrue(obj *v1.Pod, err error) (bool, error) {
	return checkConditionTrue(obj, err, ConditionInitialized)
}

// WaitForConditionInitializedTrue is a utility function that combines WaitForE with ConditionInitializedTrue.
func (core *coreClient) WaitForConditionInitializedTrue(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1.Pod, err error) {
	return core.WaitForE(ctx, namespace, name, interval, ConditionInitializedTrue)
}
