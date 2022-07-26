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
	"time"

	"github.com/google/kf/v2/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	taskinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/task"
	taskscheduleinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/taskschedule"
)

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := reconciler.NewControllerLogger(ctx, "taskschedules.kf.dev")

	spaceInformer := spaceinformer.Get(ctx)
	taskScheduleInformer := taskscheduleinformer.Get(ctx)
	taskInformer := taskinformer.Get(ctx)

	c := &Reconciler{
		Base:               reconciler.NewBase(ctx, cmw),
		spaceLister:        spaceInformer.Lister(),
		taskScheduleLister: taskScheduleInformer.Lister(),
		taskLister:         taskInformer.Lister(),
	}

	impl := controller.NewContext(ctx, c, controller.ControllerOptions{WorkQueueName: "taskschedules", Logger: logger})

	// Enqueue all TaskSchedules every 10 seconds to check cron intervals and
	// spawn Tasks
	taskScheduleInformer.Informer().AddEventHandlerWithResyncPeriod(
		controller.HandleAll(impl.Enqueue),
		10*time.Second,
	)

	return impl
}
