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
	"context"
	"fmt"
	"io"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// BuildTailer writes the build logs to out.
type BuildTailer interface {
	Tail(ctx context.Context, out io.Writer, buildName, namespace string) error
}

// BuildTailerFunc converts a func into a BuildTailer.
type BuildTailerFunc func(ctx context.Context, out io.Writer, buildName, namespace string) error

// Tail implements BuildTailer.
func (f BuildTailerFunc) Tail(ctx context.Context, out io.Writer, buildName, namespace string) error {
	return f(ctx, out, buildName, namespace)
}

// logTailer tails logs for a service. This includes the build and deploy
// step. It should be created via NewLogTailer.
type logTailer struct {
	bc cbuild.BuildV1alpha1Interface
	sc cserving.ServingV1alpha1Interface
	t  BuildTailer
}

// NewLogTailer creates a new Logs.
func NewLogTailer(
	bc cbuild.BuildV1alpha1Interface,
	sc cserving.ServingV1alpha1Interface,
	t BuildTailer,
) Logs {
	return &logTailer{
		bc: bc,
		sc: sc,
		t:  t,
	}
}

// Tail writes the logs for the build and deploy step for the resourceVersion
// to out. It blocks until the operation has completed.
func (t logTailer) Tail(out io.Writer, appName, resourceVersion, namespace string) error {
	errs := make(chan error, 2)

	ctx, cancel := context.WithCancel(context.Background())

	go func() { errs <- t.buildLogs(ctx, out, appName, resourceVersion, namespace) }()
	go func() { errs <- t.serviceLogs(cancel, out, appName, resourceVersion, namespace) }()

	err1, err2 := <-errs, <-errs
	for _, err := range []error{err1, err2} {
		if err != nil {
			return err
		}
	}

	return nil
}

func (t logTailer) buildLogs(ctx context.Context, out io.Writer, appName, resourceVersion, namespace string) error {
	wb, err := t.bc.Builds(namespace).Watch(k8smeta.ListOptions{
		ResourceVersion: resourceVersion,
		Watch:           true,
	})
	if err != nil {
		return err
	}
	defer wb.Stop()

	go func() {
		<-ctx.Done()
		wb.Stop()
	}()

	for e := range wb.ResultChan() {
		obj := e.Object.(*build.Build)
		var foundRef bool
		for _, ref := range obj.ObjectMeta.OwnerReferences {
			if ref.Name == appName {
				foundRef = true
				break
			}
		}
		if !foundRef {
			continue
		}

		if e.Type == watch.Added {
			// This blocks until the build is finished.
			if err := t.t.Tail(ctx, out, obj.Name, namespace); err != nil {
				return err
			}
		}

		for _, condition := range obj.Status.Conditions {
			if condition.Type == "Succeeded" && condition.Status == "False" {
				return fmt.Errorf("build failed: %s", condition.Message)
			}
		}
	}

	return nil
}

func (t logTailer) serviceLogs(done func(), out io.Writer, appName, resourceVersion, namespace string) error {
	defer done()
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
