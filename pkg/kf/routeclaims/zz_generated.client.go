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

package routeclaims

// Generator defined imports
import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"knative.dev/pkg/kmp"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// User defined imports
import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
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
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "routeclaims",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "kf.dev",
		Version: "v1alpha1",
		Kind:    "RouteClaim",
	}
}

// FriendlyName gets the user-facing name of the resource.
func (*ResourceInfo) FriendlyName() string {
	return "RouteClaim"
}

// Predicate is a boolean function for a v1alpha1.RouteClaim.
type Predicate func(*v1alpha1.RouteClaim) bool

// Mutator is a function that changes v1alpha1.RouteClaim.
type Mutator func(*v1alpha1.RouteClaim) error

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator Mutator) Mutator {
	return func(mutable *v1alpha1.RouteClaim) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two v1alpha1.RouteClaims and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1alpha1.RouteClaim) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "RouteClaim Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}

// List represents a collection of v1alpha1.RouteClaim.
type List []v1alpha1.RouteClaim

// Filter returns a new list items for which the predicates fails removed.
func (list List) Filter(filter Predicate) (out List) {
	for _, v := range list {
		if filter(&v) {
			out = append(out, v)
		}
	}

	return
}

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

// Client is the interface for interacting with v1alpha1.RouteClaim types as RouteClaim CF style objects.
type Client interface {
	Create(namespace string, obj *v1alpha1.RouteClaim, opts ...CreateOption) (*v1alpha1.RouteClaim, error)
	Update(namespace string, obj *v1alpha1.RouteClaim, opts ...UpdateOption) (*v1alpha1.RouteClaim, error)
	Transform(namespace string, name string, transformer Mutator) (*v1alpha1.RouteClaim, error)
	Get(namespace string, name string, opts ...GetOption) (*v1alpha1.RouteClaim, error)
	Delete(namespace string, name string, opts ...DeleteOption) error
	List(namespace string, opts ...ListOption) ([]v1alpha1.RouteClaim, error)
	Upsert(namespace string, newObj *v1alpha1.RouteClaim, merge Merger) (*v1alpha1.RouteClaim, error)
	WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1alpha1.RouteClaim, error)
	WaitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (*v1alpha1.RouteClaim, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (*v1alpha1.RouteClaim, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient      cv1alpha1.RouteClaimsGetter
	upsertMutate Mutator
}

func (core *coreClient) preprocessUpsert(obj *v1alpha1.RouteClaim) error {
	if core.upsertMutate == nil {
		return nil
	}

	return core.upsertMutate(obj)
}

// Create inserts the given v1alpha1.RouteClaim into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(namespace string, obj *v1alpha1.RouteClaim, opts ...CreateOption) (*v1alpha1.RouteClaim, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.RouteClaims(namespace).Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update(namespace string, obj *v1alpha1.RouteClaim, opts ...UpdateOption) (*v1alpha1.RouteClaim, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.RouteClaims(namespace).Update(obj)
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls.
func (core *coreClient) Transform(namespace string, name string, mutator Mutator) (*v1alpha1.RouteClaim, error) {
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
func (core *coreClient) Get(namespace string, name string, opts ...GetOption) (*v1alpha1.RouteClaim, error) {
	res, err := core.kclient.RouteClaims(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the RouteClaim with the name %q: %v", name, err)
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete(namespace string, name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.RouteClaims(namespace).Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the RouteClaim with the name %q: %v", name, err)
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
func (core *coreClient) List(namespace string, opts ...ListOption) ([]v1alpha1.RouteClaim, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.RouteClaims(namespace).List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list RouteClaims: %v", err)
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
type Merger func(newObj, oldObj *v1alpha1.RouteClaim) *v1alpha1.RouteClaim

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert(namespace string, newObj *v1alpha1.RouteClaim, merge Merger) (*v1alpha1.RouteClaim, error) {
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
func (core *coreClient) WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1alpha1.RouteClaim, error) {
	return core.WaitForE(ctx, namespace, name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by WaitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(instance *v1alpha1.RouteClaim, apiErr error) (done bool, err error)

// WaitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) WaitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (instance *v1alpha1.RouteClaim, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.RouteClaims(namespace).Get(name, metav1.GetOptions{})
		if done, err = condition(instance, err); done {
			return
		}

		select {
		case <-tick:
			// repeat instance check
		case <-ctx.Done():
			return nil, errors.New("waiting for RouteClaim timed out")
		}
	}
}

// ConditionDeleted is a ConditionFuncE that succeeds if the error returned by
// the cluster was a not found error.
func ConditionDeleted(_ *v1alpha1.RouteClaim, apiErr error) (bool, error) {
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
	return func(obj *v1alpha1.RouteClaim, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines WaitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1alpha1.RouteClaim, err error) {
	return core.WaitForE(ctx, namespace, name, interval, ConditionDeleted)
}
