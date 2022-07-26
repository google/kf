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

package serviceinstance

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/reconcilerutil"
	"github.com/google/kf/v2/pkg/reconciler/serviceinstance/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

type Reconciler struct {
	*reconciler.ServiceCatalogBase

	spaceLister       kflisters.SpaceLister
	deploymentLister  appsv1listers.DeploymentLister
	volumeLister      v1listers.PersistentVolumeLister
	volumeClaimLister v1listers.PersistentVolumeClaimLister
	k8sServiceLister  v1listers.ServiceLister
	configStore       *config.Store
}

const serviceBindingFinalizer = "serviceinstancebinding.kf.dev"

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by knative/pkg when a new event is observed by one of the
// watchers in the controller.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	ctx = r.configStore.ToContext(ctx)

	return r.reconcileServiceInstance(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileServiceInstance(ctx context.Context, namespace, name string) (err error) {
	logger := logging.FromContext(ctx)
	original, err := r.KfServiceInstanceLister.ServiceInstances(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp().IsZero():
		// Register finalizer if it doesn't already exist on the service instance
		if !reconcilerutil.HasFinalizer(original, serviceBindingFinalizer) {
			toUpdate := original.DeepCopy()
			reconcilerutil.AddFinalizer(toUpdate, serviceBindingFinalizer)
			if _, err := r.update(ctx, toUpdate); err != nil {
				logger.Warnw("Failed to update ServiceInstance with finalizer", zap.Error(err))
				return err
			}
		}

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.ObservedGeneration = toUpdate.Generation

		// Handle finalizer
		if reconcilerutil.HasFinalizer(original, serviceBindingFinalizer) {
			bindingExists, err := r.serviceBindingExistsForServiceInstance(original)
			if err != nil {
				return err
			}
			if bindingExists {
				// Set status to DeletionBlocked
				toUpdate.Status.PropagateDeletionBlockedStatus()
			} else {
				// It is necessary to get the latest service instance information directly
				// from the cluster rather than relying on lister cached state which can be stale.
				// Otherwise a deprovision request can accidentally be sent twice to the OSB Broker,
				// which does not handle this idempotently and enters an error state
				toUpdate, err = r.KfClientSet.KfV1alpha1().ServiceInstances(namespace).Get(ctx, name, metav1.GetOptions{})
				toUpdate.Status.ObservedGeneration = toUpdate.Generation
				serviceDeleted := r.deleteService(ctx, toUpdate)
				if serviceDeleted {
					// Remove finalizer once the service instance is not part of any service
					// bindings and once DeleteService returns that it's done.
					reconcilerutil.RemoveFinalizer(toUpdate, serviceBindingFinalizer)
					if _, err := r.update(ctx, toUpdate); err != nil {
						logger.Warnw("Failed to update ServiceInstance", zap.Error(err))
						return err
					}

					return nil
				}
				if _, err := r.updateStatus(ctx, toUpdate); err != nil {
					logger.Warnw("Failed to update ServiceInstance status", zap.Error(err))
					return err
				}
				if !serviceDeleted {
					return errors.New("service instance not deleted")
				}

				return nil
			}
		} else {
			// Finalizer has already been removed, set status to Terminating
			toUpdate.Status.PropagateTerminatingStatus()
		}
		if _, uErr := r.updateStatus(ctx, toUpdate); uErr != nil {
			logger.Warnw("Failed to update ServiceInstance status", zap.Error(uErr))
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
		logger.Debugf("ServiceInstance reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) || reconcilerutil.IsConflictOSBError(reconcileErr) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
		//
		// Do not update status with OSB conflict error, it could cause race condition
		// and incorrect update to service instance status.
	} else if _, uErr := r.updateStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("Failed to update Service Instance status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

func (r *Reconciler) ApplyChanges(ctx context.Context, serviceinstance *v1alpha1.ServiceInstance) error {
	logger := logging.FromContext(ctx)

	// Default values on the service instance in case it hasn't been triggered since last update
	// to spec.
	serviceinstance.SetDefaults(ctx)

	serviceinstance.Status.InitializeConditions()

	// Ensure Kf Space exists to prevent Kf objects from being created in namespaces that's not a Kf Space.
	if _, err := r.spaceLister.Get(serviceinstance.Namespace); err != nil {
		serviceinstance.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	serviceinstance.Status.MarkSpaceHealthy()

	// Check secret
	var paramsSecret *corev1.Secret
	{
		logger.Debug("reconciling params secret")
		condition := serviceinstance.Status.ParamsSecretCondition()

		// Check that params secret exists
		paramsSecretName := serviceinstance.Spec.ParametersFrom.Name
		actual, err := r.SecretLister.Secrets(serviceinstance.Namespace).Get(paramsSecretName)
		if apierrs.IsNotFound(err) {
			logger.Info("Waiting for params secret to be created; exiting early")
			// Update status to secret missing
			serviceinstance.Status.PropagateSecretStatus(nil)
			return nil
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, serviceinstance) {
			return condition.MarkChildNotOwned(actual.Name)
		}

		paramsSecret = actual
		serviceinstance.Status.PropagateSecretStatus(actual)
	}

	{
		condition := serviceinstance.Status.ParamsSecretPopulatedCondition()
		if condition.IsPending() {
			logger.Info("Waiting for params secret to be populated; exiting early")
			return nil
		}
	}

	switch {
	case serviceinstance.IsUserProvided() && !serviceinstance.IsRouteService():
		// User-provided services do not have an additional backing resource unless they are a route service.
		serviceinstance.Status.MarkBackingResourceReady()

	case serviceinstance.IsUserProvided() && serviceinstance.IsRouteService():
		// Reconcile Deployment
		{
			logger.Debug("reconciling Deployment for route service proxy")
			condition := serviceinstance.Status.BackingResourceCondition()
			desired, err := resources.MakeDeployment(serviceinstance, config.FromContext(ctx))
			if err != nil {
				return condition.MarkTemplateError(err)
			}
			actual, err := r.deploymentLister.Deployments(desired.GetNamespace()).Get(desired.Name)
			if apierrs.IsNotFound(err) {
				actual, err = r.KubeClientSet.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating deployment", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest deployment", err)
			} else if !metav1.IsControlledBy(actual, serviceinstance) {
				return condition.MarkChildNotOwned(desired.Name)
			} else if actual, err = r.ReconcileDeployment(ctx, desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing deployment", err)
			}
			serviceinstance.Status.PropagateDeploymentStatus(actual)
		}

		// Reconcile Service
		{
			logger.Debug("reconciling Service for route service proxy")
			condition := serviceinstance.Status.BackingResourceCondition()
			desired := resources.MakeService(serviceinstance)
			actual, err := r.k8sServiceLister.Services(desired.GetNamespace()).Get(desired.Name)
			if apierrs.IsNotFound(err) {
				actual, err = r.KubeClientSet.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating Service", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest Service", err)
			} else if !metav1.IsControlledBy(actual, serviceinstance) {
				return condition.MarkChildNotOwned(desired.Name)
			} else if actual, err = r.ReconcileService(ctx, desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing Service", err)
			}
			// Services don't have conditions to indicate readiness, so we do not propagate Service status.
			// The BackingResource Condition on user-provided Route Services is only dependent on the proxy Deployment status.
		}

	case serviceinstance.IsLegacyBrokered():
		// If the instance is legacy brokered, don't reconcile it but do leave
		// a message.
		condition := serviceinstance.Status.BackingResourceCondition()
		condition.MarkFalse(
			"UnsupportedVersion",
			"Kubernetes Service Catalog backed services are no longer supported.")
		return nil

	case serviceinstance.IsKfBrokered():
		condition := serviceinstance.Status.BackingResourceCondition()
		// If the instance has already been actuated, don't try again.
		if !condition.IsPending() {
			break
		}

		// If the resource isn't making progress, terminate it:
		if timeoutErr := condition.ErrorIfTimeout(time.Duration(serviceinstance.Spec.OSB.ProgressDeadlineSeconds) * time.Second); timeoutErr != nil {
			serviceinstance.Status.PropagateProvisionStatus(nil, timeoutErr)
			break
		}

		osbClient, err := r.GetClientForServiceInstance(serviceinstance)
		if err != nil {
			condition.MarkReconciliationError("InstantiatingClient", err)
			break
		}

		// If there's a pending operation on the status, try again; otherwise
		// attempt to provision.
		if serviceinstance.Status.OSBStatus.IsBlank() {
			namespace, err := r.NamespaceLister.Get(serviceinstance.Namespace)
			if err != nil {
				return condition.MarkReconciliationError("GettingNamespace", err)
			}

			request, err := resources.MakeOSBProvisionRequest(serviceinstance, namespace, paramsSecret)
			if err != nil {
				return condition.MarkTemplateError(err)
			}

			response, err := osbClient.ProvisionInstance(request)
			if reconcilerutil.IsConflictOSBError(err) {
				return err
			}
			serviceinstance.Status.PropagateProvisionStatus(response, err)
		}

		if state := serviceinstance.Status.OSBStatus.Provisioning; state != nil {
			request := resources.MakeOSBLastOperationRequest(serviceinstance, state.OperationKey)
			response, err := osbClient.PollLastOperation(request)
			serviceinstance.Status.PropagateProvisionAsyncStatus(response, err)
		}
	case serviceinstance.IsVolume():
		// Reconcile Volume
		{
			condition := serviceinstance.Status.BackingResourceCondition()
			volumeInstanceParams, err := v1alpha1.ParseVolumeInstanceParams(paramsSecret)
			if err != nil {
				return condition.MarkTemplateError(fmt.Errorf("Error parsing VolumeInstanceParams, err: %s", err))
			}

			// Reconcile PersistentVolumeClaim
			// Override capacity value to no-op value of 1Gi since it will be dynamically provisioned.
			desiredClaim, err := resources.MakePersistentVolumeClaim(serviceinstance)
			if err != nil {
				return condition.MarkTemplateError(err)
			}

			actualClaim, err := r.volumeClaimLister.PersistentVolumeClaims(desiredClaim.Namespace).Get(desiredClaim.Name)
			if apierrs.IsNotFound(err) {
				actualClaim, err = r.KubeClientSet.
					CoreV1().
					PersistentVolumeClaims(desiredClaim.Namespace).
					Create(ctx, desiredClaim, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating PersistentVolumeClaim", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting PersistentVolumeClaim", err)
			}

			// Reconcile PersistentVolume
			logger.Debug("reconciling PersistentVolume for volume service")
			desiredVolume, err := resources.MakePersistentVolume(serviceinstance, volumeInstanceParams, actualClaim)
			if err != nil {
				return condition.MarkTemplateError(err)
			}
			actualVolume, err := r.volumeLister.Get(desiredVolume.Name)
			if apierrs.IsNotFound(err) {
				actualVolume, err = r.KubeClientSet.
					CoreV1().
					PersistentVolumes().
					Create(ctx, desiredVolume, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating persistentvolume", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting persistentvolume", err)
			}

			serviceinstance.Status.PropagateVolumeServiceStatus(serviceinstance, actualVolume.Name, actualClaim.Name)

			// PersistentVolume can't be altered after creation
			serviceinstance.Status.MarkBackingResourceReady()

		}
	default:
		return fmt.Errorf("ServiceType can't be determined for service instance %s", serviceinstance.Name)
	}
	// Copy the tags, class, and plan from the service instance into the status
	serviceinstance.Status.PropagateServiceFieldsStatus(serviceinstance)

	// Copy the route service URL from the service instance into the status
	serviceinstance.Status.PropagateRouteServiceURLStatus(serviceinstance)

	return nil
}

func (r *Reconciler) deletePersistentVolumeClaimForServiceInstance(ctx context.Context, serviceInstance *v1alpha1.ServiceInstance) (done bool) {
	condition := serviceInstance.Status.BackingResourceCondition()
	existing, err := r.volumeClaimLister.PersistentVolumeClaims(serviceInstance.Namespace).Get(resources.GetPersistentVolumeClaimName(serviceInstance.Name))
	if apierrs.IsNotFound(err) {
		return true
	} else if err != nil {
		condition.MarkReconciliationError("getting PersistentVolumeClaim", err)
		return false
	}

	if err = r.KubeClientSet.
		CoreV1().
		PersistentVolumeClaims(existing.Namespace).
		Delete(ctx, existing.Name, metav1.DeleteOptions{}); err != nil {
		condition.MarkReconciliationError("deleting PersistentVolumeClaim", err)
		return false
	}

	return true
}

func (r *Reconciler) deletePersistentVolumeForServiceInstance(ctx context.Context, serviceInstance *v1alpha1.ServiceInstance) (done bool) {
	condition := serviceInstance.Status.BackingResourceCondition()
	existing, err := r.volumeLister.Get(resources.GetPersistentVolumeName(serviceInstance.Name, serviceInstance.Namespace))
	if apierrs.IsNotFound(err) {
		return true
	} else if err != nil {
		condition.MarkReconciliationError("getting PersistentVolume", err)
		return false
	}

	if err = r.KubeClientSet.
		CoreV1().
		PersistentVolumes().
		Delete(ctx, existing.Name, metav1.DeleteOptions{}); err != nil {
		condition.MarkReconciliationError("deleting PersistentVolume", err)
		return false
	}

	return true
}

// deleteService handles any logic for cleaning up resources; it returns true
// once the resource can be finalized.
func (r *Reconciler) deleteService(ctx context.Context, serviceInstance *v1alpha1.ServiceInstance) (done bool) {
	// Default values on the object in case it hasn't been triggered since last update
	// to spec.
	serviceInstance.SetDefaults(ctx)

	switch {
	case serviceInstance.IsVolume():
		return r.deletePersistentVolumeClaimForServiceInstance(ctx, serviceInstance) &&
			r.deletePersistentVolumeForServiceInstance(ctx, serviceInstance)

	case serviceInstance.IsKfBrokered():
		// If provisioning failed or the resource is already deprovisioned,
		// expect the broker to have cleaned things up.
		condition := serviceInstance.Status.BackingResourceCondition()

		if serviceInstance.Status.OSBStatus.ProvisionFailed != nil ||
			serviceInstance.Status.OSBStatus.Deprovisioned != nil {
			return true
		}

		retryDelete := (serviceInstance.Status.DeleteRequests < serviceInstance.Spec.DeleteRequests)

		// Don't try deprovisioning if it's already failed, or if the resource is
		// still provisioning, unless another delete command has been issued.
		if !retryDelete && (serviceInstance.Status.OSBStatus.DeprovisionFailed != nil ||
			serviceInstance.Status.OSBStatus.Provisioning != nil) {
			return false
		}

		// Only check for timeout error if the service instance was already in Deprovisioning state. Checking for timeout
		// error for an instance that's not in Deprovisioning state could cause unintended behavior: if a service was
		// created longer than the timeout ago, its condition's TimeSinceTransition will be greater than the timeout,
		// and thus it causes the service instance to be incorrectly marked for DeprovisionFailed and thus causes the
		// service instance can't be deleted via the OSB client (see /b/187713332).
		if serviceInstance.Status.OSBStatus.Deprovisioning != nil {
			if timeoutErr := condition.ErrorIfTimeout(time.Duration(serviceInstance.Spec.OSB.ProgressDeadlineSeconds) * time.Second); timeoutErr != nil {
				serviceInstance.Status.PropagateDeprovisionStatus(nil, timeoutErr)
				return false
			}
		}

		osbClient, err := r.GetClientForServiceInstance(serviceInstance)
		if err != nil {
			condition.MarkReconciliationError("InstantiatingClient", err)
			return false
		}

		// If the service is currently provisioned or another delete command has been issued, attempt to delete it.
		if serviceInstance.Status.OSBStatus.Provisioned != nil || retryDelete {
			request := resources.MakeOSBDeprovisionRequest(serviceInstance)
			response, err := osbClient.DeprovisionInstance(request)
			serviceInstance.Status.PropagateDeprovisionStatus(response, err)
			serviceInstance.Status.DeleteRequests = serviceInstance.Spec.DeleteRequests
		}

		// If the service is currently deprovisioning, poll it.
		if state := serviceInstance.Status.OSBStatus.Deprovisioning; state != nil {
			request := resources.MakeOSBLastOperationRequest(serviceInstance, state.OperationKey)
			response, err := osbClient.PollLastOperation(request)
			serviceInstance.Status.PropagateDeprovisionAsyncStatus(response, err)
		}

		return serviceInstance.Status.OSBStatus.Deprovisioned != nil

	}

	// If the resource isn't brokered or Volume, there's nothing to do because the lifecycle
	// is managed by the finalizers of the child objects.
	return true
}

func (r *Reconciler) update(ctx context.Context, desired *v1alpha1.ServiceInstance) (*v1alpha1.ServiceInstance, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating")
	actual, err := r.KfServiceInstanceLister.ServiceInstances(desired.GetNamespace()).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual, desired) {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := desired.DeepCopy()
	return r.KfClientSet.KfV1alpha1().ServiceInstances(existing.GetNamespace()).Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.ServiceInstance) (*v1alpha1.ServiceInstance, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating status")
	actual, err := r.KfServiceInstanceLister.ServiceInstances(desired.GetNamespace()).Get(desired.Name)
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
	return r.KfClientSet.KfV1alpha1().ServiceInstances(existing.GetNamespace()).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) serviceBindingExistsForServiceInstance(serviceinstance *v1alpha1.ServiceInstance) (bool, error) {
	bindings, err := r.KfServiceInstanceBindingLister.ServiceInstanceBindings(serviceinstance.Namespace).List(labels.Everything())
	if err != nil {
		return false, err
	}
	for _, binding := range bindings {
		if binding.Spec.InstanceRef.Name == serviceinstance.Name {
			return true, nil
		}
	}
	return false, nil
}
