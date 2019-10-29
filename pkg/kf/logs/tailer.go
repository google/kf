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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// Tailer reads the logs for a KF application. It should be created via
// NewTailer().
type Tailer interface {
	// Tail tails the logs from a KF application and writes them to the
	// writer.
	Tail(ctx context.Context, appName string, out io.Writer, opts ...TailOption) error
}

type tailer struct {
	client kubernetes.Interface
}

// NewTailer creates a new Tailer.
func NewTailer(client kubernetes.Interface) Tailer {
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
	logOpts := corev1.PodLogOptions{
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

	writer := &MutexWriter{
		Writer: out,
	}

	if err := t.watchForPods(ctx, namespace, appName, writer, logOpts, cfg.Timeout); err != nil {
		return fmt.Errorf("failed to watch pods: %s", err)
	}
	return nil
}

func (t *tailer) watchForPods(ctx context.Context, namespace, appName string, writer *MutexWriter, opts corev1.PodLogOptions, timeout time.Duration) error {
	w, err := t.client.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: "serving.knative.dev/service=" + appName,
	})
	if err != nil {
		return err
	}

	defer w.Stop()

	// We will only wait a second for the first log. If nothing happens after
	// that period of time and we're not following, then stop.
	initTimer := time.NewTimer(timeout)
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
		case e, ok := <-w.ResultChan():
			if !ok {
				return nil
			}

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				writer.WriteStringf("[WARN] watched object is not pod\n")
				continue
			}

			switch e.Type {
			case watch.Added:
				go func(e watch.Event) {
					t.readLogs(ctx, namespace, pod.Name, writer, opts)
				}(e)
			case watch.Deleted:
				err = writer.WriteStringf("[INFO] Pod '%s/%s' is deleted\n", namespace, pod.Name)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (t *tailer) readLogs(
	ctx context.Context,
	namespace string,
	name string,
	out *MutexWriter,
	opts corev1.PodLogOptions,
) {
	var err error
	var stop bool

	for ctx.Err() == nil && !stop {
		if stop, err = t.readStream(ctx, name, namespace, out, opts); err != nil {
			out.WriteStringf("[WARN] %s", err)
		}

		if !opts.Follow {
			return
		}
		// wait 5 seconds for pod running
		time.Sleep(5 * time.Second)
	}
}

func (t *tailer) readStream(ctx context.Context, name, namespace string, mw *MutexWriter, opts corev1.PodLogOptions) (bool, error) {
	pod, err := t.client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return true, fmt.Errorf("failed to get Pod '%s': %s", name, err)
	}

	if !pod.DeletionTimestamp.IsZero() {
		err = mw.WriteStringf("[INFO] Pod '%s/%s' is terminated\n", namespace, name)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	if pod.Status.Phase != corev1.PodRunning {
		err = mw.WriteStringf("[INFO] Pod '%s/%s' is not running\n", namespace, name)
		if err != nil {
			return false, err
		}

		return false, nil
	}

	err = mw.WriteStringf("[INFO] Pod '%s/%s' is running\n", namespace, pod.Name)
	if err != nil {
		return false, err
	}

	// XXX: This is not tested at a unit level and instead defers to
	// integration tests.
	req := t.client.
		CoreV1().
		Pods(namespace).
		GetLogs(name, &opts).
		Context(ctx)

	stream, err := req.Stream()
	if err != nil {
		return false, fmt.Errorf("failed to read stream: %s", err)
	}
	defer stream.Close()

	err = mw.CopyFrom(stream)
	if err != nil {
		return false, err
	}

	return false, nil
}
