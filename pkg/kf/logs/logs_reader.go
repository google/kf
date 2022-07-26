// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logs

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	time "time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// LogsReader is used to read logs from an Object (e.g., App, Build, Task).
type LogsReader struct {
	client        kubernetes.Interface
	reverseReader reverseReader

	// XXX: This function is written in a way that it can be replaced for
	// testing purposes ONLY. It should NEVER be replaced for production use.
	// There does not yet exist a solid way to test the internal functions
	// outside of an integration test. Therefore the unit tests needed a way,
	// and that way is to replace this function. It defaults to streamLogs.
	streamLogs streamLogsFunc
}

type streamLogsFunc func(
	ctx context.Context,
	client kubernetes.Interface,
	pod *corev1.Pod,
	opts corev1.PodLogOptions,
	f func(logLine) bool,
) error

// Object is an object that can have logs belong to it (or one of its
// children). This is implemented by Kf's CR's (e.g., App).
type Object interface {
	metav1.Object

	// LogSelector returns the selector used to find the Pods that will be
	// fetched or watched for logs. If we are following, then follow will be
	// true.
	LogSelector(follow bool) string
}

// NewLogsReader returns a new LogsReader object.
func NewLogsReader(client kubernetes.Interface) *LogsReader {
	windowReader := newWindowedLogReader(client)
	reverseReader := newReverseLogReader(windowReader)

	return &LogsReader{
		client:        client,
		reverseReader: reverseReader,
		streamLogs:    streamLogs,
	}
}

// Recent returns the given number of log lines for the configured labels.
//
// NOTE: To future developers or even myself... I'd like to explain some of
// the decisions made here. There are a few things that look/feel less than
// optimal. That is because they are. Let me explain:
//
// First, let's explain each abstraction:
// 1. WindowedLogReader - This reads time windows from an individual
// containere on a Pod. Each block is sorted in ascending order before being
// returned.
// 2. ReverseLogReader - This returns the most logs in descending order from a
// container. It writes each log to an enqueuer. It does this by managing a
// sliding window and requesting that time block from the WindowedLogReader.
// 3. LogReader - This is the main object. It fetches all the matching Pods
// and creates a queue for each of their containers. It then creates a
// ReverseLogReader and has it write their logs to the given queue. Each
// ReverseLogReader will fill up their queue (and block). On the main
// go-routine, it will look at each queue to see who has the latest entry. It
// will pick that one and dequeue from it. This process repeats until we have
// enough logs or the client bails.
//
// Second, let's talk about the queue:
// The queue is a simple wrapper around a channel. Read operations are
// non-blocking. If the queue is empty, they will return nil. Read operations
// include Dequeue and Peek. Enqueue however, will block. Therefore it must
// take and adhere to a context so that it can be properly canceled.o
//
// Third, let's talk about what I would have rather done (and why I didn't):
//
// * Instead of queues, I would have preferred priority-queues. I could have
// avoided sorting and potentially reading blocks in general. This fell flat
// though because channels. Our context (the type and concept) uses a channel
// to denote when it is done. Therefore, it becomes difficult to NOT use
// channels to coordinate go-routines. So for example, when I want to write to
// my priority-queue and its full, how do I unblock when it is read from?
// Normally, a conditional mutex would be appropriate. However, my context has
// a channel, so I would need to coordinate between a mutex and a channel. The
// main way I can tell to do this is by starting a go-routine to wait for the
// channel to be closed. However, this also implies that for each write I
// would have to create a go-routine. Ultimately, I became exceedingly nervous
// I was going to leak go-routines.
//
// * I would have liked to avoid reading in blocks of time. In the
// WindowedLogReader, I use time and read until I reach a certain point in
// time. Ideally, I could only read a certain number of lines (to avoid
// reading in too much data into memory). However, given the K8s API, we can
// only start a stream at a certain point in time and start reading forward.
// This makes it tough to reach a set amount of lines.
//
// All this is to say, this can be improved upon, but hopefully you can save
// some time and not go down the same wrong paths I did.
func (l *LogsReader) Recent(outerCtx context.Context, o Object, numLines int) ([]string, error) {
	ctx, cancel := context.WithCancel(outerCtx)

	// Always cancel the context. This context controls the lifecycle of the
	// streamLogs go-routines. If they are still streaming from an App when
	// this function exits, then we leak a go-routine.
	defer cancel()

	// Find all the matching Pods for the given Object.
	pods, err := l.client.
		CoreV1().
		Pods(o.GetNamespace()).
		List(ctx, metav1.ListOptions{
			LabelSelector: o.LogSelector(false),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list Pods: %v", err)
	}

	var (
		funcs []func() error
		qs    []*queue
	)

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if l.filterContainerNames(container.Name) {
				continue
			}

			q := newQueue(100)
			qs = append(qs, q)

			funcs = append(funcs, l.buildStreamWalker(ctx, &pod, container.Name, q))
		}
	}

	lines := make(chan string)
	errs := make(chan error)

	// Start all the LogStreams.
	go func() {
		// There was an error or the context got closed elsewhere. Indicate
		// that this everything should start closing.
		defer cancel()

		// This will run each function and block until they are all done. If
		// any return an error, this will return an error.
		if err := errors.AggregateGoroutines(funcs...); err != nil {
			select {
			case <-ctx.Done():
				return
			case errs <- err:
			}
		}
	}()

	// Consume from all the queues.
	go func() {
		// Indicate to the consuming go-routine that there are no more lines
		// coming.
		defer close(lines)

		// Loop until the context given to us by the user is marked done OR
		// the go-routines that are writing to the queues are all done AND the
		// queues are empty.
		for outerCtx.Err() == nil && (ctx.Err() == nil || !queues(qs).Empty()) {
			var (
				latest  *time.Time
				latestQ *queue
			)

			// Take a peek at each queue and find the one with the latest
			// message. That queue will have its message dequeued.
			for _, q := range qs {
				ll := q.Peek()

				if ll == nil {
					// Queue is empty, move on.
					continue
				}

				if latest == nil || ll.timestamp.After(*latest) {
					latest = &ll.timestamp
					latestQ = q
				}
			}

			// If we don't have a queue, then that implies they were all
			// empty. Wait a moment, and then loop again.
			if latestQ == nil {
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					continue
				case <-outerCtx.Done():
					return
				}
			}

			line := latestQ.Dequeue()
			if line == nil {
				// This should never happen... But just in case we'll check to
				// make sure this isn't nil.
				continue
			}

			// If we have a queue, then use its value.
			select {
			case lines <- line.text:
			case <-outerCtx.Done():
				return
			}
		}
	}()

	// Loop and build the output from all the pod's log lines. If we get an
	// error from any one of the pods, bail with the error. Return the output
	// once the lines channel is closed and we've read everything off of it OR
	// when we've read enough lines.
	//
	// NOTE: We don't need to watch the context as the go-routine writing to
	// the lines channel is watching the context. It will close the lines
	// channel.
	output := []string{}
	for {
		select {
		case err := <-errs:
			return nil, fmt.Errorf("streaming logs failed: %v", err)
		case line, ok := <-lines:
			// Everything has finished and the channel has dried up, bail.
			if !ok {
				return output, nil
			}

			output = append(output, line)

			// Bail if we already have enough lines.
			if len(output) >= numLines {
				return output, nil
			}
		}
	}
}

