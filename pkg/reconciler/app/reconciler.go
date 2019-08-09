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
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/pkg/client/listers/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/systemenvinjector"
	"github.com/google/kf/pkg/reconciler"
	"github.com/google/kf/pkg/reconciler/app/resources"
	"github.com/knative/serving/pkg/apis/autoscaling"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servinglisters "github.com/knative/serving/pkg/client/listers/serving/v1alpha1"
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

type Reconciler struct {
	*reconciler.Base

	knativeServiceLister  servinglisters.ServiceLister
	knativeRevisionLister servinglisters.RevisionLister
	sourceLister          kflisters.SourceLister
	appLister             kflisters.AppLister
	spaceLister           kflisters.SpaceLister
	routeLister           kflisters.RouteLister
	routeClaimLister      kflisters.RouteClaimLister
	systemEnvInjector     systemenvinjector.SystemEnvInjectorInterface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile is called by Kubernetes.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	return r.reconcileApp(ctx, namespace, name, logger)
}

func (r *Reconciler) reconcileApp(ctx context.Context, namespace, name string, logger *zap.SugaredLogger) (err error) {
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
		logger.Warnw("Failed to update Route status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

func (r *Reconciler) ApplyChanges(ctx context.Context, app *v1alpha1.App) error {

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
		r.Logger.Info("reconciling Source")
		condition := app.Status.SourceCondition()
		desired, err := resources.MakeSource(app, space, time.Now().UnixNano())
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.latestSource(app)
		if apierrs.IsNotFound(err) || !r.sourcesAreSemanticallyEqual(desired, actual) {
			// Source doesn't exist or it's for the wrong version, make a new one.
			actual, err = r.KfClientSet.KfV1alpha1().Sources(app.Namespace).Create(desired)
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		}

		app.Status.PropagateSourceStatus(actual)

		if condition.IsPending() {
			r.Logger.Info("Waiting for source; exiting early")
			return nil
		}

	}

	// TODO(josephlewis42) we should grab info to create the VCAP_SERVICES
	// environment variable here and store it in a secret that can be injected.

	// reconcile serving
	{
		r.Logger.Info("reconciling Knative Serving")
		condition := app.Status.KnativeServiceCondition()
		desired, err := resources.MakeKnativeService(app, space, r.systemEnvInjector)
		if err != nil {
			return condition.MarkTemplateError(err)
		}

		actual, err := r.knativeServiceLister.Services(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) && !app.Spec.Instances.Stopped {
			// Knative Service doesn't exist, make one.
			actual, err = r.ServingClientSet.ServingV1alpha1().Services(desired.GetNamespace()).Create(desired)
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
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
			// https://github.com/knative/serving/issues/4098 is resolved.
			if err := r.ServingClientSet.
				ServingV1alpha1().
				Services(desired.Namespace).
				Delete(desired.Name, &metav1.DeleteOptions{}); err != nil {
				return condition.MarkReconciliationError("stopping (via deleting service) existing", err)
			}
		} else if actual, err = r.reconcileKnativeService(desired, actual); err != nil {
			return condition.MarkReconciliationError("updating existing", err)
		}

		app.Status.PropagateKnativeServiceStatus(actual)
	}

	// Routes and RouteClaims
	desiredRoutes, desiredRouteClaims, err := resources.MakeRoutes(app, space)
	condition := app.Status.RouteCondition()
	if err != nil {
		return condition.MarkTemplateError(err)
	}

	// Route Reconciler
	{
		r.Logger.Info("reconciling Routes")

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
				actual, err = r.KfClientSet.KfV1alpha1().Routes(desired.GetNamespace()).Create(&desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if actual, err = r.reconcileRoute(&desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}
		}
	}

	// RouteClaim reconciler
	{
		r.Logger.Info("reconciling Route Claims")

		for _, desired := range desiredRouteClaims {
			actual, err := r.routeClaimLister.
				RouteClaims(desired.GetNamespace()).
				Get(desired.Name)
			if apierrs.IsNotFound(err) {
				// RouteClaim doesn't exist, make one.
				actual, err = r.KfClientSet.
					KfV1alpha1().
					RouteClaims(desired.GetNamespace()).
					Create(&desired)
				if err != nil {
					return condition.MarkReconciliationError("creating", err)
				}
			} else if err != nil {
				return condition.MarkReconciliationError("getting latest", err)
			} else if actual, err = r.reconcileRouteClaim(&desired, actual); err != nil {
				return condition.MarkReconciliationError("updating existing", err)
			}
		}
	}

	// Making it to the bottom of the reconciler means we've synchronized.
	app.Status.ObservedGeneration = app.Generation

	return r.gcRevisions(ctx, app)
}

func (r *Reconciler) latestSource(app *v1alpha1.App) (*v1alpha1.Source, error) {
	// NOTE: this code polls the Kubernetes cluster directly rather than the
	// cache to prevent multiple builds from kicking off.
	selector := resources.MakeSourceLabels(app)
	listOps := metav1.ListOptions{LabelSelector: labels.Set(selector).String()}
	list, err := r.KfClientSet.KfV1alpha1().Sources(app.Namespace).List(listOps)
	if err != nil {
		return nil, err
	}

	items := list.Items

	// sort descending
	sort.Slice(items, func(i int, j int) bool {
		return items[j].CreationTimestamp.Before(&items[i].CreationTimestamp)
	})

	if err == nil && len(items) > 0 {
		return &items[0], nil
	}

	return nil, apierrs.NewNotFound(v1alpha1.Resource("sources"), fmt.Sprintf("source for %s", app.Name))
}

func (*Reconciler) sourcesAreSemanticallyEqual(desired, actual *v1alpha1.Source) bool {
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	return semanticEqual
}

func (r *Reconciler) reconcileKnativeService(desired, actual *serving.Service) (*serving.Service, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff serving: %v", err)
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec
	return r.ServingClientSet.ServingV1alpha1().Services(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileRoute(desired, actual *v1alpha1.Route) (*v1alpha1.Route, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff serving: %v", err)
	}

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

func (r *Reconciler) reconcileRouteClaim(desired, actual *v1alpha1.RouteClaim) (*v1alpha1.RouteClaim, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff serving: %v", err)
	}

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

func (r *Reconciler) updateStatus(desired *v1alpha1.App) (*v1alpha1.App, error) {
	r.Logger.Info("updating status")
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
// TODO: Reevaluate once https://github.com/knative/serving/issues/4183 is
// resolved.
func (r *Reconciler) gcRevisions(ctx context.Context, app *v1alpha1.App) error {
	r.Logger.Debugf("Checking for revisions that need to adjust %s...", autoscaling.MinScaleAnnotationKey)
	defer r.Logger.Debugf("Done checking for revisions that need to adjust %s.", autoscaling.MinScaleAnnotationKey)

	if app.Status.LatestCreatedRevisionName != app.Status.LatestReadyRevisionName {
		r.Logger.Debugf("Not willing to garbage collection Revisions while the latest is not ready...")
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
			r.Logger.Warnf("Revision did not contain ConfigurationGeneration")
			return -1
		}

		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			r.Logger.Warnf("Revision had an invalid ConfigurationGeneration: %s", err)
			return -1
		}
		return x
	}

	// descending
	sort.Slice(revs, func(i int, j int) bool {
		return parseGeneration(revs[j]) < parseGeneration(revs[i])
	})

	// delete everything after the latest generation
	for _, rev := range revs[1:] {
		r.Logger.Infof("Garbage collecting Revision %s...", rev.Name)
		if err := revisionClient.Delete(rev.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
