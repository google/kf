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

package logs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Tailer reads the logs for a KF application. It should be created via
// NewTailer().
type Tailer interface {
	// Tail tails the logs from a KF application and writes them to the
	// writer.
	Tail(ctx context.Context, appName string, out io.Writer, opts ...TailOption) error
}

type tailer struct {
	client corev1.CoreV1Interface
}

// NewTailer creates a new Tailer.
func NewTailer(client corev1.CoreV1Interface) Tailer {
	return &tailer{
		client: client,
	}
}

// Tail tails the logs from a KF application and writes them to the writer.
func (t *tailer) Tail(ctx context.Context, appName string, out io.Writer, opts ...TailOption) error {
	cfg := TailOptionDefaults().Extend(opts).toConfig()
	if appName == "" {
		return errors.New("appName is empty")
	}

	if cfg.NumberLines < 0 {
		return errors.New("number of lines must be greater than or equal to 0")
	}

	namespace := cfg.Namespace
	logOpts := v1.PodLogOptions{
		// We need to specify which container we want to use. 'user-container'
		// is the container where the user's application is ran (as opposed to
		// a side-car such as istio-proxy).
		Container: "user-container",
		Follow:    cfg.Follow,
	}

	if cfg.NumberLines != 0 {
		n := int64(cfg.NumberLines)
		logOpts.TailLines = &(n)
	}

	if err := t.watchForPods(ctx, namespace, appName, out, logOpts); err != nil {
		return fmt.Errorf("failed to watch pods: %s", err)
	}
	return nil
}

func (t *tailer) watchForPods(ctx context.Context, namespace, appName string, out io.Writer, opts v1.PodLogOptions) error {
	w, err := t.client.Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: "serving.knative.dev/service=" + appName,
	})
	if err != nil {
		return err
	}

	defer w.Stop()

	// We will only wait a second for the first log. If nothing happens after
	// that period of time and we're not following, then stop.
	initTimer := time.NewTimer(time.Second)
	if opts.Follow {
		initTimer.Stop()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-initTimer.C:
			// Time out waiting for a log.
			return nil
		case e := <-w.ResultChan():
			switch e.Type {
			case watch.Added:
				go func(e watch.Event) {
					t.readLogs(ctx, e.Object.(*v1.Pod).Name, namespace, out, opts)
					cancel()
				}(e)
			case watch.Deleted:
				if _, err := io.WriteString(out, fmt.Sprintf("Pod '%s' is deleted\n", e.Object.(*v1.Pod).Name)); err != nil {
					return err
				}
			}
		}
	}
}

func (t *tailer) readLogs(ctx context.Context, name, namespace string, out io.Writer, opts v1.PodLogOptions) {
	for ctx.Err() == nil {
		if err := t.readStream(ctx, name, namespace, out, opts); err != nil {
			log.Printf("[WARN] %s", err)
		}

		if !opts.Follow {
			return
		}
	}
}

func (t *tailer) readStream(ctx context.Context, name, namespace string, out io.Writer, opts v1.PodLogOptions) error {
	pod, err := t.client.Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Pod '%s': %s", name, err)
	}

	if pod.DeletionTimestamp != nil {
		if _, err := io.WriteString(out, fmt.Sprintf("Pod '%s' is terminated\n", name)); err != nil {
			return err
		}

		return nil
	}

	// XXX: This is not tested at a unit level and instead defers to
	// integration tests.
	req := t.client.
		Pods(namespace).
		GetLogs(name, &opts).
		Context(ctx)

	stream, err := req.Stream()
	if err != nil {
		return fmt.Errorf("failed to read stream: %s", err)
	}
	defer stream.Close()

	if _, err := io.Copy(out, stream); err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}

	return nil
}
