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

package reconciler

import (
	"context"
	"fmt"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclientset "github.com/google/kf/v2/pkg/client/kf/clientset/versioned"
	kfscheme "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/scheme"
	kfclient "github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/algorithms"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	podconvert "github.com/tektoncd/pipeline/pkg/pod"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1listers "k8s.io/client-go/listers/core/v1"
	"knative.dev/pkg/apis"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	namespaceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/namespace"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
)

// Base implements the core controller logic, given a Reconciler.
type Base struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet kubernetes.Interface

	// KfClientSet allows us to configure Kf objects
	KfClientSet kfclientset.Interface

	// ConfigMapWatcher allows us to watch for ConfigMap changes.
	ConfigMapWatcher configmap.Watcher

	// NamespaceLister allows us to list Namespaces. We use this to check for
	// terminating namespaces.
	NamespaceLister v1listers.NamespaceLister

	// SecretLister allows us to list Secrets.
	SecretLister v1listers.SecretLister

	// SecretInformer allows us to AddEventHandlers for Secrets.
	SecretInformer informerscorev1.SecretInformer
}

// NewBase instantiates a new instance of Base implementing
// the common & boilerplate code between our reconcilers.
func NewBase(ctx context.Context, cmw configmap.Watcher) *Base {
	kubeClient := kubeclient.Get(ctx)
	nsInformer := namespaceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	base := &Base{
		KubeClientSet:    kubeClient,
		KfClientSet:      kfclient.Get(ctx),
		ConfigMapWatcher: cmw,
		SecretLister:     secretInformer.Lister(),
		SecretInformer:   secretInformer,

		NamespaceLister: nsInformer.Lister(),
	}

	return base
}

func NewControllerLogger(ctx context.Context, resource string) *zap.SugaredLogger {
	return logging.FromContext(ctx).
		Named(resource).
		With(logkey.ControllerType, resource)
}

// IsNamespaceTerminating returns true if the namespace is marked as terminating
// and false if the state is unknown or not terminating.
func (base *Base) IsNamespaceTerminating(namespace string) bool {
	ns, err := base.NamespaceLister.Get(namespace)
	if err != nil || ns == nil {
		return false
	}

	return ns.Status.Phase == corev1.NamespaceTerminating
}

