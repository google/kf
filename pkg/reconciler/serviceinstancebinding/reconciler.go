// Copyright 2020 Google LLC
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

package serviceinstancebinding

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"github.com/google/kf/v2/pkg/reconciler/serviceinstancebinding/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
)

const serviceInstanceBindingFinalizer = "serviceinstancebinding.kf.dev"

type Reconciler struct {
	*reconciler.ServiceCatalogBase

	spaceLister kflisters.SpaceLister
	appLister   kflisters.AppLister

	persistentVolumeClaimLister v1listers.PersistentVolumeClaimLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of the
// watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	return r.reconcileServiceBinding(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileServiceBinding(ctx context.Context, namespace, name string) (err error) {
	logger := logging.FromContext(ctx)
	original, err := r.KfServiceInstanceBindingLister.ServiceInstanceBindings(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp().IsZero():
		// Register finalizer if it doesn't already exist on the resource so the
		// OSB flow can have time to complete before a resource gets deleted.
		if !reconcilerutil.HasFinalizer(original, serviceInstanceBindingFinalizer) {
			toUpdate := original.DeepCopy()
			reconcilerutil.AddFinalizer(toUpdate, serviceInstanceBindingFinalizer)
			if err := r.update(ctx, toUpdate); err != nil {
				logger.Warnw("Failed to update resource with finalizer", zap.Error(err))
				return err
			}
		}

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.ObservedGeneration = toUpdate.Generation

		// Handle finalizer
		if reconcilerutil.HasFinalizer(original, serviceInstanceBindingFinalizer) {
			// Remove finalizer once deleteServiceBinding returns that it's done.
			if r.deleteServiceBinding(ctx, toUpdate) {
				reconcilerutil.RemoveFinalizer(toUpdate, serviceInstanceBindingFinalizer)
				if err := r.update(ctx, toUpdate); err != nil {
					logger.Warnw("Failed to update resource", zap.Error(err))
					return err
				}
				return nil
			}
		} else {
			// Finalizer has already been removed, set status to Terminating
			toUpdate.Status.PropagateTerminatingStatus()
		}
		if _, uErr := r.updateStatus(ctx, toUpdate); uErr != nil {
			logger.Warnw("Failed to update resource status", zap.Error(uErr))
			return uErr
		}
		return nil
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Info("namespace is terminating, skipping reconciliation")
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

	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if reconcileErr != nil {
		logger.Debugf("ServiceInstanceBinding reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) || reconcilerutil.IsConflictOSBError(reconcileErr) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
		//
		// Do not update status with OSB conflict error, it could cause race condition
		// and incorrect update to service instance binding status.
	} else if _, uErr := r.updateStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("Failed to update Service Instance status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

func (r *Reconciler) ApplyChanges(ctx context.Context, binding *v1alpha1.ServiceInstanceBinding) error {
	logger := logging.FromContext(ctx)
	// Default values on the service instance binding in case it hasn't been triggered since last update
	// to spec.
	binding.SetDefaults(ctx)
	binding.Status.InitializeConditions()

	var paramsSecret *v1.Secret
	// Check secret
	{
		logger.Debug("reconciling params secret")
		condition := binding.Status.ParamsSecretCondition()

		// Check that params secret exists
		paramsSecretName := binding.Spec.ParametersFrom.Name
		actual, err := r.SecretLister.Secrets(binding.Namespace).Get(paramsSecretName)
		if apierrs.IsNotFound(err) {
			logger.Info("Waiting for params secret to be created; exiting early")
			// Update status to secret missing
			binding.Status.PropagateParamsSecretStatus(nil)
			return nil
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		}

		binding.Status.PropagateParamsSecretStatus(actual)
		paramsSecret = actual
	}

	{
		condition := binding.Status.ParamsSecretPopulatedCondition()
		if condition.IsPending() {
			logger.Info("Waiting for params secret to be populated; exiting early")
			return nil
		}
	}

	// Check service instance
	condition := binding.Status.ServiceInstanceCondition()
	serviceInstance, err := r.GetInstanceForBinding(binding)
	if err != nil {
		return condition.MarkReconciliationError("getting service instance", err)
	}

	// If the service instance is not ready, do not continue with the binding.
	binding.Status.PropagateServiceInstanceStatus(serviceInstance)
	if condition.IsPending() {
		logger.Info("Waiting for service instance in binding to be ready; exiting early")
		return nil
	}

	// Propagate service fields from referenced service instance to the binding.
	binding.Status.PropagateServiceFieldsStatus(serviceInstance)

	// Propagate binding name to the service instance binding status.
	binding.Status.PropagateBindingNameStatus(binding)

	// Propagate route service URL from referenced service instance to the binding.
	binding.Status.PropagateRouteServiceURLStatus(serviceInstance)

	// Propagate volume status from referenced service instance to the binding.
	binding.Status.PropagateVolumeStatus(serviceInstance, paramsSecret)

	switch {
	// UserProvidedService and VolumeService don't have backing resources.
	case serviceInstance.HasNoBackingResources():
		binding.Status.MarkBackingResourceReady()
	case serviceInstance.IsLegacyBrokered():
		// If the instance is legacy brokered, don't reconcile it but do leave
		// a message.
		condition := binding.Status.BackingResourceCondition()
		condition.MarkFalse(
			"UnsupportedVersion",
			"Kubernetes Service Catalog backed services are no longer supported.")
		return nil
	case serviceInstance.IsKfBrokered():
		condition := binding.Status.BackingResourceCondition()
		// If the instance has already been actuated, don't try again.
		if !condition.IsPending() {
			break
		}

		// If the resource isn't making progress, terminate it:
		if timeoutErr := condition.ErrorIfTimeout(time.Duration(binding.Spec.ProgressDeadlineSeconds) * time.Second); timeoutErr != nil {
			binding.Status.PropagateBindStatus(nil, timeoutErr)
			break
		}

		osbClient, err := r.GetClientForServiceInstance(serviceInstance)
		if err != nil {
			condition.MarkReconciliationError("InstantiatingClient", err)
			break
		}

		// Attempt to provision if there is no existing state.
		if binding.Status.OSBStatus.IsBlank() {
			namespace, err := r.NamespaceLister.Get(serviceInstance.Namespace)
			if err != nil {
				return condition.MarkReconciliationError("GettingNamespace", err)
			}

			request, err := resources.MakeOSBBindRequest(serviceInstance, binding, namespace, paramsSecret)
			if err != nil {
				return condition.MarkTemplateError(err)
			}

			response, err := osbClient.Bind(request)
			if reconcilerutil.IsConflictOSBError(err) {
				return err
			}
			binding.Status.PropagateBindStatus(response, err)

			// If the response wasn't async, write the secret now.
			if err == nil && !response.Async {
				r.createBindingSecretAndUpdateStatus(ctx, binding, func() (*corev1.Secret, error) {
					return resources.MakeCredentialsForOSBService(binding, response.Credentials)
				})
			}
		}

		// Poll if there's a pending operation on the status.
		if state := binding.Status.OSBStatus.Binding; state != nil {
			request := resources.MakeOSBBindingLastOperationRequest(serviceInstance, binding, state.OperationKey)
			response, err := osbClient.PollBindingLastOperation(request)
			binding.Status.PropagateBindLastOperationStatus(response, err)
		}
	default:
		return fmt.Errorf("ServiceType can't be determined for service instance %s", serviceInstance.Name)
	}

	// Reconcile binding credentials secret
	{
		logger.Debug("reconciling binding credentials secret")
		condition := binding.Status.CredentialsSecretCondition()

		// NOTE: Resources ALWAYS create a secret (even if blank) because
		// credentials can be returned from any binding and because not
		// creating one would leak information to people who have the ability
		// to see reosources but not their contents.

		switch {
		case serviceInstance.HasNoBackingResources():
			credentialsSecretName := serviceInstance.Spec.ParametersFrom.Name
			credentialsSecret, err := r.SecretLister.Secrets(serviceInstance.Namespace).Get(credentialsSecretName)
			if apierrs.IsNotFound(err) {
				logger.Info("Waiting for service instance credentials secret to be created; exiting early")
				return nil
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			}
			// Create credentials secret with original credentials merged with any parameters provided in the binding.
			r.createBindingSecretAndUpdateStatus(ctx, binding, func() (*corev1.Secret, error) {
				return resources.MergeCredentialsSecretForBinding(*binding, *credentialsSecret, *paramsSecret)
			})
		case serviceInstance.IsLegacyBrokered():
			condition.MarkFalse(
				"UnsupportedVersion",
				"Kubernetes Service Catalog backed services are no longer supported.")
			return nil
		case serviceInstance.IsKfBrokered():
			// If the state isn't bound, don't create the backing secrets.
			if binding.Status.OSBStatus.Bound == nil {
				logger.Info("Waiting binding to be ready, exiting early")
				return nil
			}

			// If secret is already created, don't poll the OSB resource
			// again.
			if binding.Status.CredentialsSecretRef.Name != "" {
				logger.Info("Secret already exists")
				return nil
			}

			osbClient, oerr := r.GetClientForServiceInstance(serviceInstance)
			if oerr != nil {
				return condition.MarkReconciliationError("InstantiatingClient", oerr)
			}

			bindingCreds, berr := osbClient.GetBinding(resources.MakeOSBGetBindingRequest(serviceInstance, binding))
			if berr != nil {
				return condition.MarkReconciliationError("GettingBindingCreds", berr)
			}

			// XXX: Handle additional binding types here.
			r.createBindingSecretAndUpdateStatus(ctx, binding, func() (*corev1.Secret, error) {
				return resources.MakeCredentialsForOSBService(binding, bindingCreds.Credentials)
			})
		default:
			return fmt.Errorf("ServiceType can't be determined for service instance %s", serviceInstance.Name)
		}
	}

	return nil
}

func (r *Reconciler) createBindingSecretAndUpdateStatus(ctx context.Context, binding *v1alpha1.ServiceInstanceBinding, secretBuilder func() (*corev1.Secret, error)) error {
	logger := logging.FromContext(ctx)
	condition := binding.Status.CredentialsSecretCondition()

	desiredSecret, err := secretBuilder()
	if err != nil {
		return condition.MarkTemplateError(err)
	}

	actual, err := r.SecretLister.Secrets(desiredSecret.GetNamespace()).Get(desiredSecret.Name)
	if apierrs.IsNotFound(err) {
		actual, err = r.KubeClientSet.CoreV1().Secrets(desiredSecret.GetNamespace()).Create(ctx, desiredSecret, metav1.CreateOptions{})
		if err != nil {
			return condition.MarkReconciliationError("creating", err)
		}
	} else if err != nil {
		return condition.MarkReconciliationError("getting latest", err)
	} else {
		// Update to desired.
		diff, err := kmp.SafeDiff(desiredSecret.Data, actual.Data)
		if err != nil {
			return fmt.Errorf("failed to diff secret spec: %v", err)
		}
		logger.Debug("Secret.Data diff:", diff)

		// Don't modify the informers copy.
		existing := actual.DeepCopy()
		existing.Data = desiredSecret.Data

		actual, err = r.KubeClientSet.
			CoreV1().
			Secrets(existing.Namespace).
			Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	binding.Status.PropagateCredentialsSecretStatus(actual)

	return nil
}

// deleteServiceBinding handles any logic for cleaning up resources; it returns
// true once the resource can be finalized.
func (r *Reconciler) deleteServiceBinding(ctx context.Context, binding *v1alpha1.ServiceInstanceBinding) (done bool) {
	// Default values on the object in case it hasn't been triggered since last update
	// to spec.
	binding.SetDefaults(ctx)

	condition := binding.Status.BackingResourceCondition()

	serviceInstance, err := r.GetInstanceForBinding(binding)
	if err != nil {
		condition.MarkReconciliationError("GettingInstance", err)
		return false
	}

	// If the resource isn't brokered, there's nothing to do because the lifecycle
	// is managed by the finalizers of the child objects.
	if !serviceInstance.IsKfBrokered() {
		return true
	}

	// If provisioning failed or the resource is already unbound,
	// expect the broker to have cleaned things up.
	if binding.Status.OSBStatus.BindFailed != nil ||
		binding.Status.OSBStatus.Unbound != nil {
		return true
	}

	// Don't try unbinding if the resource is still binding
	if binding.Status.OSBStatus.Binding != nil {
		return false
	}

	retryUnbind := (binding.Status.UnbindRequests != binding.Spec.UnbindRequests)

	// Don't try unbinding if it's already failed, and user has not attempted another try
	if !retryUnbind && binding.Status.OSBStatus.UnbindFailed != nil {
		return false
	}

	// Only check for timeout error if the service instance binding was already in Unbinding state. Checking for timeout
	// error for an instance that's not in Unbinding state could cause unintended behavior: if a service instance binding was
	// created longer than the timeout ago, its condition's TimeSinceTransition will be greater than the timeout,
	// and thus it causes the service instance binding to be incorrectly marked for UnbindFailed and thus causes the
	// service instance binding can't be deleted via the OSB client (see /b/187713332).
	if binding.Status.OSBStatus.Unbinding != nil {
		if timeoutErr := condition.ErrorIfTimeout(time.Duration(binding.Spec.ProgressDeadlineSeconds) * time.Second); timeoutErr != nil {
			binding.Status.PropagateUnbindStatus(nil, timeoutErr)

			return false
		}
	}

	osbClient, err := r.GetClientForServiceInstance(serviceInstance)
	if err != nil {
		condition.MarkReconciliationError("InstantiatingClient", err)
		return false
	}

	// If the service is currently bound, attempt to delete it.
	if binding.Status.OSBStatus.Bound != nil || retryUnbind {
		request := resources.MakeOSBUnbindRequest(serviceInstance, binding)
		response, err := osbClient.Unbind(request)
		binding.Status.PropagateUnbindStatus(response, err)

		// Set binding status to spec instead of +1, to avoid multiple tries even if user tried many unbind request in a short period of time.
		binding.Status.UnbindRequests = binding.Spec.UnbindRequests
	}

	// If the service is currently unbinding, poll it.
	if state := binding.Status.OSBStatus.Unbinding; state != nil {
		request := resources.MakeOSBBindingLastOperationRequest(serviceInstance, binding, state.OperationKey)
		response, err := osbClient.PollBindingLastOperation(request)
		binding.Status.PropagateUnbindLastOperationStatus(response, err)
	}

	return binding.Status.OSBStatus.Unbound != nil
}

func (r *Reconciler) update(ctx context.Context, desired *v1alpha1.ServiceInstanceBinding) error {
	logger := logging.FromContext(ctx)
	logger.Info("updating")
	actual, err := r.KfServiceInstanceBindingLister.
		ServiceInstanceBindings(desired.GetNamespace()).
		Get(desired.Name)
	if err != nil {
		return err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual, desired) {
		return nil
	}

	// Don't modify the informers copy.
	existing := desired.DeepCopy()

	_, err = r.KfClientSet.KfV1alpha1().
		ServiceInstanceBindings(existing.GetNamespace()).
		Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.ServiceInstanceBinding) (*v1alpha1.ServiceInstanceBinding, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating status")
	actual, err := r.KfServiceInstanceBindingLister.ServiceInstanceBindings(desired.GetNamespace()).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().ServiceInstanceBindings(existing.GetNamespace()).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}
