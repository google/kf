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

package space

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/GoogleCloudPlatform/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/GoogleCloudPlatform/kf/pkg/reconciler/space/resources"
	v1 "k8s.io/api/core/v1"
	rv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"

	"github.com/knative/pkg/kmp"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/client-go/tools/cache"
)

type Reconciler struct {
	corev1      corev1.CoreV1Interface
	rolesv1     rbacv1.RolesGetter
	spaceLister kflisters.SpaceLister
}

func NewReconciler(corev1 corev1.CoreV1Interface, rolesv1 rbacv1.RolesGetter, spaceLister kflisters.SpaceLister) {

}

// Reconcile is called by Kubernetes.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := log.New(os.Stderr, "reconciler > ", 0)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	original, err := r.spaceLister.Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Printf("space %q no longer exists\n", name)
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		return nil
	}

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(toReconcile)

	// TODO: actually sync back to the server

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the space.
func (r *Reconciler) ApplyChanges(space *v1alpha1.Space) error {
	space.Status.InitializeConditions()
	namespaceName := resources.NamespaceName(space)

	// Sync Namespace
	{
		desired, err := resources.MakeNamespace(space)
		if err != nil {
			return err
		}

		actual, err := r.corev1.Namespaces().Get(desired.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			actual, err = r.corev1.Namespaces().Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(desired, space) {
			space.Status.MarkNamespaceNotOwned(namespaceName)
			return fmt.Errorf("space: %q does not own namespace: %q", space.Name, namespaceName)
		} else if actual, err = r.reconcileNs(desired, actual); err != nil {
			return err
		}

		space.Status.PropagateNamespaceStatus(actual)
	}

	// Sync developer role
	{
		desired, err := resources.MakeDeveloperRole(space)
		if err != nil {
			return err
		}

		actual, err := r.rolesv1.Roles(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			actual, err = r.rolesv1.Roles(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkDeveloperRoleNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own role: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileGenericRole(desired, actual); err != nil {
			return err
		}

		space.Status.PropagateDeveloperRoleStatus(actual)
	}

	// Sync auditor role
	{
		desired, err := resources.MakeAuditorRole(space)
		if err != nil {
			return err
		}

		actual, err := r.rolesv1.Roles(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			actual, err = r.rolesv1.Roles(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkAuditorRoleNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own role: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileGenericRole(desired, actual); err != nil {
			return err
		}

		space.Status.PropagateAuditorRoleStatus(actual)
	}

	return nil
}

func (r *Reconciler) reconcileNs(desired, actual *v1.Namespace) (*v1.Namespace, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff Namespace: %v", err)
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.Spec = desired.Spec
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	return r.corev1.Namespaces().Update(existing)
}

func (r *Reconciler) reconcileGenericRole(desired, actual *rv1.Role) (*rv1.Role, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Rules, actual.Rules)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Rules, actual.Rules); err != nil {
		return nil, fmt.Errorf("failed to diff Rules: %v", err)
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Rules = desired.Rules
	return r.rolesv1.Roles(existing.Namespace).Update(existing)
}
