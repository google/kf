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

package cluster

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
	cv1beta1 "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/typed/servicecatalog/v1beta1"
	v1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
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
	return false
}

// GroupVersionResource gets the GVR struct for the resource.
func (*ResourceInfo) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "servicecatalog.k8s.io",
		Version:  "v1beta1",
		Resource: "clusterservicebrokers",
	}
}

// GroupVersionKind gets the GVK struct for the resource.
func (*ResourceInfo) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "servicecatalog.k8s.io",
		Version: "v1beta1",
		Kind:    "ClusterServiceBroker",
	}
}

// FriendlyName gets the user-facing name of the resource.
func (*ResourceInfo) FriendlyName() string {
	return "ClusterServiceBroker"
}

var (
	ConditionReady = apis.ConditionType(v1beta1.ServiceBrokerConditionReady)
)

// Predicate is a boolean function for a v1beta1.ClusterServiceBroker.
type Predicate func(*v1beta1.ClusterServiceBroker) bool

// Mutator is a function that changes v1beta1.ClusterServiceBroker.
type Mutator func(*v1beta1.ClusterServiceBroker) error

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator Mutator) Mutator {
	return func(mutable *v1beta1.ClusterServiceBroker) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two v1beta1.ClusterServiceBrokers and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1beta1.ClusterServiceBroker) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "ClusterServiceBroker Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}

// List represents a collection of v1beta1.ClusterServiceBroker.
type List []v1beta1.ClusterServiceBroker

// Filter returns a new list items for which the predicates fails removed.
func (list List) Filter(filter Predicate) (out List) {
	for _, v := range list {
		if filter(&v) {
			out = append(out, v)
		}
	}

	return
}

