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

package kf

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	"github.com/knative/pkg/apis"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Logs handles build and deploy logs.
type Logs interface {
	// DeployLogs writes the logs for the build and deploy stage to the given
	// out.  The method exits once the logs are done streaming.
	DeployLogs(out io.Writer, appName, resourceVersion, namespace string) error
}

// logTailer tails logs for a service. This includes the build and deploy
// step. It should be created via NewLogTailer.
type logTailer struct {
	sc cserving.ServingV1alpha1Interface
}

// NewLogTailer creates a new Logs.
func NewLogTailer(sc cserving.ServingV1alpha1Interface) Logs {
	return &logTailer{
		sc: sc,
	}
}

// DeployLogs writes the logs for the deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (t logTailer) DeployLogs(out io.Writer, appName, resourceVersion, namespace string) error {
	logger := log.New(out, "\033[32m[deploy-revision]\033[0m ", 0)
	logger.Printf("Starting app: %s\n", appName)
	start := time.Now()

	ws, err := t.sc.Services(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
		FieldSelector:   fields.OneTermEqualSelector("metadata.name", appName).String(),
		Watch:           true,
	})
	if err != nil {
		return err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		s := e.Object.(*serving.Service)

		// Don't use status' that are reflecting old states
		if s.Status.ObservedGeneration != s.Generation {
			continue
		}

		if condition := s.Status.GetCondition(apis.ConditionReady); condition != nil {
			switch condition.Status {
			case corev1.ConditionTrue:
				duration := time.Now().Sub(start)
				logger.Printf("Started in %0.2f seconds\n", duration.Seconds())
				return nil
			case corev1.ConditionFalse:
				logger.Printf("Failed to start: %s\n", condition.Message)

				return fmt.Errorf("deployment failed: %s", condition.Message)
			default:
				if condition.Message != "" {
					logger.Printf("Updated state to: %s\n", condition.Message)
				}
			}
		}
	}
	// Lost connection before ready, unknown status.
	return errors.New("lost connection to Kubernetes")
}
