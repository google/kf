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

	"github.com/google/kf/v2/pkg/kf/internal/tools/generator"
)

var clientTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

{{ $nssig := "" }}
{{ $ns := "" }}
{{ $nsparam := "" }}
{{ $type := .Type }}

{{ if .Kubernetes.Namespaced }}
  {{ $nssig = "namespace string," }}
  {{ $ns = "namespace" }}
	{{ $nsparam = "namespace," }}
{{ end }}

// Client is the interface for interacting with {{.Type}} types as {{.CF.Name}} CF style objects.
type Client interface {
	Create(ctx context.Context, {{ $nssig }} obj *{{.Type}}) (*{{.Type}}, error)
	Transform(ctx context.Context, {{ $nssig }} name string, transformer Mutator) (*{{.Type}}, error)
	Get(ctx context.Context, {{ $nssig }} name string) (*{{.Type}}, error)
	Delete(ctx context.Context, {{ $nssig }} name string) error
	List(ctx context.Context, {{ $nssig }}) ([]{{.Type}}, error)
	Upsert(ctx context.Context, {{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error)
	WaitFor(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition Predicate) (*{{.Type}}, error)

	// Utility functions
	WaitForDeletion(ctx context.Context, {{ $nssig }} name string, interval time.Duration) (*{{.Type}}, error)
	{{ if .SupportsConditions }}{{ range .Kubernetes.Conditions }}{{.WaitForName}}(ctx context.Context, {{ $nssig }} name string, interval time.Duration) (*{{$type}}, error)
	{{ end }}
	{{ end }}

	// ClientExtension can be used by the developer to extend the client.
	ClientExtension
}

type coreClient struct {
	kclient {{.ClientType}}
}

// Create inserts the given {{.Type}} into the cluster.
// The value to be inserted will be preprocessed and validated before being sent.
func (core *coreClient) Create(ctx context.Context, {{ $nssig }} obj *{{.Type}}) (*{{.Type}}, error) {
	return core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Create(ctx, obj, metav1.CreateOptions{})
}

// Transform performs a read/modify/write on the object with the given name
// and returns the updated object. Transform manages the options for the Get and
// Update calls. The transform will be retried as long as the resource is in
// conflict.
func (core *coreClient) Transform(ctx context.Context, {{ $nssig }} name string, mutator Mutator) (*{{.Type}}, error) {
	for {
		obj, err := core.Get(ctx, {{ $nsparam }} name)
		if err != nil {
			return nil, err
		}

		if err := mutator(obj); err != nil {
			return nil, err
		}

		result, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Update(ctx, obj, metav1.UpdateOptions{})

		if apierrors.IsConflict(err) {
			continue
		}
		return result, err
	}
}

// Get retrieves an existing object in the cluster with the given name.
// The function will return an error if an object is retrieved from the cluster
// but doesn't pass the membership test of this client.
func (core *coreClient) Get(ctx context.Context, {{ $nssig }} name string) (*{{.Type}}, error) {
	res, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete removes an existing object in the cluster.
// The deleted object is NOT tested for membership before deletion.
// The object is only deleted once all of the objects it owns are deleted.
func (core *coreClient) Delete(ctx context.Context, {{ $nssig }} name string) error {
	foreground := metav1.DeletePropagationForeground
	if err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &foreground}); err != nil {
		return fmt.Errorf("couldn't delete the {{.CF.Name}} with the name %q: %v", name, err)
	}

	return nil
}

// List gets objects in the cluster and filters the results based on the
// internal membership test.
func (core *coreClient) List(ctx context.Context, {{ $nssig }}) ([]{{.Type}}, error) {
	res, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list {{.CF.Name}}s: %v", err)
	}

	return res.Items, nil
}

// Merger is a type to merge an existing value with a new one.
type Merger func(newObj, oldObj *{{.Type}}) *{{.Type}}

