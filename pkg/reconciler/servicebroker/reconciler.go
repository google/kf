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

package servicebroker

import (
	"context"
	"reflect"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/osbutil"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/clusterservicebroker"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const serviceBrokerFinalizer = "servicebroker.kf.dev"

// Reconciler implements controller.Reconciler.
type Reconciler struct {
	*reconciler.Base
	kfServiceBrokerLister   kflisters.ServiceBrokerLister
	kfServiceInstanceLister kflisters.ServiceInstanceLister
}

var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative.dev/pkg when a new event is observed by one of
// the watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// NOTE: we don't use knative.dev/pkg's generated reconcilers here
	// because they assume the finalizer will only ever need to be called
	// once. Kf uses the finalizers to prevent deletion if subresources still
	// exist.

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	logger := logging.FromContext(ctx)

	original, err := r.kfServiceBrokerLister.ServiceBrokers(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		logger.Errorw("couldn't get resource", zap.Error(err))
		return err

	case original.GetDeletionTimestamp().IsZero():
		// Register finalizer if it doesn't already exist on the service instance
		if !reconcilerutil.HasFinalizer(original, serviceBrokerFinalizer) {
			toUpdate := original.DeepCopy()
			reconcilerutil.AddFinalizer(toUpdate, serviceBrokerFinalizer)
			if err := r.update(ctx, toUpdate); err != nil {
				logger.Warnw("couldn't add finalizer", zap.Error(err))
				return err
			}
			return nil
		}

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.ObservedGeneration = toUpdate.Generation

		// Handle finalizer
		if reconcilerutil.HasFinalizer(original, serviceBrokerFinalizer) {
			childrenExist, err := r.serviceInstanceExistsForServiceBroker(original)
			if err != nil {
				return err
			}
			if childrenExist {
				// Set status to DeletionBlocked
				toUpdate.Status.PropagateDeletionBlockedStatus()
			} else {
				// Remove finalizer once the broker is not part of any service instances.
				reconcilerutil.RemoveFinalizer(toUpdate, serviceBrokerFinalizer)
				if err := r.update(ctx, toUpdate); err != nil {
					logger.Warnw("failed to update broker", zap.Error(err))
					return err
				}

				return nil
			}
		} else {
			// Finalizer has already been removed, set status to Terminating
			toUpdate.Status.PropagateTerminatingStatus()
		}
		if err := r.updateStatus(ctx, toUpdate); err != nil {
			logger.Warnw("failed to update status", zap.Error(err))
			return err
		}
		return nil
	}

	// the following code is outside of the switch statement because
	// the original.GetDeletionTimestamp().IsZero() condition case
	// falls through so actuation can occur off the same triggering
	// event as the finalizer is added on.

	if r.IsNamespaceTerminating(namespace) {
		logger.Debug("namespace is terminating, skipping reconciliation")
		return nil
	}

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
	r.applyChanges(ctx, toReconcile)
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if err := r.updateStatus(ctx, toReconcile); err != nil {
		logger.Warnw("failed to update resource status", zap.Error(err))
		return err
	}

	return nil
}

// applyChanges reconciles the object to the cluster. This function observes and
// mutates the cluster to bring it into alignment with spec then updates the
// broker status to reflect the newly observed state.
func (r *Reconciler) applyChanges(ctx context.Context, broker *v1alpha1.ServiceBroker) {
	logger := logging.FromContext(ctx)

	// Default values on the service instance in case it hasn't been triggered since last update
	// to spec.
	broker.SetDefaults(ctx)

	broker.Status.InitializeConditions()

	// Check that the credential Secret has been created and is owned by
	// this resource. Ownership is important because it prevents a bad
	// actor from reading data from arbitrary Secrets.
	var credentialsSecret *corev1.Secret
	{
		logger.Debug("reconciling creds secret")
		condition := broker.Status.CredsSecretCondition()

		// Check that params secret exists
		credsRef := broker.Spec.Credentials
		actual, err := r.SecretLister.Secrets(broker.Namespace).Get(credsRef.Name)
		if apierrs.IsNotFound(err) {
			logger.Info("Waiting for secret to be created; exiting early")
			// Update status to secret missing
			broker.Status.PropagateSecretStatus(nil)
			return
		} else if err != nil {
			condition.MarkReconciliationError("getting latest", err)
			return
		} else if !metav1.IsControlledBy(actual, broker) {
			condition.MarkChildNotOwned(actual.Name)
			return
		}
		broker.Status.PropagateSecretStatus(actual)

		if _, err := osbutil.NewConfigFromSecret(actual); err != nil {
			broker.Status.CredsSecretPopulatedCondition().MarkTemplateError(err)
			return
		}
		// Otherwise, the secret is good
		broker.Status.CredsSecretPopulatedCondition().MarkSuccess()
		credentialsSecret = actual
	}

	logger.Info("Reconciling catalog")
	if err := clusterservicebroker.ReconcileCatalog(
		credentialsSecret,
		&broker.Spec.CommonServiceBrokerSpec,
		&broker.Status,
	); err != nil {
		logger.Errorf("couldn't reconcile catalog %#v", err)
		return
	}
}

func (r *Reconciler) update(ctx context.Context, desired *v1alpha1.ServiceBroker) error {
	logger := logging.FromContext(ctx)
	logger.Info("updating")
	actual, err := r.kfServiceBrokerLister.ServiceBrokers(desired.GetNamespace()).Get(desired.Name)
	if err != nil {
		return err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual, desired) {
		return nil
	}

	// Don't modify the informers copy.
	existing := desired.DeepCopy()

	_, err = r.KfClientSet.KfV1alpha1().ServiceBrokers(existing.GetNamespace()).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.ServiceBroker) error {
	actual, err := r.kfServiceBrokerLister.ServiceBrokers(desired.GetNamespace()).Get(desired.Name)
	if err != nil {
		return err
	}

	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()
	existing.Status = desired.Status

	_, err = r.KfClientSet.KfV1alpha1().ServiceBrokers(existing.GetNamespace()).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
	return err
}

func (r *Reconciler) serviceInstanceExistsForServiceBroker(serviceBroker *v1alpha1.ServiceBroker) (bool, error) {
	children, err := r.kfServiceInstanceLister.ServiceInstances(serviceBroker.Namespace).List(labels.Everything())
	if err != nil {
		return false, err
	}
	for _, child := range children {
		if brokered := child.Spec.ServiceType.Brokered; brokered != nil {
			if brokered.Broker == serviceBroker.Name && brokered.Namespaced {
				return true, nil
			}
		}
	}
	return false, nil
}
