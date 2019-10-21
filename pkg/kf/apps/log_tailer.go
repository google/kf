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
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type pushLogTailer struct {
	client               *appsClient
	out                  io.Writer
	logger               *log.Logger
	appName              string
	resourceVersion      string
	namespace            string
	noStart              bool
	buildStartTime       time.Time
	deployStartTime      time.Time
	ctx                  context.Context
	ctxCancel            func()
	tailBuildLogsOnce    sync.Once
	checkSourceReadyOnce sync.Once
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

	t.logger = log.New(out, "\033[32m[build]\033[0m ", 0)
	t.logger.Printf("Starting app: %s\n", appName)
	t.buildStartTime = time.Now()
	t.ctx, t.ctxCancel = context.WithCancel(context.Background())
	return t
}

// DeployLogsForApp gets the deployment logs for an application. It blocks until
// the operation has completed.
func (a *appsClient) DeployLogsForApp(out io.Writer, app *v1alpha1.App) error {
	return a.DeployLogs(out, app.Name, app.ResourceVersion, app.Namespace, app.Spec.Instances.Stopped)
}

// DeployLogs writes the logs for the deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (a *appsClient) DeployLogs(
	out io.Writer,
	appName string,
	resourceVersion string,
	namespace string,
	noStart bool,
) error {

	t := newPushLogTailer(a, out, appName, resourceVersion, namespace, noStart)
	defer t.ctxCancel()

	for {
		done, err := t.handleWatch()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

func (t *pushLogTailer) handleWatch() (bool, error) {

	ws, err := t.client.kclient.Apps(t.namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: t.resourceVersion,

		FieldSelector: fields.OneTermEqualSelector("metadata.name", t.appName).String(),
		Watch:         true,
	})
	if err != nil {
		return true, err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		switch obj := e.Object.(type) {
		case *v1alpha1.App:
			// skip out of date apps
			if obj.Generation != obj.Status.ObservedGeneration {
				continue
			}

			done, err := t.handleUpdate(obj)
			if err != nil {
				return true, err
			}
			if done {
				return true, nil
			}
		case *k8smeta.Status:
			t.resourceVersion = obj.ResourceVersion
			if obj.Status != k8smeta.StatusFailure {
				// TODO: I'm not sure when/if we would get this.
				continue
			}

			t.logger.Printf("status error: %s:%s\n", obj.Reason, obj.Message)

			return false, nil
		default:
			t.logger.Printf("Unexpected type in watch stream: %T\n", e.Object)
			continue
		}
	}

	return false, nil
}

func (t *pushLogTailer) handleUpdate(
	app *v1alpha1.App,
) (bool, error) {

	sourceReady := app.Status.GetCondition(v1alpha1.AppConditionSourceReady)
	if sourceReady == nil {
		// source might still be creating
		return false, nil
	}
	if sourceReady.Message != "" {
		t.logger.Printf("Updated state to: %s\n", sourceReady.Message)
	}

	switch sourceReady.Status {
	case corev1.ConditionTrue:
		// Only handle source success case once
		t.checkSourceReadyOnce.Do(func() {
			duration := time.Now().Sub(t.buildStartTime)
			t.logger.Printf("Built in %0.2f seconds\n", duration.Seconds())
			t.ctxCancel()
			t.deployStartTime = time.Now()
		})
	case corev1.ConditionFalse:
		t.logger.Printf("Failed to build: %s\n", sourceReady.Message)
		t.ctxCancel()
		return true, fmt.Errorf("build failed: %s", sourceReady.Message)
	default:

		// This case should mean the Source is still in progress.
		// It should be safe to tail the logs to show the user what's happening.
		go t.tailBuildLogsOnce.Do(
			func() {
				// ignoring tail errs because they are spurious
				t.client.sourcesClient.Tail(t.ctx, t.namespace, app.Status.LatestCreatedSourceName, t.out)
			},
		)
		return false, nil
	}

	if t.noStart {
		t.logger.Printf("Total deploy time %0.2f seconds\n", time.Now().Sub(t.deployStartTime).Seconds())
		return true, nil
	}

	appReady := app.Status.GetCondition(v1alpha1.AppConditionReady)
	if appReady == nil {
		return false, nil
	}
	if appReady.Message != "" {
		t.logger.Printf("Updated state to: %s\n", appReady.Message)
	}

	switch appReady.Status {
	case corev1.ConditionTrue:
		now := time.Now()
		duration := now.Sub(t.buildStartTime)
		deployDuration := now.Sub(t.deployStartTime)
		t.logger.Printf("App took %0.2f seconds to become ready.\n", deployDuration.Seconds())
		t.logger.Printf("Total deploy time %0.2f seconds\n", duration.Seconds())
		return true, nil
	case corev1.ConditionFalse:
		t.logger.Printf("Failed to deploy: %s\n", appReady.Message)
		return true, fmt.Errorf("deployment failed: %s", appReady.Message)
	}

	return false, nil
}