func (l *LogsReader) buildStreamWalker(
	ctx context.Context,
	pod *corev1.Pod,
	container string,
	q enqueuer,
) func() error {
	return func() error {
		return l.reverseReader.Read(ctx, pod, container, q)
	}
}

// streamContainerLogs reads data from a container and writes it to the lines
// channel. If there is an error, it will write that to the errs channel. This
// function does not block. It will invoke the done function when exiting.
func (l *LogsReader) streamContainerLogs(
	ctx context.Context,
	since time.Time,
	pod *corev1.Pod,
	container string,
	done func(),
	lines chan<- string,
	errs chan<- error,
) {
	go func() {
		defer done()

		opts := corev1.PodLogOptions{
			Container: container,
			Follow:    true,
			SinceTime: &metav1.Time{Time: since},

			// We don't use the timestamp, however the streamLogs function
			// tries to parse it, so go ahead and turn on timestamps.
			Timestamps: true,
		}

		if err := l.streamLogs(ctx, l.client, pod, opts, func(ll logLine) bool {
			select {
			case <-ctx.Done():
			case lines <- ll.text:
			}
			return true
		}); err != nil {
			// Looks like it failed, send the error down the channel.
			select {
			case <-ctx.Done():
			case errs <- err:
			}
		}
	}()
}

// parseLogLine reads the RFC3339 prefix from the log line. If for some
// reason, it isn't there or fails to parse, the current timestamp will be
// used in its place.
func parseLogLine(line string) logLine {
	// Parse the timestamp from log line. Given we enabled "Timestamps" in the
	// LogOptions, log line will be prefixed with the timestamp.
	parts := strings.SplitN(line, " ", 2)

	if len(parts) != 2 {
		// Malformed from Kubernetes. This is highly unlikely.
		return logLine{
			text:      line,
			timestamp: time.Now(),
		}
	}

	t, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		// Malformed from Kubernetes. This is highly unlikely.
		return logLine{
			text:      line,
			timestamp: time.Now(),
		}
	}

	// Success!
	return logLine{
		text:      parts[1],
		timestamp: t,
	}
}

// Follow watches for matching pods, reads from them until they reach a
// completed stage (Succeeded/Failed). The resulting channel is closed when
// the context is marked done.
func (l *LogsReader) Follow(ctx context.Context, o Object, since time.Time) (<-chan string, <-chan error) {
	lines := make(chan string)
	errs := make(chan error)

	go l.follow(ctx, o, since, lines, errs)

	return lines, errs
}

