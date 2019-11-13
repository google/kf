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
	"reflect"
	"sort"
	"strconv"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned"
	servicecataloglisters "github.com/google/kf/pkg/client/servicecatalog/listers/servicecatalog/v1beta1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/cfutil"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/google/kf/third_party/knative-serving/pkg/apis/autoscaling"
	serving "github.com/google/kf/third_party/knative-serving/pkg/apis/serving/v1alpha1"
	servinglisters "github.com/google/kf/third_party/knative-serving/pkg/client/listers/serving/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
)

var (
	restageNeededErr = errors.New("a restage is needed to reflect the latest build settings")
)

type Reconciler struct {
	*reconciler.Base

	serviceCatalogClient  servicecatalogclient.Interface
	knativeServiceLister  servinglisters.ServiceLister
	knativeRevisionLister servinglisters.RevisionLister
	sourceLister          kflisters.SourceLister
	appLister             kflisters.AppLister
	spaceLister           kflisters.SpaceLister
	routeLister           kflisters.RouteLister
	routeClaimLister      kflisters.RouteClaimLister
	serviceBindingLister  servicecataloglisters.ServiceBindingLister
	serviceInstanceLister servicecataloglisters.ServiceInstanceLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by Kubernetes.
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
		logger.Errorf("app %q no longer exists\n", name)
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		return nil
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Errorf("skipping sync for app %q, namespace %q is terminating\n", name, namespace)
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

	app.Status.InitializeConditions()

	space, err := r.spaceLister.Get(app.Namespace)
	switch {
	case apierrs.IsNotFound(err):
		space = &v1alpha1.Space{}
		space.SetDefaults(context.Background())
	case err != nil:
		app.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	app.Status.MarkSpaceHealthy()

	// reconcile source
	{
		logger.Debug("reconciling Source")
		condition := app.Status.SourceCondition()
		desired, err := resources.MakeSource(app, space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.sourceLister.Sources(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			// Source doesn't exist, create a new one
			actual, err = r.KfClientSet.KfV1alpha1().Sources(desired.GetNamespace()).Create(desired)
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if !r.sourcesAreSemanticallyEqual(ctx, desired, actual) {
			// This condition happens if properties the operator configures changes
			// after an app is deployed. For example, the builder image or container
			// registry.
			//
			// We don't want all the apps to automatically rebuild because the update
			// might be partial or breaking. Instead it should be the job of a person
			// or process to update all the apps after something like that.
			return condition.MarkReconciliationError("synchronizing", restageNeededErr)
		}

		app.Status.PropagateSourceStatus(actual)

		if condition.IsPending() {
			logger.Info("Waiting for source; exiting early")
			return nil
		}
	}

	// reconcile service bindings
	var actualServiceBindings []servicecatalogv1beta1.ServiceBinding
	{
		desiredServiceBindings, err := resources.MakeServiceBindings(app)
		condition := app.Status.ServiceBindingCondition()
		if err != nil {
			return condition.MarkTemplateError(err)
		}
		logger.Debug("reconciling Service Bindings")

		// Delete Stale Service Bindings
		existing, err := r.serviceBindingLister.
			ServiceBindings(app.GetNamespace()).
			List(resources.MakeServiceBindingAppSelector(app.Name))
		if err != nil {
			return condition.MarkReconciliationError("scanning for stale service bindings", err)
		}

		// Search to see if any of the existing bindings are not in the desired
		// list of and therefore stale. If they are, delete them.
		for _, binding := range existing {
			if algorithms.Search(
				0,
				v1alpha1.ServiceBindings{*binding},
				v1alpha1.ServiceBindings(desiredServiceBindings),
			) {
				continue
			}

			// Not found in desired, must be stale.
			if err := r.serviceCatalogClient.
				ServicecatalogV1beta1().
				ServiceBindings(binding.Namespace).
				Delete(binding.Name, &metav1.DeleteOptions{}); err != nil {
				return condition.MarkReconciliationError("deleting existing service binding", err)
			}
		}

		for _, desired := range desiredServiceBindings {
			actual, err := r.serviceBindingLister.
				ServiceBindings(desired.GetNamespace()).
				Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// ServiceBindings doesn't exist, make one.
				actual, err = r.serviceCatalogClient.
					ServicecatalogV1beta1().
					ServiceBindings(desired.GetNamespace()).
					Create(&desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if actual, err = r.reconcileServiceBinding(ctx, &desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}
			actualServiceBindings = append(actualServiceBindings, *actual)
		}
		app.Status.PropagateServiceBindingsStatus(actualServiceBindings)
		if condition.IsPending() {
			logger.Info("Waiting for service bindings; exiting early")
			return nil
		}
	}

	// Reconcile VCAP env vars secret
	{
		logger.Debug("reconciling env vars secret")
		condition := app.Status.EnvVarSecretCondition()
		systemEnvInjector := cfutil.NewSystemEnvInjector(r.serviceCatalogClient, r.KubeClientSet)
		desired, err := resources.MakeKfInjectedEnvSecret(app, space, actualServiceBindings, systemEnvInjector)

		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.SecretLister.Secrets(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Secrets(desired.GetNamespace()).Create(desired)
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

	// reconcile serving
	{
		logger.Debug("reconciling Knative Serving")
		condition := app.Status.KnativeServiceCondition()
		desired, err := resources.MakeKnativeService(app, space)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.knativeServiceLister.
			Services(desired.GetNamespace()).
			Get(desired.Name)
		if apierrs.IsNotFound(err) {
			if !app.Spec.Instances.Stopped {
				// Knative Service doesn't exist, make one.
				actual, err = r.ServingClientSet.
					ServingV1alpha1().
					Services(desired.GetNamespace()).
					Create(desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if app.Spec.Instances.Stopped {
			// Found service for stopped app. We delete the service otherwise
			// knative will bring back a single pod, even if when we set
			// scaling to 0:
			// TODO: Reevaluate once
			// https://github.com/knative/serving/issues/4098
			if err := r.ServingClientSet.
				ServingV1alpha1().
				Services(desired.Namespace).
				Delete(desired.Name, &metav1.DeleteOptions{}); err != nil {
				return condition.MarkReconciliationError(
					"stopping (via deleting service) existing",
					err,
				)
			}
		} else if actual, err = r.reconcileKnativeService(ctx, desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}

		app.Status.PropagateKnativeServiceStatus(actual)
	}

	// Update the human-readable app instances after the backing service has been
	// synchronized so we always display the current configuration.
	app.Status.PropagageInstanceStatus(app.Spec.Instances.Status())

	// Routes and RouteClaims
	desiredRoutes, desiredRouteClaims, err := resources.MakeRoutes(app, space)
	condition := app.Status.RouteCondition()
	if err != nil {
		return condition.MarkTemplateError(err)
	}

	// Route Reconciler
	{
		logger.Debug("reconciling Routes")

		// Delete Stale Routes
		existingRoutes, err := r.routeLister.
			Routes(app.GetNamespace()).
			List(resources.MakeRouteAppSelector(app))
		if err != nil {
			return condition.MarkReconciliationError("scanning for stale routes", err)
		}

		// Search to see if any of the existing routes are not in the desired
		// list of routes and therefore stale. If they are, delete them.
		for _, route := range existingRoutes {
			if algorithms.Search(
				0,
				v1alpha1.Routes{*route},
				v1alpha1.Routes(desiredRoutes),
			) {
				continue
			}

			// Not found in desired, must be stale.

			if err := r.KfClientSet.
				KfV1alpha1().
				Routes(route.GetNamespace()).
				Delete(route.Name, &metav1.DeleteOptions{}); err != nil {
				return condition.MarkReconciliationError("deleting existing route", err)
			}
		}

		for _, desired := range desiredRoutes {
			actual, err := r.routeLister.Routes(desired.GetNamespace()).Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// Route doesn't exist, make one.
				_, err = r.KfClientSet.KfV1alpha1().Routes(desired.GetNamespace()).Create(&desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if _, err = r.reconcileRoute(ctx, &desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}
		}
	}

	// RouteClaim reconciler
	{
		logger.Debug("reconciling Route Claims")

		for _, desired := range desiredRouteClaims {
			actual, err := r.routeClaimLister.
				RouteClaims(desired.GetNamespace()).
				Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// RouteClaim doesn't exist, make one.
				_, err = r.KfClientSet.
					KfV1alpha1().
					RouteClaims(desired.GetNamespace()).
					Create(&desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if _, err = r.reconcileRouteClaim(ctx, &desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}
		}
	}

	// If there are no errors reconciling Routes and RouteClaims, mark RouteReady as true
	app.Status.PropagateRouteStatus(desiredRoutes)

	return r.gcRevisions(ctx, app)
}

func (*Reconciler) sourcesAreSemanticallyEqual(
	ctx context.Context,
	desired *v1alpha1.Source,
	actual *v1alpha1.Source,
) bool {
	logger := logging.FromContext(ctx)

	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return true
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		logger.Warnf("failed to diff Source: %v", err)
		return false
	}
	logger.Debug("Source.Spec diff:", diff)

	return false
}

func (r *Reconciler) reconcileKnativeService(
	ctx context.Context,
	desired *serving.Service,
	actual *serving.Service,
) (*serving.Service, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff serving: %v", err)
	}
	logger.Debug("Service.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.ServingClientSet.ServingV1alpha1().Services(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileRoute(
	ctx context.Context,
	desired *v1alpha1.Route,
	actual *v1alpha1.Route,
) (*v1alpha1.Route, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff Route: %v", err)
	}
	logger.Debug("Route.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.KfClientSet.
		KfV1alpha1().
		Routes(existing.Namespace).
		Update(existing)
}

func (r *Reconciler) reconcileRouteClaim(
	ctx context.Context,
	desired *v1alpha1.RouteClaim,
	actual *v1alpha1.RouteClaim,
) (*v1alpha1.RouteClaim, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff RouteClaim: %v", err)
	}
	logger.Debug("RouteClaim.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.KfClientSet.
		KfV1alpha1().
		RouteClaims(existing.Namespace).
		Update(existing)
}

func (r *Reconciler) reconcileServiceBinding(
	ctx context.Context,
	desired *servicecatalogv1beta1.ServiceBinding,
	actual *servicecatalogv1beta1.ServiceBinding,
) (*servicecatalogv1beta1.ServiceBinding, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	diff, err := kmp.SafeDiff(desired.Spec, actual.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to diff binding: %v", err)
	}
	logger.Debug("ServiceBinding.Spec diff:", diff)

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.serviceCatalogClient.
		ServicecatalogV1beta1().
		ServiceBindings(existing.Namespace).
		Update(existing)
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

	return r.KfClientSet.KfV1alpha1().Apps(existing.GetNamespace()).UpdateStatus(existing)
}

// gcRevisions is necessary because Knative won't scale down revisions
// that have a `minScale` greater than 0. Therefore we are going to delete the
// older revisions. The revisions are keeping pods around when app has been
// scaled up. Therefore, if we don't GC the revisions, we leak pods.
// TODO: Reevaluate once https://github.com/google/kf/third_party/knative-serving//issues/4183 is
// resolved.
func (r *Reconciler) gcRevisions(ctx context.Context, app *v1alpha1.App) error {
	logger := logging.FromContext(ctx)
	logger.Debugf("Checking for revisions that need to adjust %s...", autoscaling.MinScaleAnnotationKey)
	defer logger.Debugf("Done checking for revisions that need to adjust %s.", autoscaling.MinScaleAnnotationKey)

	if app.Status.LatestCreatedRevisionName != app.Status.LatestReadyRevisionName {
		logger.Debugf("Not willing to garbage collection Revisions while the latest is not ready...")
		return nil
	}

	selector := labels.Set{"serving.knative.dev/configuration": app.Name}.AsSelector()
	revs, err := r.knativeRevisionLister.Revisions(app.Namespace).List(selector)
	if err != nil {
		return err
	}

	if len(revs) == 0 {
		return nil
	}

	revisionClient := r.ServingClientSet.ServingV1alpha1().Revisions(app.Namespace)

	parseGeneration := func(rev *serving.Revision) int64 {
		v, ok := rev.Labels["serving.knative.dev/configurationGeneration"]
		if !ok {
			logger.Warnf("Revision did not contain ConfigurationGeneration")
			return -1
		}

		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			logger.Warnf("Revision had an invalid ConfigurationGeneration: %s", err)
			return -1
		}
		return x
	}

	// descending
	sort.Slice(revs, func(i int, j int) bool {
		return parseGeneration(revs[j]) < parseGeneration(revs[i])
	})

	// Find the latest generation that is ready
	firstReadyIdx := -1
	for i, rev := range revs {
		if !rev.Status.IsReady() {
			continue
		}
		firstReadyIdx = i
		break
	}

	if firstReadyIdx < 0 {
		// Didn't find any ready revisions. Move on
		return nil
	}

	// delete everything after the latest ready generation
	for _, rev := range revs[firstReadyIdx+1:] {
		logger.Infof("Garbage collecting Revision %s...", rev.Name)
		if err := revisionClient.Delete(rev.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
