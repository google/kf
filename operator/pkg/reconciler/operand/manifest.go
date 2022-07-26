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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-multierror"
	mf "github.com/manifestival/manifestival"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"knative.dev/pkg/logging"

	poperand "kf-operator/pkg/operand"
	"kf-operator/pkg/operand/injection/dynamichelper"
)

const (
	mode = "operator.knative.dev/mode"
)

var (
	configmaps           = mf.ByKind("ConfigMap")
	namespace            = mf.ByKind("Namespace")
	role                 = mf.Any(mf.ByKind("Role"), mf.ByKind("ClusterRole"))
	rolebinding          = mf.Any(mf.ByKind("RoleBinding"), mf.ByKind("ClusterRoleBinding"))
	webhookConfiguration = mf.Any(mf.ByKind("ValidatingWebhookConfiguration"), mf.ByKind("MutatingWebhookConfiguration"))
	deployment           = mf.ByKind("Deployment")
	applicationOrder     = []mf.Predicate{
		mf.CRDs,
		namespace,
		// As an aggregated ClusterRole exists that should grant the Operator any roles created
		// here, we install roles and rolbindings first. This also prevents us from binding any
		// non-existant roles, which would require cluster-admin levels of access to the cluster.
		role,
		rolebinding,
		configmaps,
		// The admission webhooks reject the creation of ksvc and subresources by default until
		// the webhook service and deployment are ready.
		webhookConfiguration,
		// Exclude all other types so that delete is also deterministic.
		mf.Not(mf.Any(namespace, mf.CRDs, role, rolebinding, configmaps, webhookConfiguration, deployment)),
		// Deployments are installed last so that all necessary config is installed, and they
		// are uninstalled first, because this is the most obvious change to the customer.
		deployment,
	}
)

type manifestReconciler struct {
	dh dynamichelper.Interface
}

// NewManifestReconciler returns a ResourceReconciler which reconciles
// resources using the provided manifestival client.
func NewManifestReconciler(dh dynamichelper.Interface) poperand.ResourceReconciler {
	return manifestReconciler{dh}
}

// ApplyManifest applies a Manifest on the cluster.
func (r manifestReconciler) Apply(ctx context.Context, resources []unstructured.Unstructured) error {
	manifest, err := mf.ManifestFrom((mf.Slice)(resources))
	if err != nil {
		return err
	}
	var result *multierror.Error
	for _, filter := range applicationOrder {
		if err := r.partialApply(ctx, manifest.Filter(filter).Resources()); err != nil {
			logging.FromContext(ctx).Warn("Failed to partiallyApply: ", err)
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

// Appendix method, can be removed / merged with healthcheck.go
func (r manifestReconciler) GetState(ctx context.Context, resources []unstructured.Unstructured) (string, error) {
	return poperand.Installed, nil
}

func (r manifestReconciler) partialApply(ctx context.Context, resources []unstructured.Unstructured) error {
	var result *multierror.Error
	for _, resource := range resources {
		gvk := resource.GroupVersionKind()
		c, err := r.dh.Lookup(gvk.GroupKind(), resource.GetNamespace(), gvk.Version)
		if err != nil {
			result = multierror.Append(result, resourceError(err, &resource))
			continue
		}
		label, exists := resource.GetAnnotations()[mode]
		if !exists {
			// TODO: Remove defaulting and fail here.
			label = "Reconcile"
		}
		var a map[string]string
		if a = resource.GetAnnotations(); a == nil {
			a = make(map[string]string)
		}
		a[mode] = label
		resource.SetAnnotations(a)
		_, err = c.Get(ctx, resource.GetName(), metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			_, err = c.Create(ctx, &resource, metav1.CreateOptions{FieldManager: "kuberun-operator"})
			result = multierror.Append(result, resourceError(err, &resource))
			continue
		} else if err != nil {
			result = multierror.Append(result, resourceError(err, &resource))
			continue
		}

		if label == "EnsureExists" {
			continue
		}
		b, err := json.Marshal(resource.Object)
		if err != nil {
			result = multierror.Append(result, resourceError(err, &resource))
			continue
		}
		var po metav1.PatchOptions
		po.Force = pointer.BoolPtr(true)
		po.FieldManager = "kuberun-operator"
		if _, err := c.Patch(ctx, resource.GetName(), types.ApplyPatchType, []byte(b), po); err != nil {
			result = multierror.Append(result, resourceError(err, &resource))
			continue
		}

	}
	return result.ErrorOrNil()
}

func resourceError(err error, spec *unstructured.Unstructured) error {
	if err == nil {
		return nil
	}
	return errors.Wrapf(err, "Failed resource %s", resourceInfo(spec))
}

func resourceInfo(spec *unstructured.Unstructured) string {
	return fmt.Sprintf("%v: %v/%v", spec.GetKind(), spec.GetNamespace(), spec.GetName())
}
