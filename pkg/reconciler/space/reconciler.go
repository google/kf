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
	"reflect"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/space/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles a Space object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	// listers index properties about resources
	spaceLister          kflisters.SpaceLister
	namespaceLister      v1listers.NamespaceLister
	roleLister           rbacv1listers.RoleLister
	resourceQuotaLister  v1listers.ResourceQuotaLister
	limitRangeLister     v1listers.LimitRangeLister
	serviceAccountLister v1listers.ServiceAccountLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by Kubernetes.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	original, err := r.spaceLister.Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Errorf("space %q no longer exists\n", name)
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
	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(toReconcile); uErr != nil {
		logger.Warnw("Failed to update Space status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the space.
func (r *Reconciler) ApplyChanges(ctx context.Context, space *v1alpha1.Space) error {
	logger := logging.FromContext(ctx)
	space.Status.InitializeConditions()
	namespaceName := resources.NamespaceName(space)

	// Sync Namespace
	{
		logger.Debug("reconciling Namespace")
		desired, err := resources.MakeNamespace(space)
		if err != nil {
			return err
		}

		actual, err := r.namespaceLister.Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Namespaces().Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkNamespaceNotOwned(namespaceName)
			return fmt.Errorf("space: %q does not own namespace: %q", space.Name, namespaceName)
		} else if actual, err = r.reconcileNs(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateNamespaceStatus(actual)
	}

	// If the namespace isn't ready (either it's coming up or being deleted)
	// then this reconciliation process can't continue until we get notified it is.
	if cond := space.Status.GetCondition(v1alpha1.SpaceConditionNamespaceReady); cond != nil {
		if !cond.IsTrue() {
			logger.Infof("can't continue reconciling until namespace %q is ready", namespaceName)
			return nil
		}
	}

	// Sync developer role
	{
		logger.Debug("reconciling develop Role")
		desired, err := resources.MakeDeveloperRole(space)
		if err != nil {
			return err
		}

		actual, err := r.roleLister.Roles(desired.Namespace).Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.RbacV1().Roles(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkDeveloperRoleNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own role: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileGenericRole(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateDeveloperRoleStatus(actual)
	}

	// Sync auditor role
	{
		logger.Debug("reconciling auditor Role")
		desired, err := resources.MakeAuditorRole(space)
		if err != nil {
			return err
		}

		actual, err := r.roleLister.Roles(desired.Namespace).Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.RbacV1().Roles(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkAuditorRoleNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own role: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileGenericRole(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateAuditorRoleStatus(actual)
	}

	// Sync resource quota
	{
		logger.Debug("reconciling ResourceQuota")
		desired, err := resources.MakeResourceQuota(space)
		if err != nil {
			return err
		}

		actual, err := r.resourceQuotaLister.ResourceQuotas(desired.Namespace).Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().ResourceQuotas(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkResourceQuotaNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own resourcequota: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileResourceQuota(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateResourceQuotaStatus(actual)
	}

	// Sync limit range
	{
		logger.Debug("reconciling LimitRange")
		desired, err := resources.MakeLimitRange(space)
		if err != nil {
			return err
		}

		actual, err := r.limitRangeLister.LimitRanges(desired.Namespace).Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().LimitRanges(desired.Namespace).Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkLimitRangeNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own limit range: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileLimitRange(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateLimitRangeStatus(actual)
	}

	// Sync build service account
	{
		logger.Debug("reconciling build service account")
		desired, err := resources.MakeBuildServiceAccount(space)
		if err != nil {
			return err
		}

		actual, err := r.serviceAccountLister.
			ServiceAccounts(desired.Namespace).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				CoreV1().ServiceAccounts(desired.Namespace).
				Create(desired)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !metav1.IsControlledBy(actual, space) {
			space.Status.MarkBuildServiceAccountNotOwned(desired.Name)
			return fmt.Errorf("space: %q does not own ServiceAccount: %q", space.Name, desired.Name)
		} else if actual, err = r.reconcileServiceAccount(ctx, desired, actual); err != nil {
			return err
		}

		space.Status.PropagateBuildServiceAccountStatus(actual)
	}

	return nil
}

func (r *Reconciler) reconcileNs(
	ctx context.Context,
	desired *v1.Namespace,
	actual *v1.Namespace,
) (*v1.Namespace, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff Namespace: %v", err)
	}
	logger.Debug("Namespace.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.Spec = desired.Spec
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	return r.KubeClientSet.CoreV1().Namespaces().Update(existing)
}

func (r *Reconciler) reconcileGenericRole(
	ctx context.Context,
	desired *rv1.Role,
	actual *rv1.Role,
) (*rv1.Role, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Rules, actual.Rules)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Rules, actual.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to diff Rules: %v", err)
	}
	logger.Debug("Role.Rules diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Rules = desired.Rules
	return r.KubeClientSet.RbacV1().Roles(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileResourceQuota(
	ctx context.Context,
	desired *v1.ResourceQuota,
	actual *v1.ResourceQuota,
) (*v1.ResourceQuota, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff Spec (ResourceList): %v", err)
	}
	logger.Debug("ResourceQuota.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.KubeClientSet.CoreV1().ResourceQuotas(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileLimitRange(
	ctx context.Context,
	desired *v1.LimitRange,
	actual *v1.LimitRange,
) (*v1.LimitRange, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff Spec (LimitRange): %v", err)
	}
	logger.Debug("LimitRange.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.KubeClientSet.CoreV1().LimitRanges(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileServiceAccount(
	ctx context.Context,
	desired *corev1.ServiceAccount,
	actual *corev1.ServiceAccount,
) (*corev1.ServiceAccount, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Labels, actual.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to diff ServiceAccounts Labels: %v", err)
	}
	logger.Debug("ServiceAccounts.Labels diff:", diff)

	diff, err = kmp.SafeDiff(desired.Secrets, actual.Secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to diff ServiceAccounts Secrets: %v", err)
	}
	logger.Debug("ServiceAccount.Secrets diff:", diff)

	diff, err = kmp.SafeDiff(desired.ImagePullSecrets, actual.ImagePullSecrets)
	if err != nil {
		return nil, fmt.Errorf("failed to diff ServiceAccounts ImagePullSecrets: %v", err)
	}
	logger.Debug("ServiceAccounts.ImagePullSecrets diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Secrets = desired.Secrets
	existing.ImagePullSecrets = desired.ImagePullSecrets
	return r.KubeClientSet.
		CoreV1().
		ServiceAccounts(existing.Namespace).
		Update(existing)
}

func (r *Reconciler) updateStatus(desired *v1alpha1.Space) (*v1alpha1.Space, error) {
	actual, err := r.spaceLister.Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()
	existing.Status = desired.Status

	return r.KfClientSet.KfV1alpha1().Spaces().UpdateStatus(existing)
}
