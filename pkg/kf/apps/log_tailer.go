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
	"time"

	"k8s.io/apimachinery/pkg/fields"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeployLogs writes the logs for the deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (t *appsClient) DeployLogs(out io.Writer, appName, resourceVersion, namespace string) error {
	logger := log.New(out, "\033[32m[build]\033[0m ", 0)
	logger.Printf("Starting app: %s\n", appName)
	start := time.Now()

	ws, err := t.kclient.Apps(namespace).Watch(k8smeta.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", appName).String(),
		Watch:         true,
	})
	if err != nil {
		return err
	}
	defer ws.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	for e := range ws.ResultChan() {

		s := e.Object.(*v1alpha1.App)

		//var once sync.Once
		if condition := s.Status.GetCondition(v1alpha1.AppConditionSourceReady); condition != nil {
			switch condition.Status {
			case corev1.ConditionTrue:
				duration := time.Now().Sub(start)
				logger.Printf("Built in %0.2f seconds\n", duration.Seconds())
				cancel()
				return nil
			case corev1.ConditionFalse:
				logger.Printf("Failed to build: %s\n", condition.Message)
				cancel()
				return fmt.Errorf("build failed: %s", condition.Message)
			default:
				go func() {
					if err := t.sourcesClient.Tail(ctx, namespace, s.Status.LatestCreatedSourceName, out); err != nil {
						fmt.Println(err)
					}
				}()
				if condition.Message != "" {
					logger.Printf("Updated state to: %s\n", condition.Message)
				}
			}
		}
	}
	// Lost connection before ready, unknown status.
	return errors.New("lost connection to Kubernetes")
}
