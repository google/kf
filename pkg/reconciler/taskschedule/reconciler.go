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

package taskschedule

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	"github.com/google/kf/v2/pkg/reconciler/taskschedule/resources"
	werrors "github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

// Reconciler reconciles a TaskSchedule object with the K8s cluster.
type Reconciler struct {
	*reconciler.Base

	spaceLister        kflisters.SpaceLister
	taskScheduleLister kflisters.TaskScheduleLister
	taskLister         kflisters.TaskLister
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

	return r.reconcileTaskSchedule(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) reconcileTaskSchedule(ctx context.Context, namespace, name string) error {
	logger := logging.FromContext(ctx)

	original, err := r.taskScheduleLister.TaskSchedules(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
	case err != nil:
		return err
	case original.GetDeletionTimestamp() != nil:
		logger.Info("resource deletion requested")
		toUpdate := original.DeepCopy()
		toUpdate.Status.PropagateTerminatingStatus()
		if _, uErr := r.updateStatus(ctx, toUpdate); uErr != nil {
			logger.Warnw("Failed to update TaskSchedule status", zap.Error(uErr))
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
		logger.Debugf("TaskSchedule reconcilerErr is not empty: %+v", reconcileErr)
	}
	if equality.Semantic.DeepEqual(original.Status, toReconcile.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, toReconcile); uErr != nil {
		logger.Warnw("Failed to update TaskSchedule status", zap.Error(uErr))
		return uErr
	}

	return reconcileErr
}

// isTaskFinished checks if the Task has completed, either successfully or in failure.
func isTaskFinished(task *v1alpha1.Task) bool {
	succeeded := task.Status.GetCondition(v1alpha1.TaskConditionSucceeded)
	return succeeded.IsFalse() || succeeded.IsTrue()
}

// inActiveList checks if the Task is in the TaskSchedule's status.active list.
func inActiveList(ts *v1alpha1.TaskSchedule, task *v1alpha1.Task) bool {
	for _, t := range ts.Status.Active {
		if t.Name == task.Name {
			return true
		}
	}
	return false
}

func removeFromActive(ts *v1alpha1.TaskSchedule, task *v1alpha1.Task) {
	index := -1
	for i, t := range ts.Status.Active {
		if t.Name == task.Name {
			index = i
			break
		}
	}
	if index >= 0 {
		length := len(ts.Status.Active)
		ts.Status.Active[index] = ts.Status.Active[length-1]
		ts.Status.Active = ts.Status.Active[:length-1]
	}
}

// ApplyChanges updates the linked resources in the cluster with the current
// status of the TaskSchedule.
func (r *Reconciler) ApplyChanges(ctx context.Context, ts *v1alpha1.TaskSchedule) error {
	logger := logging.FromContext(ctx)
	ts.Status.InitializeConditions()

	// Ensure Kf Space exists to prevent Kf objects from being created in namespaces that's not a Kf Space.
	_, err := r.spaceLister.Get(ts.GetNamespace())
	if err != nil {
		ts.Status.MarkSpaceUnhealthy("GettingSpace", err.Error())
		return err
	}
	ts.Status.MarkSpaceHealthy()

	tasks, err := r.getChildren(ts)
	if err != nil {
		logger.Warnw("Failed to get child tasks", zap.Error(err))
		return err
	}

	for _, task := range tasks {
		if inActiveList(ts, task) && isTaskFinished(task) {
			removeFromActive(ts, task)
		}
	}

	// Skip reconciliation if TaskSchedule is suspended.
	if !ts.Spec.Suspend {
		return r.scheduleTask(ctx, ts)
	}

	return nil
}

func (r *Reconciler) scheduleTask(ctx context.Context, ts *v1alpha1.TaskSchedule) error {
	logger := logging.FromContext(ctx)
	sched, err := cron.ParseStandard(ts.Spec.Schedule)
	if err != nil {
		ts.Status.MarkScheduleError(werrors.Wrap(err, "Failed to parse schedule"))
		return err
	}

	scheduledTime, err := getNextScheduleTime(*ts, time.Now(), sched)
	if err != nil {
		ts.Status.MarkScheduleError(werrors.Wrap(err, "Failed to find next schedule time"))
		return err
	}

	if scheduledTime == nil {
		// There have been no missed execution times since last run.
		return nil
	}

	// Check ConcurrencyPolicy to decide how to handle existing running tasks.
	if ts.Spec.ConcurrencyPolicy != v1alpha1.ConcurrencyPolicyAlways && len(ts.Status.Active) > 0 {
		switch {
		case ts.Spec.ConcurrencyPolicy == v1alpha1.ConcurrencyPolicyForbid:
			logger.Infof("Skipping scheduling Task as previous Task is still running and ConcurrencyPolicy is Forbid")
			return nil
		case ts.Spec.ConcurrencyPolicy == v1alpha1.ConcurrencyPolicyReplace:
			logger.Infof("Terminating previous Task(s) as they are still running and ConcurrencyPolicy is Replace")
			for _, t := range ts.Status.Active {
				if err := r.terminateTask(ctx, ts.GetNamespace(), t.Name); err != nil {
					logger.Infof("FOO task termination failed: %v", err)
					return err
				}
			}
		}
	}

	desiredTask := resources.MakeTask(ts, *scheduledTime)

	_, err = r.taskLister.
		Tasks(desiredTask.GetNamespace()).
		Get(desiredTask.Name)

	if errors.IsNotFound(err) {
		logger.Infof("Creating Task on schedule %q at scheduled time %s", ts.Spec.Schedule, scheduledTime)
		_, err = r.KfClientSet.
			KfV1alpha1().
			Tasks(desiredTask.GetNamespace()).
			Create(ctx, desiredTask, metav1.CreateOptions{})
		switch {
		case errors.IsAlreadyExists(err):
			logger.Warnf("Task %q already exists", desiredTask.Name)
		case err != nil:
			logger.Warnw("Failed to create Task", zap.Error(err))
			return err
		case err == nil:
			ts.Status.LastScheduleTime = &metav1.Time{Time: *scheduledTime}
			ts.Status.Active = append(ts.Status.Active, v1.LocalObjectReference{
				Name: desiredTask.Name,
			})
		}
	}

	return nil
}

func (r *Reconciler) getChildren(ts *v1alpha1.TaskSchedule) ([]*v1alpha1.Task, error) {
	req, err := labels.NewRequirement(resources.OwningTaskSchedule, selection.Equals, []string{ts.Name})
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*req)
	return r.taskLister.Tasks(ts.GetNamespace()).List(selector)
}

func (r *Reconciler) terminateTask(ctx context.Context, namespace, name string) error {
	mergePatch := map[string]interface{}{
		"spec": map[string]interface{}{
			"terminated": true,
		},
	}
	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}
	_, err = r.KfClientSet.KfV1alpha1().
		Tasks(namespace).
		Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}