// Upsert inserts the object into the cluster if it doesn't already exist, or else
// calls the merge function to merge the existing and new then performs an Update.
// If the update results in a conflict error, then it is retried with the new
// object. Meaning, the merge function is invoked again.
func (core *coreClient) Upsert(ctx context.Context, {{ $nssig }} newObj *{{.Type}}, merge Merger) (*{{.Type}}, error) {
	for ctx.Err() == nil {
		// kclient must be used so the error code can be validated by apierrors
		oldObj, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(ctx, newObj.Name, metav1.GetOptions{})

		switch {
		case apierrors.IsNotFound(err):
			return core.Create(ctx, {{ $nsparam }} newObj)
		case err != nil:
			return nil, err
		}

		updated, err := core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Update(ctx, merge(newObj, oldObj), metav1.UpdateOptions{})
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
func (core *coreClient) WaitFor(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition Predicate) (*{{.Type}}, error) {
	return core.waitForE(ctx, {{ $nsparam }} name, interval, wrapPredicate(condition))
}

// ConditionFuncE is a callback used by waitForE. Done should be set to true
// once the condition succeeds and shouldn't be called anymore. The error
// will be passed back to the user.
//
// This function MAY retrieve a nil instance and an apiErr. It's up to the
// function to decide how to handle the apiErr.
type ConditionFuncE func(instance *{{.Type}}, apiErr error) (done bool, err error)

// waitForE polls for the given object every interval until the condition
// function becomes done or the timeout expires. The first poll occurs
// immediately after the function is invoked.
//
// The function polls infinitely if no timeout is supplied.
func (core *coreClient) waitForE(ctx context.Context, {{ $nssig }} name string, interval time.Duration, condition ConditionFuncE) (instance *{{.Type}}, err error) {
	var done bool
	tick := time.Tick(interval)

	for {
		instance, err = core.kclient.{{ .Kubernetes.Plural }}({{ $ns }}).Get(ctx, name, metav1.GetOptions{})
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
// error is not nil{{ if .SupportsConditions }} or if the Status has a False condition.{{ end }}
func wrapPredicate(condition Predicate) ConditionFuncE {
	return func(obj *{{.Type}}, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		{{ if and .SupportsObservedGeneration .SupportsConditions }}
		if ObservedGenerationMatchesGeneration(obj) {
			for _, cond := range ExtractConditions(obj) {
				if cond.Status == corev1.ConditionFalse {
					return true, fmt.Errorf("Reason: %q, Message: %q", cond.Reason, cond.Message)
				}
			}
		}
		{{ else if .SupportsConditions }}
		for _, cond := range ExtractConditions(obj) {
			if cond.Status == corev1.ConditionFalse {
				return true, fmt.Errorf("Reason: %q, Message: %q", cond.Reason, cond.Message)
			}
		}
		{{ end }}

		return condition(obj), nil
	}
}

// WaitForDeletion is a utility function that combines waitForE with ConditionDeleted.
func (core *coreClient) WaitForDeletion(ctx context.Context, {{ $nssig }} name string, interval time.Duration) (instance *{{.Type}}, err error) {
	return core.waitForE(ctx, {{ $nsparam }} name, interval, ConditionDeleted)
}

{{ if .SupportsConditions }}
func checkConditionTrue(obj *{{.Type}}, err error, condition apis.ConditionType) (bool, error) {
	if err != nil {
		return true, err
	}

	{{ if .SupportsObservedGeneration }}// don't propagate old statuses
	if !ObservedGenerationMatchesGeneration(obj){
		return false, nil
	}
	{{ end }}
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

{{ range .Kubernetes.Conditions }}
// {{.PredicateName}} is a ConditionFuncE that waits for Condition{{.}} to
// become true and fails with an error if the condition becomes false.
func {{.PredicateName}}(obj *{{$type}}, err error) (bool, error) {
	return checkConditionTrue(obj, err, {{.ConditionName}})
}

// {{.WaitForName}} is a utility function that combines waitForE with {{.PredicateName}}.
func (core *coreClient) {{.WaitForName}}(ctx context.Context, {{ $nssig }} name string, interval time.Duration) (instance *{{$type}}, err error) {
	return core.waitForE(ctx, {{ $nsparam }} name, interval, {{.PredicateName}})
}
{{ end }}

{{ end }}

`))
