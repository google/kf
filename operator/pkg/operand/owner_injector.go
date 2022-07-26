// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operand

import (
	"context"
	"kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand/injection/dynamichelper"
	"strings"

	"github.com/google/go-cmp/cmp"
	multierror "github.com/hashicorp/go-multierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"
)

var (
	ignoreUID = cmp.FilterPath(func(p cmp.Path) bool {
		return strings.HasSuffix(p.String(), "Uid")
	}, cmp.Ignore())
)

type ownerInjector struct {
	dh dynamichelper.Interface
}

// OwnerInjector is able to inject the given Owner to an arbitrary
// set of objects referred to by LiveRefs.
type OwnerInjector interface {
	InjectOwnerRefs(context.Context, *metav1.OwnerReference, []v1alpha1.LiveRef) error
}

// CreateOwnerInjector creates an ownerInjector by extracting the dynamicclient
// from the context given.
func CreateOwnerInjector(ctx context.Context) OwnerInjector {
	return &ownerInjector{
		dh: dynamichelper.Get(ctx),
	}
}

// InjectOwnerRefs injects the given OwnerRefence into all objects passed in as a LiveRef.
func (o ownerInjector) InjectOwnerRefs(ctx context.Context, ownerRef *metav1.OwnerReference, live []v1alpha1.LiveRef) error {
	// HACK, we don't want to be a controller here.
	*ownerRef.Controller = false
	var result *multierror.Error
	for _, ref := range live {
		lo := ownerRef
		if shouldSkipOwnerRef(ref) {
			lo = nil
		}
		err := o.injectOwnerRef(ctx, ownerRef.Name, lo, ref)
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

func (o ownerInjector) deriveGVR(ref v1alpha1.LiveRef) (*schema.GroupVersionResource, error) {
	if ref.GroupVersionResource() != nil {
		return ref.GroupVersionResource(), nil
	}
	m, err := o.dh.RESTMapping(*ref.GroupKind())
	if err != nil {
		return nil, err
	}
	return &m.Resource, nil

}

func shouldSkipOwnerRef(ref v1alpha1.LiveRef) bool {
	// skip setting owner refs on namespaces
	if ref.Group == "" && (ref.Kind == "Namespace" || ref.Resource == "namespaces") {
		return true
	}
	// skip setting owner refs CRDs. We want CRDs to only persist.
	if ref.Group == "apiextensions.k8s.io" && ref.Kind == "CustomResourceDefinition" {
		return true
	}
	return false
}

func (o ownerInjector) desiredOwnerRefs(unstr *unstructured.Unstructured, ownerRef *metav1.OwnerReference) ([]metav1.OwnerReference, bool) {
	dr := []metav1.OwnerReference{}
	modified := false
	if ownerRef == nil {
		if len(unstr.GetOwnerReferences()) == 0 {
			return dr, modified
		}
		// Filter out things we have added in the past, leave others untouched.
		for _, ref := range unstr.GetOwnerReferences() {
			if ref.Kind != "ClusterActiveOperand" && ref.Kind != "ActiveOperand" {
				dr = append(dr, ref)
			} else {
				modified = true
			}
		}
	} else {
		refs := unstr.GetOwnerReferences()
		for _, ref := range refs {
			if cmp.Equal(ref, *ownerRef, ignoreUID) {
				return unstr.GetOwnerReferences(), false
			}
		}
		dr = append(unstr.GetOwnerReferences(), *ownerRef)
		modified = true
	}
	return dr, modified
}

func (o ownerInjector) injectOwnerRef(ctx context.Context, name string, ownerRef *metav1.OwnerReference, ref v1alpha1.LiveRef) error {
	gvr, err := o.deriveGVR(ref)
	if err != nil {
		return err
	}

	var r dynamic.ResourceInterface = o.dh.Resource(*gvr)
	if ref.Namespace != "" {
		r = r.(dynamic.NamespaceableResourceInterface).Namespace(ref.Namespace)
	}
	unstr, err := r.Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	desired, changed := o.desiredOwnerRefs(unstr, ownerRef)
	if !changed {
		return nil
	}
	unstr.SetOwnerReferences(desired)
	_, err = r.Update(ctx, unstr, metav1.UpdateOptions{FieldManager: name})
	return err
}
