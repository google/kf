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
	Transform({{ $nssig }} name string, transformer Mutator) error
	Get({{ $nssig }} name string, opts ...GetOption) (*{{.Type}}, error)
	Delete({{ $nssig }} name string, opts ...DeleteOption) error
	List({{ $nssig }} opts ...ListOption) ([]{{.Type}}, error)
	Upsert({{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error)

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient {{.ClientType}}

	upsertMutate        MutatorList
	membershipValidator Predicate
}

func (core *coreClient) preprocessUpsert(obj *{{.Type}}) error {
	if err := core.upsertMutate.Apply(obj); err != nil {
		return err
	}

	return nil
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

// Transform performs a read/modify/write on the object with the given name.
// Transform manages the options for the Get and Update calls.
func (core *coreClient) Transform({{ $nssig }} name string, mutator Mutator) error {
	obj, err := core.Get({{ $nsparam }} name)
	if err != nil {
		return err
	}

	if err := mutator(obj); err != nil {
		return err
	}

	if _, err := core.Update({{ $nsparam }} obj); err != nil {
		return err
	}

	return nil
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get({{ $nssig }} name string, opts ...GetOption) (*{{.Type}}, error) {
	res, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't get the {{.CF.Name}} with the name %q: %v", name, err)
	}

	if core.membershipValidator(res) {
		return res, nil
	}

	return nil, fmt.Errorf("an object with the name %s exists, but it doesn't appear to be a {{.CF.Name}}", name)
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

	if cfg.DeleteImmediately {
		resp.GracePeriodSeconds = new(int64)
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
type Merger func(newObj, oldObj *{{.Type}}) *{{.Type}}

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
func (core *coreClient) Upsert({{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error) {
	// NOTE: the field selector may be ignored by some Kubernetes resources
	// so we double check down below.
	existing, err := core.List({{ $nsparam }} WithListfieldSelector(map[string]string{"metadata.name": newObj.Name}))
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
`))
