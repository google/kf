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

package task

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/dockerutil"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/task/resources"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	taskrunclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"
	tektonListers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reconciler reconciles a Task object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	tektonClient taskrunclient.TektonV1beta1Interface

	// listers index properties about resources
	appLister     kflisters.AppLister
	spaceLister   kflisters.SpaceLister
	taskLister    kflisters.TaskLister
	taskRunLister tektonListers.TaskRunLister
	configStore   *config.Store
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

	return r.reconcileTask(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileTask(
	ctx context.Context,
	namespace string,
	name string,
) (err error) {
	logger := logging.FromContext(ctx)

	original, err := r.taskLister.Tasks(namespace).Get(name)

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
			logger.Warnw("Failed to update Task status", zap.Error(uErr))
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

	// ALWAYS update the ObservedGenration: "If the primary resource your
	// controller is reconciling supports ObservedGeneration in its status, make
	// sure you correctly set it to metadata.Generation whenever the values
	// between the two fields mismatches."
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
	toReconcile.Status.ObservedGeneration = toReconcile.Generation

	// Reconcile this copy of the service and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ApplyChanges(ctx, toReconcile, namespace)
	if reconcileErr != nil {
		logger.Debugf("Task reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, namespace, toReconcile); uErr != nil {
		logger.Warnw("Failed to update Task status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the Task.
func (r *Reconciler) ApplyChanges(ctx context.Context, task *v1alpha1.Task, namespace string) error {
	logger := logging.FromContext(ctx)
	task.Status.InitializeConditions()

	// Tasks only get run once regardless of success or failure status.
	if v1alpha1.IsStatusFinal(task.Status.Status) {
		return nil
	}

	// Ensure Kf Space exists to prevent Kf objects from being created in namespaces that's not a Kf Space.
	space, err := r.spaceLister.Get(namespace)
	if err != nil {
		task.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	task.Status.MarkSpaceHealthy()

	appCondition := task.Status.AppCondition()

	app, err := r.appLister.Apps(namespace).Get(task.Spec.AppRef.Name)

	if err != nil {
		return appCondition.MarkTemplateError(err)
	}

	appCondition.MarkSuccess()

	// Update Task ID if not set.
	// Task ID is server generated (not authorized by clients). It is implemented by keeping a counter in
	// the App, and each Task is assigned an App-wide unique ID by incrementing the App counter.
	if task.Status.ID == 0 {
		appUpdate := app.DeepCopy()
		appUpdate.Status.Tasks.UpdateRequests++
		if _, err := r.KfClientSet.KfV1alpha1().Apps(namespace).UpdateStatus(ctx, appUpdate, metav1.UpdateOptions{}); err != nil {
			return err
		}
		task.Status.ID = appUpdate.Status.Tasks.UpdateRequests
	}

	taskUpdate := task.DeepCopy()

	// Update Task DisplayName to be the same as Task Name if not set.
	if len(task.Spec.DisplayName) == 0 {
		taskUpdate.Spec.DisplayName = task.Name
	}

	taskUpdate.Labels = v1alpha1.UnionMaps(taskUpdate.Labels, map[string]string{
		v1alpha1.VersionLabel: fmt.Sprint(task.Status.ID),
	})

	if _, err := r.KfClientSet.KfV1alpha1().Tasks(namespace).Update(ctx, taskUpdate, metav1.UpdateOptions{}); err != nil {
		return err
	}

	configCondition := task.Status.ConfigCondition()

	configDefaults, err := config.FromContext(ctx).Defaults()

	if err != nil {
		return configCondition.MarkTemplateError(err)
	}

	configCondition.MarkSuccess()

	// Sync TaskRun
	{
		// TODO(b/209466387): We have to keep PipelineRunCondition until v2.8
		// for upgrade purproses.
		task.Status.PipelineRunCondition().MarkSuccess()

		taskRunCondition := task.Status.TaskRunCondition()

		// Tekton requires us to have ta command set else the Tekton
		// controller will go and try to fetch the image. This will fail if
		// the user didn't install the controller to be able to read from the
		// container registry (which is likely).
		containerCommand, err := r.fetchContainerCommand(app)
		if err != nil {
			return taskRunCondition.MarkReconciliationError("fetching container command", err)
		}

		desiredTaskRun, err := resources.MakeTaskRun(configDefaults, task, app, space, containerCommand)
		if err != nil {
			return taskRunCondition.MarkTemplateError(err)
		}

		logger.Debug("reconciling TaskRun ", desiredTaskRun.Name)

		actual, err := r.taskRunLister.
			TaskRuns(task.Namespace).
			Get(desiredTaskRun.Name)
		switch {
		case errors.IsNotFound(err):
			actual, err = r.tektonClient.
				TaskRuns(desiredTaskRun.Namespace).
				Create(ctx, desiredTaskRun, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		case !metav1.IsControlledBy(actual, task):
			return taskRunCondition.MarkChildNotOwned(desiredTaskRun.Name)
		}

		if _, err := r.reconcileTaskRun(ctx, desiredTaskRun, actual); err != nil {
			return taskRunCondition.MarkReconciliationError("synchronizing", err)
		}

		task.Status.PropagateTaskStatus(actual)
	}

	return nil
}

// reconcileTaskRun syncs the existing TaskRun to the desired TaskRun.
func (r *Reconciler) reconcileTaskRun(ctx context.Context, desired, actual *tektonv1beta1.TaskRun) (*tektonv1beta1.TaskRun, error) {
	logger := logging.FromContext(ctx)

	// Check for differences, if none we don't need to reconcile.
	if reconciler.NewSemanticEqualityBuilder(logger, "PipelineRun").
		Append("metadata.labels", desired.ObjectMeta.Labels, actual.ObjectMeta.Labels).
		Append("spec.taskSpec", desired.Spec.TaskSpec, actual.Spec.TaskSpec).
		Append("spec.serviceAccountName", desired.Spec.ServiceAccountName, actual.Spec.ServiceAccountName).
		Append("spec.status", desired.Spec.Status, actual.Spec.Status).
		IsSemanticallyEqual() {
		return actual, nil
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	// Preserve the rest of the object (e.g. ObjectMeta except for labels).
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	existing.Spec = desired.Spec

	return r.tektonClient.TaskRuns(actual.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) updateStatus(ctx context.Context, namespace string, desired *v1alpha1.Task) (*v1alpha1.Task, error) {
	actual, err := r.taskLister.Tasks(namespace).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().Tasks(namespace).UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}

func (r *Reconciler) fetchContainerCommand(app *v1alpha1.App) ([]string, error) {
	imageRef, err := name.ParseReference(app.Status.Image, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(imageRef, dockerutil.GetAuthKeyChain())
	if err != nil {
		return nil, err
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return configFile.Config.Entrypoint, nil
}

// maybeCleanupTaskRunPod removes Pod sidecars that were injected by mutating webhooks
// e.g. Istio that Tekton doesn't clean up if the Tekton TaskRun is completed.
func (r *Reconciler) maybeCleanupTaskRunPod(ctx context.Context, task *v1alpha1.Task) {
	logger := logging.FromContext(ctx)

	taskRun, err := r.taskRunLister.
		TaskRuns(task.Namespace).
		Get(resources.TaskRunName(task))
	if err != nil {
		logger.Errorf("Couldn't get TaskRun: %v", err)
		return
	}

	// Don't modify TaskRuns unless they belong to the Task.
	if !metav1.IsControlledBy(taskRun, task) {
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
