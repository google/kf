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

package clientgen

import (
	"text/template"

	"github.com/google/kf/pkg/kf/internal/tools/generator"
)

var clientTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

{{ $nssig := "" }}
{{ $ns := "" }}
{{ $nsparam := "" }}

{{ if .Kubernetes.Namespaced }}
  {{ $nssig = "namespace string," }}
  {{ $ns = "namespace" }}
	{{ $nsparam = "namespace," }}
{{ end }}

// Client is the interface for interacting with {{.Type}} types as {{.CF.Name}} CF style objects.
type Client interface {
	Create({{ $nssig }} obj *{{.Type}}, opts ...CreateOption) (*{{.Type}}, error)
	Update({{ $nssig }} obj *{{.Type}}, opts ...UpdateOption) (*{{.Type}}, error)
	Transform({{ $nssig }} name string, transformer Mutator) (*{{.Type}}, error)
	Get({{ $nssig }} name string, opts ...GetOption) (*{{.Type}}, error)
	Delete({{ $nssig }} name string, opts ...DeleteOption) error
	List({{ $nssig }} opts ...ListOption) ([]{{.Type}}, error)
	Upsert({{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error)
	WaitFor(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition Predicate) (*{{.Type}}, error)
	WaitForE(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition ConditionFuncE) (*{{.Type}}, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient {{.ClientType}}
	upsertMutate Mutator
}

func (core *coreClient) preprocessUpsert(obj *{{.Type}}) error {
	if core.upsertMutate == nil {
		return nil
	}

	return core.upsertMutate(obj)
}

// Create inserts the given {{.Type}} into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create({{ $nssig }} obj *{{.Type}}, opts ...CreateOption) (*{{.Type}}, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Create(obj)
}

// Update replaces the existing object in the cluster with the new one.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Update({{ $nssig }} obj *{{.Type}}, opts ...UpdateOption) (*{{.Type}}, error) {
	if err := core.preprocessUpsert(obj); err != nil {
		return nil, err
	}

	return core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Update(obj)
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls.
func (core *coreClient) Transform({{ $nssig }} name string, mutator Mutator) (*{{.Type}}, error) {
	obj, err := core.Get({{ $nsparam }} name)
	if err != nil {
		return nil, err
	}

	if err := mutator(obj); err != nil {
		return nil, err
	}

	return core.Update({{ $nsparam }} obj)
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get({{ $nssig }} name string, opts ...GetOption) (*{{.Type}}, error) {
	res, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the {{.CF.Name}} with the name %q: %v", name, err)
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
func (core *coreClient) Delete({{ $nssig }} name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	if err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Delete(name, cfg.ToDeleteOptions()); err != nil {
		return fmt.Errorf("couldn't delete the {{.CF.Name}} with the name %q: %v", name, err)
	}

	return nil
}

func (cfg deleteConfig) ToDeleteOptions() (*metav1.DeleteOptions) {
	resp := metav1.DeleteOptions{}

	if cfg.ForegroundDeletion {
		propigationPolicy := metav1.DeletePropagationForeground
		resp.PropagationPolicy = &propigationPolicy
	}

	return &resp
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List({{ $nssig }} opts ...ListOption) ([]{{.Type}}, error) {
  cfg := ListOptionDefaults().Extend(opts).toConfig()

	res, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).List(cfg.ToListOptions())
	if err != nil {
		return nil, fmt.Errorf("couldn't list {{.CF.Name}}s: %v", err)
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
type Merger func(newObj, oldObj *{{.Type}}) *{{.Type}}

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert({{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error) {
	// NOTE: the field selector may be ignored by some Kubernetes resources
	// so we double check down below.
	existing, err := core.List({{ $nsparam }} WithListFieldSelector(map[string]string{"metadata.name": newObj.Name}))
	if err != nil {
		return nil, err
	}

	for _, oldObj := range existing {
		if oldObj.Name == newObj.Name {
			return core.Update({{ $nsparam }} merge(newObj, &oldObj))
		}
	}

	return core.Create({{ $nsparam }} newObj)
}

// WaitFor is a convenience wrapper for WaitForE that fails if the error
// passed is non-nil. It allows the use of Predicates instead of ConditionFuncE.
func (core *coreClient) WaitFor(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition Predicate) (*{{.Type}}, error) {
	return core.WaitForE(ctx, {{ $nsparam }} name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by WaitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(instance *{{.Type}}, apiErr error) (done bool, err error)

// WaitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) WaitForE(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition ConditionFuncE) (instance *{{.Type}}, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(name, metav1.GetOptions{})
		if done, err = condition(instance, err); done {
			return
		}

		select {
		case <-tick:
			// repeat instance check
		case <-ctx.Done():
			return nil, errors.New("waiting for {{.CF.Name}} timed out")
		}
	}
}

// ConditionDeleted is a ConditionFuncE that succeeds if the error returned by
// the cluster was a not found error.
func ConditionDeleted(_ *{{.Type}}, apiErr error) (bool, error) {
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
	return func(obj *{{.Type}}, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		return condition(obj), nil
	}
}
`))
