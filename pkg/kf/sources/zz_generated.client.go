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

package sources

// Generator defined imports
import (
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmp"
)

// User defined imports
import (
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
)

////////////////////////////////////////////////////////////////////////////////
// Functional Utilities
////////////////////////////////////////////////////////////////////////////////

const (
	// Kind contains the kind for the backing Kubernetes API.
	Kind = "Source"

	// APIVersion contains the version for the backing Kubernetes API.
	APIVersion = "kf.dev/v1alpha1"
)

// Predicate is a boolean function for a v1alpha1.Source.
type Predicate func(*v1alpha1.Source) bool

// AllPredicate is a predicate that passes if all children pass.
func AllPredicate(children ...Predicate) Predicate {
	return func(obj *v1alpha1.Source) bool {
		for _, filter := range children {
			if !filter(obj) {
				return false
			}
		}

		return true
	}
}

// Mutator is a function that changes v1alpha1.Source.
type Mutator func(*v1alpha1.Source) error

// DiffWrapper wraps a mutator and prints out the diff between the original object
// and the one it returns if there's no error.
func DiffWrapper(w io.Writer, mutator Mutator) Mutator {
	return func(mutable *v1alpha1.Source) error {
		before := mutable.DeepCopy()

		if err := mutator(mutable); err != nil {
			return err
		}

		FormatDiff(w, "old", "new", before, mutable)

		return nil
	}
}

// FormatDiff creates a diff between two v1alpha1.Sources and writes it to the given
// writer.
func FormatDiff(w io.Writer, leftName, rightName string, left, right *v1alpha1.Source) {
	diff, err := kmp.SafeDiff(left, right)
	switch {
	case err != nil:
		fmt.Fprintf(w, "couldn't format diff: %s\n", err.Error())

	case diff == "":
		fmt.Fprintln(w, "No changes")

	default:
		fmt.Fprintf(w, "Build Diff (-%s +%s):\n", leftName, rightName)
		// go-cmp randomly chooses to prefix lines with non-breaking spaces or
		// regular spaces to prevent people from using it as a real diff/patch
		// tool. We normalize them so our outputs will be consistent.
		fmt.Fprintln(w, strings.ReplaceAll(diff, " ", " "))
	}
}

// List represents a collection of v1alpha1.Source.
type List []v1alpha1.Source

// Filter returns a new list items for which the predicates fails removed.
func (list List) Filter(filter Predicate) (out List) {
	for _, v := range list {
		if filter(&v) {
			out = append(out, v)
		}
	}

	return
}

// MutatorList is a list of mutators.
type MutatorList []Mutator

// Apply passes the given value to each of the mutators in the list failing if
// one of them returns an error.
func (list MutatorList) Apply(svc *v1alpha1.Source) error {
	for _, mutator := range list {
		if err := mutator(svc); err != nil {
			return err
		}
	}

	return nil
}

// LabelSetMutator creates a mutator that sets the given labels on the object.
func LabelSetMutator(labels map[string]string) Mutator {
	return func(obj *v1alpha1.Source) error {
		if obj.Labels == nil {
			obj.Labels = make(map[string]string)
		}

		for key, value := range labels {
			obj.Labels[key] = value
		}

		return nil
	}
}

// LabelEqualsPredicate validates that the given label exists exactly on the object.
func LabelEqualsPredicate(key, value string) Predicate {
	return func(obj *v1alpha1.Source) bool {
		return obj.Labels[key] == value
	}
}

// LabelsContainsPredicate validates that the given label exists on the object.
func LabelsContainsPredicate(key string) Predicate {
	return func(obj *v1alpha1.Source) bool {
		_, ok := obj.Labels[key]
		return ok
	}
}

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

// Client is the interface for interacting with v1alpha1.Source types as Build CF style objects.
type Client interface {
	Create(namespace string, obj *v1alpha1.Source, opts ...CreateOption) (*v1alpha1.Source, error)
	Update(namespace string, obj *v1alpha1.Source, opts ...UpdateOption) (*v1alpha1.Source, error)
	Transform(namespace string, name string, transformer Mutator) error
	Get(namespace string, name string, opts ...GetOption) (*v1alpha1.Source, error)
	Delete(namespace string, name string, opts ...DeleteOption) error
	List(namespace string, opts ...ListOption) ([]v1alpha1.Source, error)
	Upsert(namespace string, newObj *v1alpha1.Source, merge Merger) (*v1alpha1.Source, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient cv1alpha1.SourcesGetter

	upsertMutate        MutatorList
	membershipValidator Predicate
}

func (core *coreClient) preprocessUpsert(obj *v1alpha1.Source) error {
	if err := core.upsertMutate.Apply(obj); err != nil {
		return err
	}

	return nil
}

// Create inserts the given v1alpha1.Source into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(namespace string, obj *v1alpha1.Source, opts ...CreateOption) (*v1alpha1.Source, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Sources(namespace).Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update(namespace string, obj *v1alpha1.Source, opts ...UpdateOption) (*v1alpha1.Source, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Sources(namespace).Update(obj)
}

// Transform performs a read/modify/write on the object with the given name.
// Transform manages the options for the Get and Update calls.
func (core *coreClient) Transform(namespace string, name string, mutator Mutator) error {
	obj, err := core.Get(namespace, name)
	if err != nil {
		return err
	}

	if err := mutator(obj); err != nil {
		return err
	}

	if _, err := core.Update(namespace, obj); err != nil {
		return err
	}

	return nil
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(namespace string, name string, opts ...GetOption) (*v1alpha1.Source, error) {
	res, err := core.kclient.Sources(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the Build with the name %q: %v", name, err)
	}

	if core.membershipValidator(res) {
		return res, nil
	}

	return nil, fmt.Errorf("an object with the name %s exists, but it doesn't appear to be a Build", name)
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete(namespace string, name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.Sources(namespace).Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the Build with the name %q: %v", name, err)
	}

	return nil
}

func (cfg deleteConfig) ToDeleteOptions() *metav1.DeleteOptions {
	resp := metav1.DeleteOptions{}

	if cfg.ForegroundDeletion {
		propigationPolicy := metav1.DeletePropagationForeground
		resp.PropagationPolicy = &propigationPolicy
	}

	if cfg.DeleteImmediately {
		resp.GracePeriodSeconds = new(int64)
	}

	return &resp
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List(namespace string, opts ...ListOption) ([]v1alpha1.Source, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.Sources(namespace).List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list Builds: %v", err)
	}

	return List(res.Items).
		Filter(core.membershipValidator).
		Filter(AllPredicate(cfg.filters...)), nil
}

func (cfg listConfig) ToListOptions() (resp metav1.ListOptions) {
	if cfg.fieldSelector != nil {
		resp.FieldSelector = metav1.FormatLabelSelector(metav1.SetAsLabelSelector(cfg.fieldSelector))
	}

	if cfg.labelSelector != nil {
		resp.LabelSelector = metav1.FormatLabelSelector(metav1.SetAsLabelSelector(cfg.labelSelector))
	}

	return
}

// Merger is a type to merge an existing value with a new one.
type Merger func(newObj, oldObj *v1alpha1.Source) *v1alpha1.Source

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert(namespace string, newObj *v1alpha1.Source, merge Merger) (*v1alpha1.Source, error) {
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
