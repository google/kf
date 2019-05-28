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

package apps

// Generator defined imports
import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// User defined imports
import (
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

////////////////////////////////////////////////////////////////////////////////
// Functional Utilities
////////////////////////////////////////////////////////////////////////////////

const (
	// Kind contains the kind for the backing Kubernetes API.
	Kind = "Service"

	// APIVersion contains the version for the backing Kubernetes API.
	APIVersion = "serving.knative.dev/v1alpha1"
)

// Predicate is a boolean function for a serving.Service.
type Predicate func(*serving.Service) bool

// AllPredicate is a predicate that passes if all children pass.
func AllPredicate(children ...Predicate) Predicate {
	return func(obj *serving.Service) bool {
		for _, filter := range children {
			if !filter(obj) {
				return false
			}
		}

		return true
	}
}

// Mutator is a function that changes serving.Service.
type Mutator func(*serving.Service) error

// List represents a collection of serving.Service.
type List []serving.Service

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
func (list MutatorList) Apply(svc *serving.Service) error {
	for _, mutator := range list {
		if err := mutator(svc); err != nil {
			return err
		}
	}

	return nil
}

// LabelSetMutator creates a mutator that sets the given labels on the object.
func LabelSetMutator(labels map[string]string) Mutator {
	return func(obj *serving.Service) error {
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
	return func(obj *serving.Service) bool {
		return obj.Labels[key] == value
	}
}

// LabelsContainsPredicate validates that the given label exists on the object.
func LabelsContainsPredicate(key string) Predicate {
	return func(obj *serving.Service) bool {
		_, ok := obj.Labels[key]
		return ok
	}
}

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

// Client is the interface for interacting with serving.Service types as App CF style objects.
type Client interface {
	Create(namespace string, obj *serving.Service, opts ...CreateOption) (*serving.Service, error)
	Update(namespace string, obj *serving.Service, opts ...UpdateOption) (*serving.Service, error)
	Transform(namespace string, name string, transformer Mutator) error
	Get(namespace string, name string, opts ...GetOption) (*serving.Service, error)
	Delete(namespace string, name string, opts ...DeleteOption) error
	List(namespace string, opts ...ListOption) ([]serving.Service, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient cserving.ServingV1alpha1Interface

	upsertMutate        MutatorList
	membershipValidator Predicate
}

func (core *coreClient) preprocessUpsert(obj *serving.Service) error {
	if err := core.upsertMutate.Apply(obj); err != nil {
		return err
	}

	return nil
}

// Create inserts the given serving.Service into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(namespace string, obj *serving.Service, opts ...CreateOption) (*serving.Service, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Services(namespace).Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update(namespace string, obj *serving.Service, opts ...UpdateOption) (*serving.Service, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.Services(namespace).Update(obj)
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
func (core *coreClient) Get(namespace string, name string, opts ...GetOption) (*serving.Service, error) {
	res, err := core.kclient.Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the App with the name %q: %v", name, err)
	}

	if core.membershipValidator(res) {
		return res, nil
	}

	return nil, fmt.Errorf("an object with the name %s exists, but it doesn't appear to be a App", name)
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete(namespace string, name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.Services(namespace).Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the App with the name %q: %v", name, err)
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
func (core *coreClient) List(namespace string, opts ...ListOption) ([]serving.Service, error) {
	cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.Services(namespace).List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list Apps: %v", err)
	}

	return List(res.Items).
		Filter(core.membershipValidator).
		Filter(AllPredicate(cfg.filters...)), nil
}

func (cfg listConfig) ToListOptions() (resp metav1.ListOptions) {
	if cfg.fieldSelector != nil {
		resp.FieldSelector = metav1.SetAsLabelSelector(cfg.fieldSelector).String()
	}

	if cfg.labelSelector != nil {
		resp.LabelSelector = metav1.SetAsLabelSelector(cfg.labelSelector).String()
	}

	return
}
