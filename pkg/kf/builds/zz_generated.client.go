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

package builds

// Generator defined imports
import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// User defined imports
import (
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
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
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "builds",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind(context.Context) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "kf.dev",
		Version: "v1alpha1",
		Kind:    "Build",
	}
}

// FriendlyName gets the user-facing name of the resource.
func (*ResourceInfo) FriendlyName() string {
	return "Build"
}

// Predicate is a boolean function for a v1alpha1.Build.
type Predicate func(*v1alpha1.Build) bool

// Mutator is a function that changes v1alpha1.Build.
type Mutator func(*v1alpha1.Build) error

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

// Client is the interface for interacting with v1alpha1.Build types as Build CF style objects.
type Client interface {
	Create(ctx context.Context, namespace string, obj *v1alpha1.Build) (*v1alpha1.Build, error)
	Transform(ctx context.Context, namespace string, name string, transformer Mutator) (*v1alpha1.Build, error)
	Get(ctx context.Context, namespace string, name string) (*v1alpha1.Build, error)
	Delete(ctx context.Context, namespace string, name string) error
	List(ctx context.Context, namespace string) ([]v1alpha1.Build, error)
	Upsert(ctx context.Context, namespace string, newObj *v1alpha1.Build, merge Merger) (*v1alpha1.Build, error)
	WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1alpha1.Build, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (*v1alpha1.Build, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient cv1alpha1.BuildsGetter
}

// Create inserts the given v1alpha1.Build into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(ctx context.Context, namespace string, obj *v1alpha1.Build) (*v1alpha1.Build, error) {
	return core.kclient.Builds(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls. The transform will be retried as long as the resource is in
// conflict.
func (core *coreClient) Transform(ctx context.Context, namespace string, name string, mutator Mutator) (*v1alpha1.Build, error) {
	for {
		obj, err := core.Get(ctx, namespace, name)
		if err != nil {
			return nil, err
		}

		if err := mutator(obj); err != nil {
			return nil, err
		}

		result, err := core.kclient.Builds(namespace).Update(ctx, obj, metav1.UpdateOptions{})

		if apierrors.IsConflict(err) {
			continue
		}
		return result, err
	}
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(ctx context.Context, namespace string, name string) (*v1alpha1.Build, error) {
	res, err := core.kclient.Builds(namespace).Get(ctx, name, metav1.GetOptions{})
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
	if err := core.kclient.Builds(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &foreground}); err != nil {
		return fmt.Errorf("couldn't delete the Build with the name %q: %v", name, err)
	}

	return nil
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List(ctx context.Context, namespace string) ([]v1alpha1.Build, error) {
	res, err := core.kclient.Builds(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list Builds: %v", err)
	}

	return res.Items, nil
}

// Merger is a type to merge an existing value with a new one.
type Merger func(newObj, oldObj *v1alpha1.Build) *v1alpha1.Build

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
// If the update results in a conflict error, then it is retried with the new
// object. Meaning, the merge function is invoked again.
func (core *coreClient) Upsert(ctx context.Context, namespace string, newObj *v1alpha1.Build, merge Merger) (*v1alpha1.Build, error) {
	for ctx.Err() == nil {
		// kclient must be used so the error code can be validated by apierrors
		oldObj, err := core.kclient.Builds(namespace).Get(ctx, newObj.Name, metav1.GetOptions{})

		switch {
		case apierrors.IsNotFound(err):
			return core.Create(ctx, namespace, newObj)
		case err != nil:
			return nil, err
		}

		updated, err := core.kclient.Builds(namespace).Update(ctx, merge(newObj, oldObj), metav1.UpdateOptions{})
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
func (core *coreClient) WaitFor(ctx context.Context, namespace string, name string, interval time.Duration, condition Predicate) (*v1alpha1.Build, error) {
	return core.waitForE(ctx, namespace, name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by waitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(ctx context.Context, instance *v1alpha1.Build, apiErr error) (done bool, err error)

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
func (core *coreClient) waitForE(ctx context.Context, namespace string, name string, interval time.Duration, condition ConditionFuncE) (instance *v1alpha1.Build, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.Builds(namespace).Get(ctx, name, metav1.GetOptions{})
		if done, err = condition(ctx, instance, err); done {
			return
		}

		select {
		case <-tick:
			// repeat instance check
		case <-ctx.Done():
			return nil, errors.New("waiting for Build timed out")
		}
	}
}

// ConditionDeleted is a ConditionFuncE that succeeds if the error returned by
// the cluster was a not found error.
func ConditionDeleted(ctx context.Context, _ *v1alpha1.Build, apiErr error) (bool, error) {
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
	return func(ctx context.Context, obj *v1alpha1.Build, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines waitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, namespace string, name string, interval time.Duration) (instance *v1alpha1.Build, err error) {
	return core.waitForE(ctx, namespace, name, interval, ConditionDeleted)
}
