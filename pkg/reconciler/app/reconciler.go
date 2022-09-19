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

package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/app/resources"
	spaces "github.com/google/kf/v2/pkg/reconciler/space/resources"
	"go.uber.org/zap"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	autoscalingv1listers "k8s.io/client-go/listers/autoscaling/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

var (
	restageNeededErr        = errors.New("a restage is needed to reflect the latest build settings")
	adxBuildNotInstalledErr = errors.New("ADX builds is not installed however BuildRef is set. This could be that the controller pod needs to be restarted")
)

type Reconciler struct {
	*reconciler.Base

	buildLister                  kflisters.BuildLister
	appLister                    kflisters.AppLister
	spaceLister                  kflisters.SpaceLister
	routeLister                  kflisters.RouteLister
	serviceInstanceBindingLister kflisters.ServiceInstanceBindingLister
	deploymentLister             appsv1listers.DeploymentLister
	serviceLister                v1listers.ServiceLister
	serviceAccountLister         v1listers.ServiceAccountLister
	autoscalingLister            autoscalingv1listers.HorizontalPodAutoscalerLister
	adxBuildLister               cache.GenericLister

	kfConfigStore *kfconfig.Store
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

	return r.reconcileApp(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileApp(ctx context.Context, namespace, name string) (err error) {
	logger := logging.FromContext(ctx)
	original, err := r.appLister.Apps(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.PropagateTerminatingStatus()
		if _, uErr := r.updateStatus(ctx, toUpdate); uErr != nil {
			logger.Warnw("Failed to update App status", zap.Error(uErr))
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
		logger.Debugf("App reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("Failed to update App status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

func (r *Reconciler) ApplyChanges(ctx context.Context, app *v1alpha1.App) error {
	logger := logging.FromContext(ctx)
	ctx = r.kfConfigStore.ToContext(ctx)

	// Default values on the app in case it hasn't been triggered since last update
	// to spec.
	app.SetDefaults(ctx)

	app.Status.InitializeConditions()

	// Ensure Kf Space exists to prevent Kf objects from being created in namespaces that's not a Kf Space.
	space, err := r.spaceLister.Get(app.Namespace)
	if err != nil {
		app.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	app.Status.MarkSpaceHealthy()

	// Reconcile service instance bindings related to the app.
	var bindingsWithApp []v1alpha1.ServiceInstanceBinding
	// Gather volume bindings to reconcile next.
	var volumeBindings []*v1alpha1.ServiceInstanceBinding

	{
		logger.Debug("reconciling service instance bindings")
		condition := app.Status.ServiceInstanceBindingsCondition()
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		existingBindings, err := r.serviceInstanceBindingLister.
			ServiceInstanceBindings(app.GetNamespace()).
			List(labels.NewSelector())
		if err != nil {
			return condition.MarkReconciliationError("getting service instance bindings", err)
		}
		for _, binding := range existingBindings {
			if binding.Spec.App != nil && binding.Status.IsReady() {
				if binding.Spec.App.Name == app.Name {
					bindingsWithApp = append(bindingsWithApp, *binding)
					if binding.Status.VolumeStatus != nil {
						volumeBindings = append(volumeBindings, binding)
					}
				}
			}
		}

		// Propagate service instance bindings status.
		app.Status.PropagateServiceInstanceBindingsStatus(bindingsWithApp)
		if condition.IsPending() {
			logger.Info("Waiting for service instance bindings; exiting early")
			return nil
		}

		// Reconcile volume bindings. This has to be done before pods are started to get the correct volumes.
		app.Status.PropagateVolumeBindingsStatus(volumeBindings)
	}

	// Reconcile VCAP env vars secret
	{
		logger.Debug("reconciling env vars secret")
		condition := app.Status.EnvVarSecretCondition()
		systemEnvInjector := cfutil.NewSystemEnvInjector(r.KubeClientSet)
		desired, err := resources.MakeKfInjectedEnvSecret(ctx, app, space, bindingsWithApp, systemEnvInjector)

		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.SecretLister.Secrets(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Secrets(desired.GetNamespace()).Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileSecret(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}
		app.Status.PropagateEnvVarSecretStatus(actual)
	}

	// reconcile service
	{
		logger.Debug("reconciling service")
		condition := app.Status.ServiceCondition()
		desired := resources.MakeService(app)

		actual, err := r.serviceLister.Services(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileService(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}

		app.Status.PropagateServiceStatus(actual)
	}

	// Reconcile route, this MUST be done before pods are started to get correct
	// VCAP_APPLICATION.
	{
		logger.Debug("reconciling Routes")
		condition := app.Status.RouteCondition()

		desiredRoutes, desiredBindings, err := resources.MakeRoutes(app, space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		var actualRoutes []v1alpha1.Route

		for _, desired := range desiredRoutes {
			actual, err := r.routeLister.
				Routes(desired.GetNamespace()).
				Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// Route doesn't exist, make one.
				actual, err = r.KfClientSet.
					KfV1alpha1().
					Routes(desired.GetNamespace()).
					Create(ctx, &desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if actual, err = r.reconcileRoute(ctx, &desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}

			actualRoutes = append(actualRoutes, *actual.DeepCopy())
		}

		// Grab the full state of the world to see if any Routes are holding onto
		// bindings that don't exist anymore.
		desiredSet := make(map[v1alpha1.QualifiedRouteBinding]bool)
		for _, v := range desiredBindings {
			desiredSet[v] = true
		}

		allRoutes, err := r.routeLister.
			Routes(app.GetNamespace()).
			List(labels.Everything())
		if err != nil {
			return condition.MarkReconciliationError("listing all", err)
		}

		var undeclaredBindings []v1alpha1.QualifiedRouteBinding
		for _, route := range allRoutes {
			for _, destination := range route.Status.Bindings {
				qualified := v1alpha1.QualifiedRouteBinding{
					Source:      route.Spec.RouteSpecFields,
					Destination: destination,
				}

				_, found := desiredSet[qualified]

				if destination.ServiceName == app.Name && !found {
					undeclaredBindings = append(undeclaredBindings, qualified)
				}
			}
		}

		// If there are no errors reconciling Routes, propagate the status.
		app.Status.PropagateRouteStatus(desiredBindings, actualRoutes, undeclaredBindings)
	}

	// reconcile service account
	{
		logger.Debug("reconciling Service Account")
		condition := app.Status.ServiceAccountCondition()

		// Fetch the secrets referenced in the build service account, and filter out the token secret auto-generated by K8s.
		// If Workload Identity is enabled, there should be no secrets after filtering.
		// If WI is not enabled, then the secrets should have the same values as those from build.ImagePushSecrets in config-secrets.
		// These secrets will be passed into ImagePullSecrets on the App Service Account.
		// ImagePullSecrets are only accessed by the kubelet and are not mounted in the Pod.
		buildSA, err := r.serviceAccountLister.ServiceAccounts(app.Namespace).Get(spaces.BuildServiceAccountName(space))
		if err != nil {
			return condition.MarkReconciliationError("getting build SA", err)
		}

		buildSecrets := []*v1.Secret{}
		for _, s := range buildSA.Secrets {
			secret, err := r.SecretLister.Secrets(app.Namespace).Get(s.Name)
			if err != nil {
				return condition.MarkReconciliationError("getting secret", err)
			}
			buildSecrets = append(buildSecrets, secret)
		}

		filteredSecretNames := []v1.LocalObjectReference{}
		for _, s := range resources.FilterAndSortKfSecrets(buildSecrets) {
			filteredSecretNames = append(filteredSecretNames, v1.LocalObjectReference{Name: s.Name})
		}

		desired := resources.MakeServiceAccount(app, filteredSecretNames)

		actual, err := r.serviceAccountLister.ServiceAccounts(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			// Service account doesn't exist, create a new one
			actual, err = r.KubeClientSet.CoreV1().ServiceAccounts(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileServiceAccount(ctx, desired, actual, false); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}

		app.Status.PropagateServiceAccountStatus(actual)

		if condition.IsPending() {
			logger.Info("Waiting for service account; exiting early")
			return nil
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	// Anything that starts a Pod should go below this point to ensure the
	// environment the Pod runs in is set up.
	//////////////////////////////////////////////////////////////////////////////

	// reconcile Build
	{
		logger.Debug("reconciling Build")
		condition := app.Status.BuildCondition()

		if app.Spec.Build.Image != nil {
			logger.Debug("image supplied, skipping Build")
			app.Status.Image = *app.Spec.Build.Image
			condition.MarkSuccess()
		} else if app.Spec.Build.BuildRef != nil {

			// Ensure ADX builds are installed and warn if it is not.
			if !isADXBuildInstalled(ctx, r.KfClientSet, logger) {
				logger.Warn(adxBuildNotInstalledErr.Error())
				return adxBuildNotInstalledErr
			}

			buildName := app.Spec.Build.BuildRef.Name
			logger.Debug("buildRef supplied, using AppDevExperience Build")

			actual, err := r.adxBuildLister.
				ByNamespace(app.GetNamespace()).
				Get(buildName)
			if apierrs.IsNotFound(err) {
				// Not found, so wait for it.
				// NOTE: Return an error so the key is re-enqueued.
				logger.Info("Waiting for Build; exiting early")
				return fmt.Errorf("waiting for Build %s", buildName)
			} else {
				// Ensure the OwnerReference is the App.
				for _, ref := range actual.(*unstructured.Unstructured).GetOwnerReferences() {
					if ref.UID != app.UID {
						return condition.MarkChildNotOwned(buildName)
					}
				}
			}

			actualU, _ := actual.(*unstructured.Unstructured)
			switch dynamicutils.CheckCondtions(ctx, actualU) {
			case corev1.ConditionFalse:
				condition.MarkFalse("BuildFailed", "Build has failed")
			case corev1.ConditionUnknown:
				logger.Info("Waiting for Build; exiting early")
				condition.MarkUnknown("Build", "waiting for build")
				return fmt.Errorf("waiting for Build %s", buildName)
			case corev1.ConditionTrue:
				if err := app.Status.PropagateADXBuildStatus(actualU); err != nil {
					return condition.MarkReconciliationError("parse build", err)
				}
				condition.MarkSuccess()
			}

		} else {
			desired, err := resources.MakeBuild(app, space)
			if err != nil {
				return condition.MarkTemplateError(err)
			}

			// Builds are only triggered on a name change which occurs when the user
			// increments it or a webhook detects a change and triggers it.
			actual, err := r.buildLister.Builds(desired.GetNamespace()).Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// Build doesn't exist, create a new one
				actual, err = r.KfClientSet.KfV1alpha1().Builds(desired.GetNamespace()).Create(ctx, desired, metav1.CreateOptions{})
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if !metav1.IsControlledBy(actual, app) {
				return condition.MarkChildNotOwned(desired.Name)
			}

			// TODO: Add an indicator to the AppStatus when the Build is out of
			// date due to a stack (or other) change.

			app.Status.PropagateBuildStatus(actual)

			if condition.IsPending() {
				logger.Info("Waiting for build; exiting early")
				return nil
			}
		}
	}

	// GC'ing Builds
	{
		logger.Debug("GC'ing Builds")

		configDefaults, err := kfconfig.FromContext(ctx).Defaults()
		if err != nil {
			return fmt.Errorf("failed to read config-defaults: %v", err)
		}

		buildLabelSelector := fmt.Sprintf("%s=%s", v1alpha1.NameLabel, app.Name)
		listOptions := metav1.ListOptions{
			LabelSelector: buildLabelSelector,
		}
		buildList, err := r.KfClientSet.
			KfV1alpha1().
			Builds(app.GetNamespace()).
			List(ctx, listOptions)
		if err != nil {
			return err
		}

		if len(buildList.Items) > 0 {
			maxBuildCount := v1alpha1.DefaultBuildRetentionCount
			if configDefaults.BuildRetentionCount != nil {
				maxBuildCount = int(*configDefaults.BuildRetentionCount)
			}

			buildsToDelete := buildsToGC(buildList.Items, maxBuildCount)
			for _, t := range buildsToDelete {
				if err := r.KfClientSet.KfV1alpha1().
					Builds(app.GetNamespace()).
					Delete(ctx, t.Name, metav1.DeleteOptions{}); err != nil {
					return err
				}
			}
		}
	}

	// reconcile Tasks
	{
		logger.Debug("reconciling tasks")
		configDefaults, err := kfconfig.FromContext(ctx).Defaults()
		if err != nil {
			return fmt.Errorf("failed to read config-defaults: %v", err)
		}
		taskLabelSelector := fmt.Sprintf("%s=%s", v1alpha1.NameLabel, app.Name)
		listOptions := metav1.ListOptions{
			LabelSelector: taskLabelSelector,
		}
		taskList, err := r.KfClientSet.KfV1alpha1().Tasks(app.GetNamespace()).List(ctx, listOptions)
		if err != nil {
			return err
		}

		if len(taskList.Items) > 0 {
			tasksToDelete := tasksToGC(taskList.Items, v1alpha1.DefaultMaxTaskCount)

			if configDefaults.TaskRetentionCount != nil {
				tasksToDelete = tasksToGC(taskList.Items, int(*configDefaults.TaskRetentionCount))
			}

			for _, t := range tasksToDelete {
				if err := r.KfClientSet.KfV1alpha1().
					Tasks(app.GetNamespace()).
					Delete(ctx, t.Name, metav1.DeleteOptions{}); err != nil {
					return err
				}
			}
		}
	}

	instanceStatus := app.Spec.Instances.Status()

	// reconcile HorizontalPodAutoscaler
	{
		logger.Debug("reconciling HorizontalPodAutoscaler")
		actualHpa, err := r.reconcileHorizontalPodAutoscaler(ctx, app, app.Namespace, resources.AutoscalerName(app), app.Spec.Instances.Autoscaling)
		if err != nil {
			return err
		}
		// actual can be nil and is expected when deletion of HPA succeeded.
		app.Status.PropagateAutoscalerV1Status(actualHpa)

		// Propagate the human-readable app instances after HPA has been reconciled
		instanceStatus.PropagateAutoscalingStatus(app, actualHpa)
	}

	// Reconcile deployment, this MUST go after routes so VCAP_APPLICATION can
	// get the correct values.
	{
		logger.Debug("reconciling deployment")
		condition := app.Status.DeploymentCondition()
		if err != nil {
			return condition.MarkReconciliationError("Failed to read config-defaults", err)
		}
		desired, err := resources.MakeDeployment(app, space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.deploymentLister.Deployments(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.ReconcileDeployment(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}

		app.Status.PropagateDeploymentStatus(actual)
	}

	// Update the human-readable app instances after the backing service has been
	// synchronized so we always display the current configuration.
	{
		// Set the selectors.
		instanceStatus.LabelSelector = labels.
			SelectorFromSet(resources.PodLabels(app)).
			String()

		app.Status.PropagateInstanceStatus(instanceStatus)
	}

	return nil
}

func (r *Reconciler) reconcileRoute(
	ctx context.Context,
	desired *v1alpha1.Route,
	actual *v1alpha1.Route,
) (*v1alpha1.Route, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "Route").
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
	return r.KfClientSet.
		KfV1alpha1().
		Routes(existing.Namespace).
		Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) reconcileHorizontalPodAutoscaler(
	ctx context.Context,
	app *v1alpha1.App,
	namespace string,
	autoscalerName string,
	autoscalingSpec v1alpha1.AppSpecAutoscaling) (*autoscalingv1.HorizontalPodAutoscaler, error) {

	condition := app.Status.HorizontalPodAutoscalerCondition()
	desired, err := resources.MakeHorizontalPodAutoScaler(app)

	if err != nil {
		return nil, condition.MarkTemplateError(err)
	}

	if desired == nil {
		err := r.KubeClientSet.
			AutoscalingV1().
			HorizontalPodAutoscalers(namespace).
			Delete(ctx, autoscalerName, metav1.DeleteOptions{})
		if apierrs.IsNotFound(err) {
			return nil, nil
		}

		return nil, condition.MarkReconciliationError("deleting", err)
	}

	actual, err := r.autoscalingLister.HorizontalPodAutoscalers(namespace).Get(autoscalerName)
	switch {
	case apierrs.IsNotFound(err):
		actual, err = r.
			KubeClientSet.AutoscalingV1().
			HorizontalPodAutoscalers(desired.Namespace).
			Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			return nil, condition.MarkReconciliationError("creating", err)
		}
		return actual, nil
	case err != nil:
		return nil, condition.MarkReconciliationError("getting latest", err)
	case !metav1.IsControlledBy(actual, app):
		return nil, condition.MarkChildNotOwned("not owned by app")
	}

	if actual, err = r.reconcileAutoscaler(ctx, desired, actual); err != nil {
		return nil, condition.MarkReconciliationError("updating existing", err)
	}

	return actual, nil
}

// ReconcileAutoscaler syncs the existing K8s autoscaler to the desired autoscaler.
func (r *Reconciler) reconcileAutoscaler(ctx context.Context, desired, actual *autoscalingv1.HorizontalPodAutoscaler) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "HorizontalPodAutoscaler").
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
	return r.KubeClientSet.AutoscalingV1().HorizontalPodAutoscalers(existing.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.App) (*v1alpha1.App, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating status")
	actual, err := r.appLister.Apps(desired.GetNamespace()).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().Apps(existing.GetNamespace()).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}

func tasksToGC(tasks []v1alpha1.Task, maxTasks int) []v1alpha1.Task {
	if len(tasks) <= maxTasks {
		return nil
	}

	// Sort tasks first by status (final/non-final), then by creation
	// timestamp (youngest first).  This ensures we GC the oldest completed
	// Tasks.
	sort.Slice(tasks, func(i, j int) bool {
		taskA := tasks[i]
		taskB := tasks[j]
		taskACompleted := v1alpha1.IsStatusFinal(taskA.Status.Status)
		taskBCompleted := v1alpha1.IsStatusFinal(taskB.Status.Status)

		if taskACompleted == taskBCompleted {
			return taskB.CreationTimestamp.Before(&taskA.CreationTimestamp)
		}

		return !taskACompleted
	})

	var tasksToGC []v1alpha1.Task

	// Excess final tasks are deleted.
	tasksOverLimit := tasks[maxTasks:]
	for _, t := range tasksOverLimit {
		if v1alpha1.IsStatusFinal(t.Status.Status) {
			tasksToGC = append(tasksToGC, t)
		}
	}
	return tasksToGC
}

func buildsToGC(builds []v1alpha1.Build, maxBuilds int) []v1alpha1.Build {
	// Only GC final (Succeded=True/False) builds, non-final builds (Succeded=UNKNOWN)
	// will fail after timeout (default 1 hour) and turn into Succeded=False.
	var finalBuilds []v1alpha1.Build
	for _, b := range builds {
		if v1alpha1.IsStatusFinal(b.Status.Status) {
			finalBuilds = append(finalBuilds, b)
		}
	}

	// A minimum of one build should be kept -
	// It continues to be needed for reference (e.g. "kf restage" re-runs the latest build)
	maxBuilds = int(math.Max(1, float64(maxBuilds)))
	if len(finalBuilds) <= maxBuilds {
		return nil
	}

	// Sort builds by creation timestamp (youngest first).
	// This ensures we GC the oldest completed Builds.
	sort.Slice(finalBuilds, func(i, j int) bool {
		return finalBuilds[j].CreationTimestamp.Before(&finalBuilds[i].CreationTimestamp)
	})

	// Excess final builds are deleted.
	return finalBuilds[maxBuilds:]
}
