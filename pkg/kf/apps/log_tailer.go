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

package apps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/fatih/color"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/logs"
	"github.com/segmentio/textio"
	corev1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type pushLogTailer struct {
	client              *appsClient
	out                 io.Writer
	logger              *log.Logger
	appName             string
	resourceVersion     string
	namespace           string
	noStart             bool
	buildStartTime      time.Time
	deployStartTime     time.Time
	tailBuildLogsMutex  int32
	checkBuildReadyOnce sync.Once
	buildOut            io.Writer
	appTailer           logs.Tailer
	tailAppLogsMutex    int32
	appOut              io.Writer
	startingAppOnce     sync.Once
}

func newPushLogTailer(
	client *appsClient,
	out io.Writer,
	appName string,
	resourceVersion string,
	namespace string,
	noStart bool,
) *pushLogTailer {

	t := &pushLogTailer{
		client:          client,
		out:             out,
		appName:         appName,
		resourceVersion: resourceVersion,
		namespace:       namespace,
		noStart:         noStart,
	}

	logColor := color.New(color.FgGreen)

	t.logger = log.New(out, logColor.Sprintf("[deploy] "), 0)
	t.buildOut = textio.NewPrefixWriter(out, logColor.Sprintf("[build] "))
	t.appOut = textio.NewPrefixWriter(out, logColor.Sprintf("[app] "))
	t.buildStartTime = time.Now()
	return t
}

// DeployLogsForApp gets the deployment logs for an application. It blocks until
// the operation has completed.
func (a *appsClient) DeployLogsForApp(ctx context.Context, out io.Writer, app *v1alpha1.App) error {
	return a.DeployLogs(ctx, out, app.Name, app.ResourceVersion, app.Namespace, app.Spec.Instances.Stopped)
}