func getNextScheduleTime(ts v1alpha1.TaskSchedule, now time.Time, schedule cron.Schedule) (*time.Time, error) {
	var earliestTime time.Time
	if ts.Status.LastScheduleTime != nil {
		earliestTime = ts.Status.LastScheduleTime.Time
	} else {
		// Either a recently created TaskSchedule, or we have not seen a
		// recently created Task. Use the creation time of the TaskSchedule as
		// the last known start time.
		earliestTime = ts.ObjectMeta.CreationTimestamp.Time
	}

	if earliestTime.After(now) {
		// Nothing to start
		return nil, nil
	}

	t1 := schedule.Next(earliestTime)
	t2 := schedule.Next(t1)

	// If now is before the first schedule time, return nil because there isn't a schedule time before now.
	if now.Before(t1) {
		return nil, nil
	}

	// If now is between the first and the second schedule times, the most the recent schedule time is the first one.
	if now.Before(t2) {
		return &t1, nil
	}

	secondsBetweenTwoSchedules := int64(t2.Sub(t1).Round(time.Second).Seconds())
	if secondsBetweenTwoSchedules < 1 {
		return nil, fmt.Errorf("time difference between two schedules less than 1 second")
	}
	secondsElapsed := int64(now.Sub(t1).Seconds())
	numberOfMissedSchedules := (secondsElapsed / secondsBetweenTwoSchedules) + 1
	next := time.Unix(t1.Unix()+((numberOfMissedSchedules-1)*secondsBetweenTwoSchedules), 0).UTC()

	return &next, nil
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.TaskSchedule) (*v1alpha1.TaskSchedule, error) {
	actual, err := r.taskScheduleLister.TaskSchedules(desired.GetNamespace()).Get(desired.Name)
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

	return r.KfClientSet.KfV1alpha1().
		TaskSchedules(existing.GetNamespace()).
		UpdateStatus(ctx, existing, metav1.UpdateOptions{})
}
