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
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/apis/networking"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	networkingv1listers "github.com/google/kf/v2/pkg/client/kube/listers/networking/v1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"github.com/google/kf/v2/pkg/reconciler/space/resources"
	"github.com/google/kf/v2/pkg/system"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	// WorkloadIdentityRole is the role that Workload Identity uses.
	WorkloadIdentityRole = "roles/iam.workloadIdentityUser"

	// workloadIdentityFinalizer is the name of the finalizer used to remove WI
	// from a space before it's completely removed
	workloadIdentityFinalizer = "workloadidentity.kf.dev"
)

// Reconciler reconciles a Space object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	// listers index properties about resources
	spaceLister              kflisters.SpaceLister
	namespaceLister          v1listers.NamespaceLister
	serviceAccountLister     v1listers.ServiceAccountLister
	configStore              *config.Store
	serviceLister            v1listers.ServiceLister
	kfConfigStore            *kfconfig.Store
	networkPolicyLister      networkingv1listers.NetworkPolicyLister
	deploymentLister         appsv1listers.DeploymentLister
	roleBindingLister        rbacv1listers.RoleBindingLister
	roleLister               rbacv1listers.RoleLister
	clusterRoleLister        rbacv1listers.ClusterRoleLister
	clusterRoleBindingLister rbacv1listers.ClusterRoleBindingLister
	gsaPolicyLister          cache.GenericLister
	configMapLister          v1listers.ConfigMapLister

	iamClientSet dynamic.NamespaceableResourceInterface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of the
// watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	ctx = r.configStore.ToContext(ctx)
	ctx = r.kfConfigStore.ToContext(ctx)

	// Most operations in the reconciler don't involve the network and
	// therefore won't require a context.
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	original, err := r.spaceLister.Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")

		// We always have to update the IAM policies to ensure the proper list
		// of Spaces are configured by the policy.
		if err := r.updateGSAPolicies(ctx, original); err != nil {
			logger.Warnw("failed to update GSA poliy", zap.Error(err))
			return err
		}

		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.PropagateTerminatingStatus()
		if err := r.removeFinalizer(ctx, toUpdate); err != nil {
			return err
		}

		// We always have to update the IAM policies to ensure the proper list
		// of Spaces are configured by the policy.
		if err := r.updateGSAPolicies(ctx, original); err != nil {
			logger.Warnw("failed to update GSA poliy", zap.Error(err))
			return err
		}

		if _, err := r.updateStatus(ctx, toUpdate); err != nil {
			logger.Warnw("failed to update Space status", zap.Error(err))
			return err
		}
		return nil
	}

	// XXX: We have to always remove the finalizer now so that for the next
	// release, we know that Spaces don't have finalizers. Finalizers were
	// added before to try and prevent the object from being deleted UNTIL the
	// GCP IAM resource were cleaned up. This responsibility now falls on KCC.
	// However, it is possible (and likely) that existing CRs (from previous
	// releases) have the finalizer already and therefore MUST be removed.
	r.removeFinalizer(ctx, original)

	// Don't modify the informers copy
	toReconcile := original.DeepCopy()

	// ALWAYS update the ObservedGenration: "If the primary resource your
	// controller is reconciling supports ObservedGeneration in its status, make
	// sure you correctly set it to metadata.Generation whenever the values
	// between the two fields mismatches."
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
	toReconcile.Status.ObservedGeneration = toReconcile.Generation

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if reconcileErr != nil {
		logger.Debugf("Space reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("failed to update Space status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Space.
func (r *Reconciler) ApplyChanges(ctx context.Context, space *v1alpha1.Space) error {
	logger := logging.FromContext(ctx)
	space.Status.InitializeConditions()
	namespaceName := resources.NamespaceName(space)

	// Update the cluster ingresses based on the services. These values don't
	// require any action from this controller and are used in debugging so they
	// should be updated regardless of the state of the overall space.
	{
		logger.Debug("updating ingress gateways")
		condition := space.Status.IngressGatewayCondition()
		ingresses, err := system.GetClusterIngresses(r.serviceLister)
		if err != nil {
			return condition.MarkReconciliationError("fetching services", err)
		}

		space.Status.PropagateIngressGatewayStatus(ingresses)
	}

	// Update kf runtime info based on the spec and override configmaps. These
	// values don't require any action from this controller and are used by kf so
	// they should be updated regardless of the state of the overall Space.
	{
		logger.Debug("updating application runtime config")
		space.Status.PropagateRuntimeConfigStatus(space.Spec.RuntimeConfig)
	}

	{
		logger.Debug("updating network config")
		space.Status.PropagateNetworkConfigStatus(space.Spec.NetworkConfig, kfconfig.FromContext(ctx), space.Name)
	}

	{
		logger.Debug("updating build runtime config")
		space.Status.PropagateBuildConfigStatus(space.Spec, kfconfig.FromContext(ctx))
	}

	// Sync Namespace
	{
		logger.Debug("reconciling Namespace")
		condition := space.Status.NamespaceCondition()

		asmRev, err := networking.LookupASMRev(func(namespace string, selector labels.Selector) ([]*appsv1.Deployment, error) {
			return r.deploymentLister.Deployments(namespace).List(selector)
		}, func(namespace, configMapName string) (*v1.ConfigMap, error) {
			return r.configMapLister.ConfigMaps(namespace).Get(configMapName)
		})
		if err != nil {
			return condition.MarkReconciliationError("fetching ASM revision", err)
		}

		desired, err := resources.MakeNamespace(space, asmRev)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.namespaceLister.Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Namespaces().Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(namespaceName)
		} else if actual, err = r.reconcileNs(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
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

	// Sync build service account and secret
	secretCondition := space.Status.BuildSecretCondition()

	// Fetch all the secrets from the Kf namespace. These will be copied to
	// the Space.
	secretsConfig, err := config.FromContext(ctx).Secrets()
	if err != nil {
		return secretCondition.MarkTemplateError(err)
	}

	// Fetch container registry from config-defaults.
	// Build secrets need to be properly annotated with host URL for Tekton.
	defaultsConfig, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		return secretCondition.MarkTemplateError(err)
	}

	var kfSecrets []*corev1.Secret
	for _, secret := range secretsConfig.BuildImagePushSecrets {
		if secret.Name == "" {
			continue
		}

		// Get name of kf secret that will be copied into the Space.
		kfSecret, err := r.SecretLister.
			Secrets(v1alpha1.KfNamespace).
			Get(secret.Name)
		if err != nil {
			return secretCondition.MarkTemplateError(err)
		}
		kfSecrets = append(kfSecrets, kfSecret)
	}

	desiredServiceAccount, desiredSecrets, err := resources.MakeBuildServiceAccount(
		space,
		kfSecrets,
		secretsConfig.GoogleServiceAccount,
		defaultsConfig.SpaceContainerRegistry,
	)
	if err != nil {
		return secretCondition.MarkTemplateError(err)
	}

	// Sync build secrets
	for _, desiredSecret := range desiredSecrets {
		logger.Debug("reconciling build secret")

		actual, err := r.SecretLister.
			Secrets(desiredSecret.Namespace).
			Get(desiredSecret.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				CoreV1().
				Secrets(desiredSecret.Namespace).
				Create(ctx, desiredSecret, metav1.CreateOptions{})
			if err != nil {
				return secretCondition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return secretCondition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return secretCondition.MarkChildNotOwned(desiredSecret.Name)
		} else if actual, err = r.ReconcileSecret(ctx, desiredSecret, actual); err != nil {
			return secretCondition.MarkReconciliationError("updating existing", err)
		}

		// Secrets don't have any data necessary to propagate.
		secretCondition.MarkSuccess()
	}

	if len(desiredSecrets) == 0 {
		// We need to mark it successful so the Space's status isn't left
		// unknown.
		secretCondition.MarkSuccess()
	}

	// Sync build service account
	{
		logger.Debug("reconciling build service account")
		condition := space.Status.BuildServiceAccountCondition()

		actual, err := r.serviceAccountLister.
			ServiceAccounts(desiredServiceAccount.Namespace).
			Get(desiredServiceAccount.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				CoreV1().ServiceAccounts(desiredServiceAccount.Namespace).
				Create(ctx, desiredServiceAccount, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return secretCondition.MarkChildNotOwned(desiredServiceAccount.Name)
		} else if actual, err = r.ReconcileServiceAccount(ctx, desiredServiceAccount, actual, true); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		// Service accounts don't have any data necessary to propagate.
		condition.MarkSuccess()
	}

	// Sync source builder Role
	{
		logger.Debug("reconciling build role")
		condition := space.Status.BuildRoleCondition()
		desired := resources.MakeSourceBuilderRole(space)

		actual, err := r.roleLister.
			Roles(desired.Namespace).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				RbacV1().
				Roles(desired.Namespace).
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileRole(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		// Roles don't have any data necessary to propagate.
		condition.MarkSuccess()
	}

	// Sync source builder RoleBinding
	{
		logger.Debug("reconciling build role binding")
		condition := space.Status.BuildRoleBindingCondition()
		desired := resources.MakeSourceBuilderRoleBinding(space)

		actual, err := r.roleBindingLister.
			RoleBindings(desired.Namespace).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			actual, err = r.KubeClientSet.
				RbacV1().
				RoleBindings(desired.Namespace).
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileRoleBinding(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		// Role doesn't have any data necessary to propagate.
		condition.MarkSuccess()
	}

	{
		logger.Debug("reconciling app NetworkPolicy")
		condition := space.Status.AppNetworkPolicyCondition()

		desired, err := resources.MakeAppNetworkPolicy(space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.networkPolicyLister.
			NetworkPolicies(desired.Namespace).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			_, err = r.KubeClientSet.
				NetworkingV1().
				NetworkPolicies(desired.Namespace).
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if _, err = r.reconcileNetworkPolicy(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		// RoleBinding doesn't have any data necessary to propagate.
		condition.MarkSuccess()
	}

	{
		logger.Debug("reconciling build NetworkPolicy")
		condition := space.Status.BuildNetworkPolicyCondition()

		desired, err := resources.MakeBuildNetworkPolicy(space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.networkPolicyLister.
			NetworkPolicies(desired.Namespace).
			Get(desired.Name)
		if errors.IsNotFound(err) {
			_, err = r.KubeClientSet.
				NetworkingV1().
				NetworkPolicies(desired.Namespace).
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if _, err = r.reconcileNetworkPolicy(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		// NetworkPolicies don't have any data necessary to propagate.
		condition.MarkSuccess()
	}

	{
		logger.Debug("reconciling RoleBindings")
		condition := space.Status.RoleBindingsCondition()

		for _, roleName := range resources.AllRoleNames() {
			desired := resources.MakeRoleBindingForClusterRole(space, roleName)

			actual, err := r.roleBindingLister.
				RoleBindings(desired.Namespace).
				Get(desired.Name)
			if errors.IsNotFound(err) {
				_, err = r.KubeClientSet.
					RbacV1().
					RoleBindings(desired.Namespace).
					Create(ctx, desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if !metav1.IsControlledBy(actual, space) {
				return condition.MarkChildNotOwned(desired.Name)
			}
			// No reconciliation is necessary, Kf just needs to ensure the resource
			// exists so users can manage it.
		}

		condition.MarkSuccess()
	}

	{
		logger.Debug("reconciling ClusterRole")
		condition := space.Status.ClusterRoleCondition()

		desired := resources.MakeSpaceManagerClusterRole(space)
		actual, err := r.clusterRoleLister.Get(desired.Name)
		if errors.IsNotFound(err) {
			_, err = r.KubeClientSet.
				RbacV1().
				ClusterRoles().
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, space) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if _, err := r.reconcileClusterRole(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		condition.MarkSuccess()
	}

	// Create ClusterRoleBindings
	{
		logger.Debug("reconciling ClusterRoleBindings")
		condition := space.Status.ClusterRoleBindingsCondition()

		spaceBindings, err := r.roleBindingLister.
			RoleBindings(space.Name).
			List(labels.Everything())

		if err != nil {
			return condition.MarkReconciliationError("getting RoleBindings in Space", err)
		}

		desiredBindings := []*rbacv1.ClusterRoleBinding{
			resources.MakeClusterRoleBinding(
				space,
				resources.ClusterRoleName(space),
				resources.FilterSubjectsByClusterRole([]resources.RoleName{resources.SpaceManager}, spaceBindings),
			),
			resources.MakeClusterRoleBinding(
				space,
				resources.ClusterReaderRole,
				resources.FilterSubjectsByClusterRole(resources.AllRoleNames(), spaceBindings),
			),
		}

		for _, desired := range desiredBindings {
			actual, err := r.clusterRoleBindingLister.
				Get(desired.Name)
			if errors.IsNotFound(err) {
				_, err = r.KubeClientSet.
					RbacV1().
					ClusterRoleBindings().
					Create(ctx, desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if !metav1.IsControlledBy(actual, space) {
				return condition.MarkChildNotOwned(desired.Name)
			} else if _, err = r.reconcileClusterRoleBinding(ctx, desired, actual); err != nil {
				return condition.MarkReconciliationError("synchronizing", err)
			}
		}

		condition.MarkSuccess()
	}

	// Update IAMPolicies
	// WI is not enabled if a GSA is not provided.
	if err := r.updateGSAPolicies(ctx, space); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) updateGSAPolicies(
	ctx context.Context,
	space *v1alpha1.Space,
) error {
	logger := logging.FromContext(ctx)

	secretsConfig, err := config.FromContext(ctx).Secrets()
	if err != nil {
		return err
	}

	defaultsConfig, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		return err
	}

	// When a Space has been deleted, the Space will be nil. We won't need to
	// update the Status anyways, just give it a dummy Space. We HAVE to
	// always update (even when we deleted the Space) to keep the policies in
	// sync. Therefore, proceed as if we had a Space.
	if space == nil {
		space = new(v1alpha1.Space)
	}

	// IAMPolicy need both fields and these fields are immutable after
	// creation and aren't necessary when AppDevExperience Builds are being
	// used.
	if secretsConfig.GoogleServiceAccount != "" && secretsConfig.GoogleProjectID != "" && defaultsConfig.FeatureFlags.AppDevExperienceBuilds().IsDisabled() {
		logger.Debug("reconciling IAMPolicies")
		condition := space.Status.IAMPolicyCondition()

		// Get all the Spaces that should have the associated IAM policy.
		spaces, err := r.spaceLister.List(labels.Everything())
		if err != nil {
			return condition.MarkReconciliationError("getting Spaces", err)
		}

		desired, err := resources.MakeIAMPolicy(
			secretsConfig.GoogleProjectID,
			secretsConfig.GoogleServiceAccount,
			spaces,
		)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actualRuntimeObj, err := r.
			gsaPolicyLister.
			ByNamespace(v1alpha1.KfNamespace).
			Get(desired.GetName())
		actual, _ := actualRuntimeObj.(*unstructured.Unstructured)
		if errors.IsNotFound(err) {
			// Not found, create it.
			actual, err = r.iamClientSet.
				Namespace(v1alpha1.KfNamespace).
				Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if actual, err = r.reconcileIAMPolicy(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("synchronizing", err)
		}

		space.Status.PropagateIAMPolicyStatus(ctx, actual)
	} else {
		// WI is not enabled, so just mark everying as good.
		space.Status.IAMPolicyCondition().MarkSuccess()
	}

	return nil
}

func (r *Reconciler) reconcileClusterRoleBinding(ctx context.Context, desired, actual *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "ClusterRoleBinding").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("subjects", desired.Subjects, actual.Subjects).
		Append("roleRef", desired.RoleRef, actual.RoleRef).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object.
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Subjects = desired.Subjects
	existing.RoleRef = desired.RoleRef
	return r.KubeClientSet.RbacV1().
		ClusterRoleBindings().
		Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) reconcileClusterRole(ctx context.Context, desired, actual *rbacv1.ClusterRole) (*rbacv1.ClusterRole, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "ClusterRole").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("rules", desired.Rules, actual.Rules).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object.
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Rules = desired.Rules
	return r.KubeClientSet.RbacV1().ClusterRoles().Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) reconcileNetworkPolicy(ctx context.Context, desired, actual *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "NetworkPolicy").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("spec", desired.Spec, actual.Labels).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object.
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.KubeClientSet.NetworkingV1().NetworkPolicies(existing.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) reconcileIAMPolicy(
	ctx context.Context,
	desired *unstructured.Unstructured,
	actual *unstructured.Unstructured,
) (*unstructured.Unstructured, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	builder := reconciler.NewUnstructuredSemanticEqualityBuilder(logger, "IAMPolicy").
		Append("metadata.labels", desired, actual).
		Append("spec", desired, actual)

	if builder.IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	builder.Transform(existing)
	return r.iamClientSet.
		Namespace(v1alpha1.KfNamespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) reconcileNs(
	ctx context.Context,
	desired *v1.Namespace,
	actual *v1.Namespace,
) (*v1.Namespace, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	// NOTE: We don't compare the Namespace's specs as they only contain
	// Finalizers, which we don't plan to add any. If this ever changes, we'll
	// need to ensure the desired Finalizers are in place while ignoring the
	// existing ones.
	if reconciler.NewSemanticEqualityBuilder(logger, "Namespace").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	return r.KubeClientSet.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.Space) (*v1alpha1.Space, error) {
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

	return r.KfClientSet.KfV1alpha1().Spaces().UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}

func builderGSA(secretsConfig *config.SecretsConfig, spaceName string) string {
	return fmt.Sprintf(
		"serviceAccount:%s.svc.id.goog[%s/%s]",
		secretsConfig.GoogleProjectID,
		spaceName,
		v1alpha1.DefaultBuildServiceAccountName,
	)
}

func (r *Reconciler) removeFinalizer(ctx context.Context, s *v1alpha1.Space) error {
	logger := logging.FromContext(ctx)
	if !reconcilerutil.HasFinalizer(s, workloadIdentityFinalizer) {
		return nil
	}

	toUpdate := s.DeepCopy()

	// XXX: We have to ALWAYS remove the finalizer for upgrade
	// purposes. If a user were to upgrade and have existing Spaces,
	// then we should remove them. They were here before we used KCC
	// to manage the IAM policies.
	// NOTE: This can be deleted after v2.5.0 is released.
	reconcilerutil.RemoveFinalizer(toUpdate, workloadIdentityFinalizer)
	// Update finalizer and status
	if _, err := r.
		KfClientSet.
		KfV1alpha1().
		Spaces().
		Update(ctx, toUpdate, metav1.UpdateOptions{}); err != nil {
		logger.Warnw("failed to remove Workload Identity finalizer", zap.Error(err))
		return err
	}

	return nil
}