// ExtractConditions converts the native condition types into an apis.Condition
// array with the Type, Status, Reason, and Message fields intact.
func ExtractConditions(obj *v1beta1.ClusterServiceBroker) (extracted []apis.Condition) {
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

// Client is the interface for interacting with v1beta1.ClusterServiceBroker types as ClusterServiceBroker CF style objects.
type Client interface {
	Create(obj *v1beta1.ClusterServiceBroker, opts ...CreateOption) (*v1beta1.ClusterServiceBroker, error)
	Update(obj *v1beta1.ClusterServiceBroker, opts ...UpdateOption) (*v1beta1.ClusterServiceBroker, error)
	Transform(name string, transformer Mutator) (*v1beta1.ClusterServiceBroker, error)
	Get(name string, opts ...GetOption) (*v1beta1.ClusterServiceBroker, error)
	Delete(name string, opts ...DeleteOption) error
	List(opts ...ListOption) ([]v1beta1.ClusterServiceBroker, error)
	Upsert(newObj *v1beta1.ClusterServiceBroker, merge Merger) (*v1beta1.ClusterServiceBroker, error)
	WaitFor(ctx context.Context, name string, interval time.Duration, condition Predicate) (*v1beta1.ClusterServiceBroker, error)
	WaitForE(ctx context.Context, name string, interval time.Duration, condition ConditionFuncE) (*v1beta1.ClusterServiceBroker, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, name string, interval time.Duration) (*v1beta1.ClusterServiceBroker, error)
	WaitForConditionReadyTrue(ctx context.Context, name string, interval time.Duration) (*v1beta1.ClusterServiceBroker, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient      cv1beta1.ClusterServiceBrokersGetter
	upsertMutate Mutator
}

func (core *coreClient) preprocessUpsert(obj *v1beta1.ClusterServiceBroker) error {
	if core.upsertMutate == nil {
		return nil
	}

	return core.upsertMutate(obj)
}

// Create inserts the given v1beta1.ClusterServiceBroker into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(obj *v1beta1.ClusterServiceBroker, opts ...CreateOption) (*v1beta1.ClusterServiceBroker, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.ClusterServiceBrokers().Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update(obj *v1beta1.ClusterServiceBroker, opts ...UpdateOption) (*v1beta1.ClusterServiceBroker, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.ClusterServiceBrokers().Update(obj)
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls.
func (core *coreClient) Transform(name string, mutator Mutator) (*v1beta1.ClusterServiceBroker, error) {
	obj, err := core.Get(name)
	if err != nil {
		return nil, err
	}

	if err := mutator(obj); err != nil {
		return nil, err
	}

	return core.Update(obj)
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(name string, opts ...GetOption) (*v1beta1.ClusterServiceBroker, error) {
	res, err := core.kclient.ClusterServiceBrokers().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the ClusterServiceBroker with the name %q: %v", name, err)
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete(name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.ClusterServiceBrokers().Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the ClusterServiceBroker with the name %q: %v", name, err)
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
func (core *coreClient) List(opts ...ListOption) ([]v1beta1.ClusterServiceBroker, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.ClusterServiceBrokers().List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list ClusterServiceBrokers: %v", err)
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
type Merger func(newObj, oldObj *v1beta1.ClusterServiceBroker) *v1beta1.ClusterServiceBroker

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert(newObj *v1beta1.ClusterServiceBroker, merge Merger) (*v1beta1.ClusterServiceBroker, error) {
	// NOTE: the field selector may be ignored by some Kubernetes resources
	// so we double check down below.
	existing, err := core.List(WithListFieldSelector(map[string]string{"metadata.name": newObj.Name}))
	if err != nil {
		return nil, err
	}

	for _, oldObj := range existing {
		if oldObj.Name == newObj.Name {
			return core.Update(merge(newObj, &oldObj))
		}
	}

	return core.Create(newObj)
}

// WaitFor is a convenience wrapper for WaitForE that fails if the error
// passed is non-nil. It allows the use of Predicates instead of ConditionFuncE.
func (core *coreClient) WaitFor(ctx context.Context, name string, interval time.Duration, condition Predicate) (*v1beta1.ClusterServiceBroker, error) {
	return core.WaitForE(ctx, name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by WaitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(instance *v1beta1.ClusterServiceBroker, apiErr error) (done bool, err error)

// WaitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) WaitForE(ctx context.Context, name string, interval time.Duration, condition ConditionFuncE) (instance *v1beta1.ClusterServiceBroker, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.ClusterServiceBrokers().Get(name, metav1.GetOptions{})
		if done, err = condition(instance, err); done {
			return
		}

		select {
		case <-tick:
			// repeat instance check
		case <-ctx.Done():
			return nil, errors.New("waiting for ClusterServiceBroker timed out")
		}
	}
}

// ConditionDeleted is a ConditionFuncE that succeeds if the error returned by
// the cluster was a not found error.
func ConditionDeleted(_ *v1beta1.ClusterServiceBroker, apiErr error) (bool, error) {
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
	return func(obj *v1beta1.ClusterServiceBroker, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines WaitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, name string, interval time.Duration) (instance *v1beta1.ClusterServiceBroker, err error) {
	return core.WaitForE(ctx, name, interval, ConditionDeleted)
}

func checkConditionTrue(obj *v1beta1.ClusterServiceBroker, err error, condition apis.ConditionType) (bool, error) {
	if err != nil {
		return true, err
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

// ConditionReadyTrue is a ConditionFuncE that waits for Condition{Ready v1beta1.ServiceBrokerConditionReady } to
// become true and fails with an error if the condition becomes false.
func ConditionReadyTrue(obj *v1beta1.ClusterServiceBroker, err error) (bool, error) {
	return checkConditionTrue(obj, err, ConditionReady)
}

// WaitForConditionReadyTrue is a utility function that combines WaitForE with ConditionReadyTrue.
func (core *coreClient) WaitForConditionReadyTrue(ctx context.Context, name string, interval time.Duration) (instance *v1beta1.ClusterServiceBroker, err error) {
	return core.WaitForE(ctx, name, interval, ConditionReadyTrue)
}