// DeployLogs writes the logs for the deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (a *appsClient) DeployLogs(
	ctx context.Context,
	out io.Writer,
	appName string,
	resourceVersion string,
	namespace string,
	noStart bool,
) error {
	ctx, cancel := context.WithCancel(ctx)

	t := newPushLogTailer(a, out, appName, resourceVersion, namespace, noStart)
	defer cancel()

	for ctx.Err() == nil {
		done, err := t.handleWatch(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		// ResourceVersion is set on list, update it before trying to tail again.
		// If we get an error, it's not really worth reporting.
		if apps, err := t.client.kclient.Apps(t.namespace).List(ctx, k8smeta.ListOptions{}); err == nil {
			t.resourceVersion = apps.ResourceVersion
		}
	}

	return ctx.Err()
}

func (t *pushLogTailer) handleWatch(ctx context.Context) (bool, error) {
	ws, err := t.client.kclient.Apps(t.namespace).Watch(ctx, k8smeta.ListOptions{
		ResourceVersion: t.resourceVersion,

		FieldSelector: fields.OneTermEqualSelector("metadata.name", t.appName).String(),
		Watch:         true,
	})
	if err != nil {
		return true, err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		switch e.Type {
		case watch.Error:
			if obj, ok := e.Object.(*k8smeta.Status); ok {
				t.logger.Printf("status error: %s:%s\n", obj.Reason, obj.Message)
			}

			return false, nil

		case watch.Deleted:
			return true, errors.New("App was deleted")

		case watch.Added, watch.Modified:
			switch obj := e.Object.(type) {
			case *v1alpha1.App:
				// skip out of date apps
				if !ObservedGenerationMatchesGeneration(obj) {
					continue
				}

				done, err := t.handleUpdate(ctx, obj)
				if err != nil {
					return true, err
				}
				if done {
					return true, nil
				}
			default:
				t.logger.Printf("Unexpected type in watch stream: %T\n", e.Object)
				continue
			}
		default:
			// NOTE: later versions of apimachinery have added the type BOOKMARK to
			// dynamically update resourceVersion as lists happen.
			t.logger.Printf("got unknown watch event: %v\n", e.Type)
		}
	}

	return false, nil
}

func (t *pushLogTailer) handleUpdate(ctx context.Context, app *v1alpha1.App) (bool, error) {

	buildReady := app.Status.GetCondition(v1alpha1.AppConditionBuildReady)
	if buildReady == nil {
		// build might still be creating
		return false, nil
	}
	if buildReady.Message != "" {
		t.logger.Printf("Updated state to: %s\n", buildReady.Message)
	}

	switch buildReady.Status {
	case corev1.ConditionTrue:
		// Only handle build success case once
		t.checkBuildReadyOnce.Do(func() {
			t.logger.Printf("Built in %0.2f seconds\n", time.Since(t.buildStartTime).Seconds())
			t.deployStartTime = time.Now()
		})
	case corev1.ConditionFalse:
		t.logger.Printf("Failed to build: %s\n", buildReady.Message)
		return true, fmt.Errorf("build failed: %s", buildReady.Message)
	default:

		// This case should mean the Build is still in progress.
		// It should be safe to tail the logs to show the user what's happening.
		t.runIfNotRunning(&t.tailBuildLogsMutex, "tailing Build logs", func() {
			utils.SuggestNextAction(utils.NextAction{
				Description: "View build logs",
				Commands: []string{
					fmt.Sprintf("kf build-logs %s --space %s", app.Status.LatestCreatedBuildName, t.namespace),
				},
			})

			// Pick which build to go look for. If
			// app.Spec.Build.BuildRef.Name is set, then that is the preferred
			// build.
			buildName := app.Status.LatestCreatedBuildName
			if app.Spec.Build.BuildRef != nil {
				buildName = app.Spec.Build.BuildRef.Name
			}

			// ignoring tail errs because they are spurious
			t.client.buildsClient.Tail(ctx, t.namespace, buildName, t.buildOut)
		})

		return false, nil
	}

	if t.noStart {
		t.logger.Printf("Total push time %0.2f seconds\n", time.Since(t.buildStartTime).Seconds())
		return true, nil
	}

	t.startingAppOnce.Do(func() {
		t.logger.Printf("Starting App: %s\n", app.Name)
	})

	appReady := app.Status.GetCondition(v1alpha1.AppConditionReady)
	if appReady == nil {
		return false, nil
	}
	if appReady.Message != "" {
		t.logger.Printf("Updated state to: %s\n", appReady.Message)
	}

	switch appReady.Status {
	case corev1.ConditionTrue:
		t.logger.Printf("App took %0.2f seconds to become ready.\n", time.Since(t.deployStartTime).Seconds())
		t.logger.Printf("Total push time %0.2f seconds\n", time.Since(t.buildStartTime).Seconds())
		return true, nil
	case corev1.ConditionFalse:
		t.logger.Printf("Failed to deploy: %s\n", appReady.Message)
		return true, fmt.Errorf("deployment failed: %s", appReady.Message)
	default:
		// This case should mean the deployment is still in progress.
		// It should be safe to tail the logs to show the user what's happening.
		t.runIfNotRunning(&t.tailAppLogsMutex, "tailing App logs", func() {
			utils.SuggestNextAction(utils.NextAction{
				Description: "View App logs",
				Commands: []string{
					fmt.Sprintf("kf logs %s --space %s", app.Name, t.namespace),
				},
			})

			// ignoring tail errs because they are spurious
			t.client.appTailer.Tail(
				ctx,
				app.Name,
				t.appOut,
				logs.WithTailSpace(t.namespace),
				logs.WithTailFollow(true),
			)
		})
	}

	return false, nil
}

func (t *pushLogTailer) runIfNotRunning(lock *int32, action string, callback func()) {
	if !atomic.CompareAndSwapInt32(lock, 0, 1) {
		// some other routine is running
		return
	}

	go func() {
		defer atomic.StoreInt32(lock, 0) // release lock

		t.logger.Printf("Start %s\n", action)
		callback()
		t.logger.Printf("End %s\n", action)
	}()
}
