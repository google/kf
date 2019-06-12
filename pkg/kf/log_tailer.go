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
	"fmt"
	"io"

	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
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
	ws, err := t.sc.Services(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
		Watch:           true,
	})
	if err != nil {
		return err
	}
	defer ws.Stop()

	for e := range ws.ResultChan() {
		s := e.Object.(*serving.Service)
		if s.Name != appName {
			continue
		}

		for _, condition := range s.Status.Conditions {
			if condition.Reason == "RevisionFailed" {
				return fmt.Errorf("deployment failed: %s", condition.Message)
			}

			if condition.Message != "" {
				fmt.Fprintf(out, "\033[32m[deploy-revision]\033[0m %s\n", condition.Message)
			}
		}
	}
	// Success
	return nil
}
