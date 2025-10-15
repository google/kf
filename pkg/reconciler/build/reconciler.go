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

package build

import (
	"context"
	"fmt"
	"reflect"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	"github.com/google/kf/v2/pkg/reconciler/build/resources"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	taskrunclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"
	taskrunlisters "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles a build object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	tektonClient taskrunclient.TektonV1beta1Interface

	// listers index properties about resources
	buildLister         kflisters.BuildLister
	spaceLister         kflisters.SpaceLister
	sourcePackageLister kflisters.SourcePackageLister
	taskRunLister       taskrunlisters.TaskRunLister
	kfConfigStore       *kfconfig.Store
	configStore         *config.Store
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

	ctx = r.configStore.ToContext(ctx)
	ctx = r.kfConfigStore.ToContext(ctx)

	return r.reconcileBuild(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileBuild(
	ctx context.Context,
	namespace string,
	name string,
) (err error) {
	logger := logging.FromContext(ctx)

	original, err := r.buildLister.Builds(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		return err

	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.PropagateTerminatingStatus()
		if _, uErr := r.updateStatus(ctx, namespace, toUpdate); uErr != nil {
			logger.Warnw("Failed to update Build status", zap.Error(uErr))
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

	// Clean up the TaskRun's backing Pod if necessary.
	r.maybeCleanupTaskRunPod(ctx, toReconcile)

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile)
	if reconcileErr != nil {
		logger.Debugf("Build reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, namespace, toReconcile); uErr != nil {
		logger.Warnw("Failed to update Build status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the build.
func (r *Reconciler) ApplyChanges(ctx context.Context, build *v1alpha1.Build) error {
	logger := logging.FromContext(ctx)
	build.Status.InitializeConditions()

	// Builds only get run once regardless of success or failure status.
	if v1alpha1.IsStatusFinal(build.Status.Status) {
		return nil
	}

	// Ensure Kf Space exists to prevent Kf objects from being created in namespaces that's not a Kf Space.
	space, err := r.spaceLister.Get(build.Namespace)
	if err != nil {
		build.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	build.Status.MarkSpaceHealthy()

	// Ensure the SourcePackage (if any) is ready.
	var sourcePackage *v1alpha1.SourcePackage
	sourcePackageCondition := build.Status.SourcePackageCondition()
	if sourcePackageName := build.Spec.SourcePackage.Name; sourcePackageName != "" {
		// Has SourcePackage, ensure we wait for it to be ready.
		sourcePackage, err = r.sourcePackageLister.
			SourcePackages(build.Namespace).
			Get(sourcePackageName)
		if errors.IsNotFound(err) {
			// SourcePackage wasn't found, maybe it hasn't been created yet or
			// the cache hasn't been updated.
			logger.Info("Waiting for SourcePackage upload; exiting early")
			return nil
		} else if err != nil {
			// Failed to find SourcePackage.
			return sourcePackageCondition.MarkReconciliationError(fmt.Sprintf("getting SourcePackage %s", sourcePackageName), err)
		}

		build.Status.PropagateSourcePackageStatus(sourcePackage)

		if sourcePackageCondition.IsPending() {
			logger.Info("Waiting for SourcePackage upload; exiting early")
			return nil
		}
	} else {
		// Does NOT have a SourcePackage, move on.
		sourcePackageCondition.MarkSuccess()
	}

	buildCondition := build.Status.TaskRunCondition()

	var taskSpec *tektonv1beta1.TaskSpec
	switch build.Spec.Kind {
	case string(tektonv1beta1.NamespacedTaskKind):
		task, err := r.tektonClient.Tasks(build.Namespace).Get(ctx, build.Spec.Name, metav1.GetOptions{})
		if err != nil {
			return buildCondition.MarkReconciliationError(fmt.Sprintf("getting %s", tektonv1beta1.NamespacedTaskKind), err)
		}
		taskSpec = &task.Spec
	case v1alpha1.BuiltinTaskKind:
		configDefaults, err := kfconfig.FromContext(ctx).Defaults()
		if err != nil {
			return buildCondition.MarkReconciliationError(fmt.Sprintf("getting %s", v1alpha1.BuiltinTaskKind), err)
		}

		secretsConfig, err := config.FromContext(ctx).Secrets()
		if err != nil {
			return buildCondition.MarkReconciliationError("getting secrets config", err)
		}

		taskSpec = resources.FindBuiltinTask(configDefaults, build.Spec, secretsConfig.GoogleServiceAccount)
	default:
		taskSpec = nil
	}

	secretsConfig, err := config.FromContext(ctx).Secrets()
	if err != nil {
		return buildCondition.MarkReconciliationError("getting secrets config", err)
	}
	defaultsConfig, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		return buildCondition.MarkReconciliationError("getting defaults config", err)
	}

	desiredTaskRun, err := resources.MakeTaskRun(
		build,
		taskSpec,
		space,
		sourcePackage,
		secretsConfig,
		defaultsConfig,
	)
	if err != nil {
		return buildCondition.MarkTemplateError(err)
	}

	// Sync TaskRun
	{
		logger.Debug("reconciling TaskRun")

		actual, err := r.taskRunLister.
			TaskRuns(build.Namespace).
			Get(desiredTaskRun.Name)
		if errors.IsNotFound(err) {
			actual, err = r.tektonClient.
				TaskRuns(desiredTaskRun.Namespace).
				Create(ctx, desiredTaskRun, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else if !metav1.IsControlledBy(actual, build) {
			return buildCondition.MarkChildNotOwned(desiredTaskRun.Name)
		}

		build.Status.PropagateBuildStatus(actual)
	}

	return nil
}

// maybeCleanupTaskRunPod removes Pod sidecars that were injected by mutating webhooks
// e.g. Istio that Tekton doesn't clean up if the Tekton TaskRun is completed.
func (r *Reconciler) maybeCleanupTaskRunPod(ctx context.Context, build *v1alpha1.Build) {
	logger := logging.FromContext(ctx)

	taskRun, err := r.taskRunLister.
		TaskRuns(build.Namespace).
		Get(resources.TaskRunName(build))
	if err != nil {
		logger.Errorf("Couldn't get TaskRun: %v", err)
		return
	}

	// Don't modify TaskRuns unless they belong to the Build.
	if !metav1.IsControlledBy(taskRun, build) {
		return
	}

	// Tekton no longer terminates sidecars unless they're explicitly added
	// to the TaskSpec. We need to terminate them so the Pods don't end up
	// with one or two sidecars running indefinitely.
	// https://github.com/tektoncd/pipeline/issues/4731
	if err := r.CleanupCompletedTaskRunSidecars(ctx, taskRun); err != nil {
		// Cleaning up sidecars should be best-effort because it frees up resources from
		// completed TaskRuns, it's fine to retry again later.
		logger.Errorf("Couldn't clean up sidecars: %v", err)
	}
}

func (r *Reconciler) updateStatus(ctx context.Context, namespace string, desired *v1alpha1.Build) (*v1alpha1.Build, error) {
	actual, err := r.buildLister.Builds(namespace).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().Builds(namespace).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}