// ReconcileRole syncs the existing K8s Role to the desired Role.
func (b *Base) ReconcileRole(
	ctx context.Context,
	desired *rbacv1.Role,
	actual *rbacv1.Role,
) (*rbacv1.Role, error) {
	logger := logging.FromContext(ctx)

	if NewSemanticEqualityBuilder(logger, "Role").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("rules", desired.Rules, actual.Rules).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Rules = desired.Rules
	return b.KubeClientSet.
		RbacV1().
		Roles(existing.Namespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

// ReconcileRoleBinding syncs the existing K8s RoleBinding to the desired RoleBinding.
func (b *Base) ReconcileRoleBinding(
	ctx context.Context,
	desired *rbacv1.RoleBinding,
	actual *rbacv1.RoleBinding,
) (*rbacv1.RoleBinding, error) {
	logger := logging.FromContext(ctx)

	if NewSemanticEqualityBuilder(logger, "RoleBinding").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("subjects", desired.Subjects, actual.Subjects).
		Append("roleRef", desired.RoleRef, actual.RoleRef).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Subjects = desired.Subjects
	existing.RoleRef = desired.RoleRef
	return b.KubeClientSet.
		RbacV1().
		RoleBindings(existing.Namespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

// ReconcileSecret syncs the existing K8s Secret to the desired Secret.
func (b *Base) ReconcileSecret(
	ctx context.Context,
	desired *corev1.Secret,
	actual *corev1.Secret,
) (*corev1.Secret, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if NewSemanticEqualityBuilder(logger, "Secret").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("data", desired.Data, actual.Data).
		Append("type", desired.Type, actual.Type).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Data = desired.Data
	existing.Type = desired.Type
	return b.KubeClientSet.
		CoreV1().
		Secrets(existing.Namespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

// ReconcileServiceAccount syncs the actual service account to the desired SA. The reconciliation
// merges the secrets on the service accounts rather than replacing them. This is because the k8s cluster creates
// a default token secret per SA after the SA is first created.
func (b *Base) ReconcileServiceAccount(
	ctx context.Context,
	desired *corev1.ServiceAccount,
	actual *corev1.ServiceAccount,
	checkForWI bool,
) (*corev1.ServiceAccount, error) {
	logger := logging.FromContext(ctx)

	// Merge the secrets and image pull secrets so the default K8s authentication
	// token can be retained.
	desired.Secrets = algorithms.Merge(
		v1alpha1.ObjectReferences(desired.Secrets),
		v1alpha1.ObjectReferences(actual.Secrets),
	).(v1alpha1.ObjectReferences)
	desired.ImagePullSecrets = algorithms.Merge(
		v1alpha1.LocalObjectReferences(desired.ImagePullSecrets),
		v1alpha1.LocalObjectReferences(actual.ImagePullSecrets),
	).(v1alpha1.LocalObjectReferences)

	// Check for differences, if none we don't need to reconcile.
	builder := NewSemanticEqualityBuilder(logger, "Deployment").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("secrets", desired.Secrets, actual.Secrets).
		Append("imagePullSecrets", desired.ImagePullSecrets, actual.ImagePullSecrets)

	if checkForWI {
		// The only annotation we want to make sure is there is the Workload
		// Identity one.
		builder.Append(
			fmt.Sprintf("metadata.annotations[%q]", v1alpha1.WorkloadIdentityAnnotation),
			desired.ObjectMeta.Annotations[v1alpha1.WorkloadIdentityAnnotation],
			actual.ObjectMeta.Annotations[v1alpha1.WorkloadIdentityAnnotation],
		)
	}

	if builder.IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.ObjectMeta.Annotations = desired.ObjectMeta.Annotations
	existing.Secrets = desired.Secrets
	existing.ImagePullSecrets = desired.ImagePullSecrets
	return b.KubeClientSet.
		CoreV1().
		ServiceAccounts(existing.Namespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

// ReconcileDeployment syncs the existing K8s Deployment to the desired Deployment.
func (b *Base) ReconcileDeployment(ctx context.Context, desired, actual *appsv1.Deployment) (*appsv1.Deployment, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if NewSemanticEqualityBuilder(logger, "Deployment").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("spec", desired.Spec, actual.Spec).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return b.KubeClientSet.AppsV1().Deployments(existing.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
}

// ReconcileService syncs the existing K8s Service to the desired Service.
func (b *Base) ReconcileService(ctx context.Context, desired, actual *corev1.Service) (*corev1.Service, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if NewSemanticEqualityBuilder(logger, "Service").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("spec.ports", desired.Spec.Ports, actual.Spec.Ports).
		Append("spec.selector", desired.Spec.Selector, actual.Spec.Selector).
		Append("spec.type", desired.Spec.Type, actual.Spec.Type).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec.Ports = desired.Spec.Ports
	existing.Spec.Selector = desired.Spec.Selector
	existing.Spec.Type = desired.Spec.Type

	return b.KubeClientSet.CoreV1().Services(existing.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
}

// CleanupCompletedTaskRunSidecars removes sidecars if the TaskRun is completed.
func (b *Base) CleanupCompletedTaskRunSidecars(ctx context.Context, tr *tektonv1beta1.TaskRun) error {
	logger := logging.FromContext(ctx)

	configDefaults, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		logger.Errorf("Couldn't get configDefaults %v", err)
		return err
	}

	// Don't continue if the TaskRun was canceled or timed out because the Pod won't exist.
	condition := tr.Status.GetCondition(apis.ConditionSucceeded)
	switch {
	case condition == nil || condition.IsUnknown():
		// TaskRun not started or is in a pending state.
		return nil

	case condition.IsFalse() || condition.IsTrue():
		if tr.Status.PodName == "" {
			logger.Infof("Couldn't stop sidecars for TaskRun %s/%s, no Pod", tr.Namespace, tr.Name, err)
			return nil
		}

		// Task has terminated, continue.
		break
	default:
		// Unknown state, bail.
		logger.Warnf("TaskRun %s/%s has unexpected succeeded condition: %#v", tr.Namespace, tr.Name, condition)
		return nil
	}

	logger.Infof("Stopping sidecars for TaskRun %s/%s", tr.Namespace, tr.Name)
	_, err = podconvert.StopSidecars(ctx, configDefaults.NopImage, b.KubeClientSet, tr.Namespace, tr.Status.PodName)
	if k8serrors.IsNotFound(err) {
		// If the Pod isn't found it has probably been cleaned up already.
		return nil
	} else if err != nil {
		logger.Errorf("Error stopping sidecars for TaskRun %q: %v", tr.Name, err)
		return err
	}
	return nil
}

// LogEnqueueError allows functions that assist with enqueing to return an error.
// Normal workflows work better when errors can be returned (instead of just
// logged). Therefore logError allows these functions to return an error and
// it will take care of swallowing and logging it.
func LogEnqueueError(logger *zap.SugaredLogger, f func(interface{}) error) func(interface{}) {
	return func(obj interface{}) {
		if err := f(obj); err != nil {
			logger.Warn(err)
		}
	}
}

func init() {
	// Add serving types to the default Kubernetes Scheme so Events can be
	// logged for serving types.
	kfscheme.AddToScheme(scheme.Scheme)
}