func (l *LogsReader) follow(ctx context.Context, o Object, since time.Time, lines chan<- string, errs chan<- error) {
	ctx, cancel := context.WithCancel(ctx)
	// Always cancel the context. This context controls the lifecycle of the
	// streamLogs go-routines. If they are still streaming from an App when
	// this function exits, then we leak a go-routine.
	defer cancel()

	// This WaitGroup will track when all the children go-routines are
	// done.
	wg := sync.WaitGroup{}

	// Watch for all the matching Pods for the given Object.
	w, err := l.client.
		CoreV1().
		Pods(o.GetNamespace()).
		Watch(ctx, metav1.ListOptions{
			LabelSelector: o.LogSelector(false),
		})
	if err != nil {
		select {
		case <-ctx.Done():
		case errs <- fmt.Errorf("failed to watch Pods: %v", err):
		}
		return
	}

	defer func() {
		// Stop the watch.
		w.Stop()

		// This will unblock once all the children go-routines exit. We have
		// to wait for this so we don't close the channel while something is
		// writing to it.
		//
		// NOTE: This MUST go after the context cancel so that the children
		// go-routines know to shut down.
		wg.Wait()

		// Close the channels so that anything consuming from the channels
		// knows that the producers are done.
		close(lines)
		close(errs)
	}()

	// Watch the event channel for changes in our Pod list. We'll do this on a
	// different go-routine and spin off go-routines for each container that
	// we find. Each new go-routine will listen to their own child context.
	// When that Pod is deleted, then that context is marked done.
	childContexes := make(map[string]func())

	for {
		select {
		case <-ctx.Done():
			return
		case e, closed := <-w.ResultChan():
			if !closed {
				return
			}

			pod, ok := e.Object.(*corev1.Pod)
			if !ok {
				// This is very unlikely to happen... If not impossible.
				log.Printf("watched object is not Pod: %T", e.Object)
				continue
			}

			switch e.Type {
			case watch.Added, watch.Modified:
				// NOTE: We have to watch for both modified Pods and Added
				// Pods because when Pods are first added, they aren't ready
				// to have their logs streamed. This means we sometimes have
				// to wait for them to become ready.
				l.followPod(ctx, since, &wg, pod, childContexes, lines, errs)
			case watch.Deleted:
				// Close the corresponding child context (if one is being
				// tracked).
				if cancel := childContexes[pod.Name]; cancel != nil {
					cancel()
					delete(childContexes, pod.Name)
				}
			}
		}
	}
}

func (l *LogsReader) followPod(
	ctx context.Context,
	since time.Time,
	wg *sync.WaitGroup,
	pod *corev1.Pod,
	childContexes map[string]func(),
	lines chan<- string,
	errs chan<- error,
) {
	// Ensure the Pod phase is something we can read logs from.
	if pod.Status.Phase == corev1.PodUnknown || pod.Status.Phase == corev1.PodPending {
		// The Pod is not in a phase where we can yet read logs.
		return
	}

	// Check to see if we're somehow already tracking this
	// Pod. If we are move on. Otherwise, create a child
	// context for it and start reading its logs.
	if _, ok := childContexes[pod.Name]; ok {
		// Already tracking this Pod, move on.
		return
	}

	// Create the child context and track it.
	childCtx, cancel := context.WithCancel(ctx)
	childContexes[pod.Name] = cancel

	for _, container := range pod.Spec.Containers {
		if l.filterContainerNames(container.Name) {
			continue
		}

		// Follow each container in the pod.
		wg.Add(1)
		l.streamContainerLogs(
			childCtx,
			since,
			pod,
			container.Name,
			wg.Done,
			lines,
			errs,
		)
	}
}

func (l *LogsReader) filterContainerNames(name string) bool {
	// TODO: Ideally, we would filter out the injected containers from the
	// Pod. However, there isn't an obvious way to detect that yet.
	return name == "istio-proxy" || name == "istio-init"
}

type logLine struct {
	timestamp time.Time
	text      string
}

type logLines []logLine

// Len is the number of elements in the collection.
func (ll logLines) Len() int {
	return len(ll)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ll logLines) Less(i int, j int) bool {
	return ll[i].timestamp.UnixNano() < ll[j].timestamp.UnixNano()
}

// Swap swaps the elements with indexes i and j.
func (ll logLines) Swap(i int, j int) {
	tmp := ll[i]
	ll[i] = ll[j]
	ll[j] = tmp
}

// streamLogs makes a GetLogs request to the given pod. The function blocks
// while it consumes the stream.
//
// NOTE: This should NOT be invoked directly. Instead use
// LogsReader.streamLogs.
func streamLogs(
	ctx context.Context,
	client kubernetes.Interface,
	pod *corev1.Pod,
	opts corev1.PodLogOptions,
	f func(logLine) bool,
) error {
	// Build the request for the given Pod.
	req := client.
		CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &opts)

	// Start reading from the stream.
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to read stream: %s", err)
	}
	defer stream.Close()

	// Start reading each line of the stream. Write each line to the given
	// channel and any errors to the errs channel.
	//
	// NOTE: This doesn't need to read from the context because the stream
	// takes care of that. This means when the context is canceled, the
	// stream is closed and the function will exit organically.
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		if !f(parseLogLine(scanner.Text())) {
			return nil
		}
	}

	return scanner.Err()
}
