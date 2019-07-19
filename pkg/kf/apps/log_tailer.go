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
	"time"

	"k8s.io/apimachinery/pkg/fields"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeployLogs writes the logs for the deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (a *appsClient) DeployLogs(out io.Writer, appName, resourceVersion, namespace string) error {
	logger := log.New(out, "\033[32m[build]\033[0m ", 0)
	logger.Printf("Starting app: %s\n", appName)
	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var once sync.Once
	var sourceReadyOnce sync.Once
	var deployStart time.Time
	for {
		ws, err := a.kclient.Apps(namespace).Watch(k8smeta.ListOptions{
			ResourceVersion: resourceVersion,
			FieldSelector:   fields.OneTermEqualSelector("metadata.name", appName).String(),
			Watch:           true,
		})
		if err != nil {
			return err
		}
		ws.Stop()

		for e := range ws.ResultChan() {
			app := e.Object.(*v1alpha1.App)

			sourceReady := app.Status.GetCondition(v1alpha1.AppConditionSourceReady)
			if sourceReady == nil {
				continue
			}
			if sourceReady.Message != "" {
				logger.Printf("Updated state to: %s\n", sourceReady.Message)
			}

			switch sourceReady.Status {
			case corev1.ConditionTrue:
				// Only handle source success case once
				sourceReadyOnce.Do(func() {
					duration := time.Now().Sub(start)
					logger.Printf("Built in %0.2f seconds\n", duration.Seconds())
					cancel()
					deployStart = time.Now()
				})
			case corev1.ConditionFalse:
				logger.Printf("Failed to build: %s\n", sourceReady.Message)
				cancel()
				return fmt.Errorf("build failed: %s", sourceReady.Message)
			default:

				// This case should mean the Source is still in progress.
				// It should be safe to tail the logs to show the user what's happening.
				go once.Do(
					func() {
						// ignoring tail errs because they are spurious
						a.sourcesClient.Tail(ctx, namespace, app.Status.LatestCreatedSourceName, out)
					},
				)
				continue
			}

			appReady := app.Status.GetCondition(v1alpha1.AppConditionReady)
			if appReady == nil {
				continue
			}
			if appReady.Message != "" {
				logger.Printf("Updated state to: %s\n", appReady.Message)
			}

			switch appReady.Status {
			case corev1.ConditionTrue:
				now := time.Now()
				duration := now.Sub(start)
				deployDuration := now.Sub(deployStart)
				logger.Printf("App took %0.2f seconds to become ready.\n", deployDuration.Seconds())
				logger.Printf("Total deploy time %0.2f seconds\n", duration.Seconds())
				return nil
			case corev1.ConditionFalse:
				logger.Printf("Failed to deploy: %s\n", appReady.Message)
				return fmt.Errorf("deployment failed: %s", appReady.Message)
			}
		}
		ws.Stop()
	}

	// Lost connection before ready, unknown status.
	return errors.New("lost connection to Kubernetes")
}
