// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package garbagecollector

import (
	"context"
	"fmt"
	"math"
	"sort"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/reconciler"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

type Reconciler struct {
	*reconciler.Base

	buildLister kflisters.BuildLister
	appLister   kflisters.AppLister

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

	return r.garbageCollectApp(
		logging.WithLogger(ctx,
			logging.FromContext(ctx).With("namespace", namespace)),
		namespace,
		name,
	)
}

func (r *Reconciler) garbageCollectApp(ctx context.Context, namespace, name string) (err error) {
	logger := logging.FromContext(ctx)
	app, err := r.appLister.Apps(namespace).Get(name)
	switch {
	case apierrs.IsNotFound(err):
		logger.Info("resource no longer exists")
		return nil

	case err != nil:
		return err
	}

	if r.IsNamespaceTerminating(namespace) {
		logger.Info("namespace is terminating, skipping reconciliation")
		return nil
	}

	configDefaults, err := kfconfig.FromContext(ctx).Defaults()
	if err != nil {
		return fmt.Errorf("failed to read config-defaults: %v", err)
	}

	// GC'ing Builds
	{
		logger.Debug("GC'ing Builds for app: %s", app.Name)

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

	// GC'ing Tasks
	{
		logger.Debug("garbage collecting tasks for app: %s", app.Name)
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

	return nil
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
