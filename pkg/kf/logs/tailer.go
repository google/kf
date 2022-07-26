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

	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

var (
	warningColor   = color.New(color.FgYellow)
	podChangeColor = color.New(color.FgWhite, color.Faint)
)

// Tailer reads the logs for a Kf application. It should be created via
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

// Tail tails the logs from a Kf application and writes them to the writer.
func (t *tailer) Tail(ctx context.Context, appName string, out io.Writer, opts ...TailOption) error {
	cfg := TailOptionDefaults().Extend(opts).toConfig()
	if appName == "" {
		return errors.New("appName is empty")
	}

	if cfg.NumberLines < 0 {
		return errors.New("number of lines must be greater than or equal to 0")
	}

	namespace := cfg.Space

	logOpts := corev1.PodLogOptions{
		// We need to specify which container we want to use. 'user-container'
		// is the container where the user's application is ran (as opposed to
		// a side-car such as istio-proxy).
		Container:  cfg.ContainerName,
		Follow:     cfg.Follow,
		Timestamps: true,
	}

	if cfg.NumberLines != 0 {
		n := int64(cfg.NumberLines)
		logOpts.TailLines = &(n)
	}

	writer := &LineWriter{
		Writer:        out,
		NumberOfLines: cfg.NumberLines,
	}

	if err := t.watchForPods(ctx, namespace, appName, writer, logOpts, cfg); err != nil {
		return fmt.Errorf("failed to watch pods: %s", err)
	}
	return nil
}

func (t *tailer) watchForPods(ctx context.Context, namespace, appName string, writer *LineWriter, opts corev1.PodLogOptions, cfg tailConfig) error {

	appLabels := v1alpha1.AppComponentLabels(appName, cfg.ComponentName)

	selectors := v1alpha1.UnionMaps(appLabels, cfg.Labels)

	w, err := t.client.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(selectors)),
	})
	if err != nil {
		return err
	}

	defer w.Stop()

	// We will only wait a second for the first log. If nothing happens after
	// that period of time and we're not following, then stop.
	initTimer := time.NewTimer(cfg.Timeout)
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
				log.Println("[WARN] watched object is not pod")
				continue
			}

			switch e.Type {
			case watch.Added:
				go func(e watch.Event) {
					t.readLogs(ctx, namespace, pod.Name, writer, opts)
				}(e)
			case watch.Deleted:
				log.Println(podChangeColor.Sprintf("[INFO] Pod '%s/%s' has been deleted", namespace, pod.Name))
			}
		}
	}
}

func (t *tailer) readLogs(
	ctx context.Context,
	namespace string,
	name string,
	out *LineWriter,
	opts corev1.PodLogOptions,
) {
	var err error
	var stop bool
	var ts time.Time

	for ctx.Err() == nil && !stop {
		if stop, ts, err = t.readStream(ctx, name, namespace, out, opts); err != nil {
			log.Println(warningColor.Sprintf("[WARN] %s", err))
		}

		if !opts.Follow {
			return
		}
		// wait 5 seconds for pod running
		time.Sleep(5 * time.Second)

		if !ts.IsZero() {
			opts.SinceTime = &metav1.Time{Time: ts}
		}
	}
}

func (t *tailer) readStream(ctx context.Context, name, namespace string, mw *LineWriter, opts corev1.PodLogOptions) (bool, time.Time, error) {
	ts := time.Now()
	if opts.SinceTime != nil {
		ts = opts.SinceTime.Time
	}

	pod, err := t.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return true, ts, fmt.Errorf("failed to get Pod '%s': %s", name, err)
	}

	if !pod.DeletionTimestamp.IsZero() {
		log.Println(podChangeColor.Sprintf("[INFO] Pod '%s/%s' is terminated", namespace, name))
		return true, ts, nil
	}

	// If we aren't following, then we want more historical logs. This implies
	// if the pod has succeeded/failed, then let it go through this (but only
	// once).
	if pod.Status.Phase != corev1.PodRunning {
		log.Println(podChangeColor.Sprintf("[INFO] Pod '%s/%s' is not running (phase=%q)", namespace, name, pod.Status.Phase))
		switch pod.Status.Phase {
		case corev1.PodSucceeded, corev1.PodFailed:
			if opts.Follow {
				return true, ts, nil
			}
			// Otherwise, read the logs.
		case corev1.PodPending, corev1.PodUnknown:
			// Can't read from a pod in this phase. Try again later.
			return false, ts, nil
		}
	}

	log.Println(podChangeColor.Sprintf("[INFO] Pod '%s/%s' is running", namespace, pod.Name))

	// XXX: This is not tested at a unit level and instead defers to
	// integration tests.
	req := t.client.
		CoreV1().
		Pods(namespace).
		GetLogs(name, &opts)

	stream, err := req.Stream(ctx)
	if err != nil {
		return false, ts, fmt.Errorf("failed to read stream: %s", err)
	}
	defer stream.Close()

	lastTime, err := mw.CopyFrom(stream)
	if err != nil {
		return false, ts, err
	}

	if !lastTime.IsZero() {
		// Advance it by a second so if we repeat on this pod, we don't pick up
		// any logs that we already have. We use a second because RFC3339
		// doesn't do partial seconds and SinceTime is formatted in RFC3339.
		ts = lastTime.Add(time.Second)
	}

	// If the pod is done, then don't read it again.
	if pod.Status.Phase != corev1.PodRunning {
		return true, ts, nil
	}

	return false, ts, nil
}
