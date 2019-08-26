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
	"github.com/knative/serving/pkg/apis/autoscaling"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servinglisters "github.com/knative/serving/pkg/client/listers/serving/v1alpha1"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1listers "k8s.io/client-go/listers/core/v1"
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
	secretLister          v1listers.SecretLister
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
		logger.Debugf("app %q no longer exists\n", name)
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
		} else if !r.sourcesAreSemanticallyEqual(desired, actual) {
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
			} else if actual, err = r.reconcileServiceBinding(&desired, actual); err != nil {
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

		actual, err := r.secretLister.Secrets(desired.GetNamespace()).Get(desired.Name)
		if apierrs.IsNotFound(err) {
			actual, err = r.KubeClientSet.CoreV1().Secrets(desired.GetNamespace()).Create(desired)
			if err != nil {
				return condition.MarkReconciliationError("creating", err)
			}
		} else if err != nil {
			return condition.MarkReconciliationError("getting latest", err)
		} else if !metav1.IsControlledBy(actual, app) {
			return condition.MarkChildNotOwned(desired.Name)
		} else if actual, err = r.reconcileSecret(desired, actual); err != nil {
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
			// https://github.com/knative/serving/issues/4098 is resolved.
			if err := r.ServingClientSet.
				ServingV1alpha1().
				Services(desired.Namespace).
				Delete(desired.Name, &metav1.DeleteOptions{}); err != nil {
				return condition.MarkReconciliationError(
					"stopping (via deleting service) existing",
					err,
				)
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
		logger.Debug("reconciling Route Claims")

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

func (r *Reconciler) reconcileSecret(desired, actual *v1.Secret) (*v1.Secret, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Data, actual.Data)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Data, actual.Data); err != nil {
		return nil, fmt.Errorf("failed to diff secret: %v", err)
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Data = desired.Data
	return r.KubeClientSet.CoreV1().Secrets(existing.Namespace).Update(existing)
}

func (r *Reconciler) reconcileServiceBinding(desired, actual *servicecatalogv1beta1.ServiceBinding) (*servicecatalogv1beta1.ServiceBinding, error) {
	// Check for differences, if none we don't need to reconcile.
	semanticEqual := equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, actual.ObjectMeta.Labels)
	semanticEqual = semanticEqual && equality.Semantic.DeepEqual(desired.Spec, actual.Spec)

	if semanticEqual {
		return actual, nil
	}

	if _, err := kmp.SafeDiff(desired.Spec, actual.Spec); err != nil {
		return nil, fmt.Errorf("failed to diff binding: %v", err)
	}

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
// TODO: Reevaluate once https://github.com/knative/serving/issues/4183 is
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

	// delete everything after the latest generation
	for _, rev := range revs[1:] {
		logger.Infof("Garbage collecting Revision %s...", rev.Name)
		if err := revisionClient.Delete(rev.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
