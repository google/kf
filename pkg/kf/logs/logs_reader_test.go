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
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestLogsReader_Recent(t *testing.T) {
	t.Parallel()

	type fakes struct {
		client        *fake.Clientset
		reverseReader *fakeReverseReader
	}

	testCases := []struct {
		name   string
		lines  int
		obj    Object
		setup  func(t *testing.T, f *fakes)
		assert func(t *testing.T, output []string, err error)
	}{
		{
			name: "listing pods fails",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				f.client.Fake.PrependReactor("list", "pods",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some-error")
					},
				)
			},
			assert: func(t *testing.T, output []string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to list Pods: some-error"), err)
			},
		},
		{
			name: "reverseReader returns an error",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			lines: 30,
			setup: func(t *testing.T, f *fakes) {
				f.reverseReader.err = errors.New("some-error")
				f.client.Fake.PrependReactor("list", "pods",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, podList("some-ns", "some-name", 2), nil
					},
				)
			},
			assert: func(t *testing.T, output []string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("streaming logs failed: some-error"), err)
			},
		},
		{
			name: "fetches the latest data from the reverseReader queues",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			lines: 4,
			setup: func(t *testing.T, f *fakes) {
				f.reverseReader.m = map[string][]logLine{
					"a": {
						{timestamp: time.Unix(100, 0), text: "a"},
						{timestamp: time.Unix(98, 0), text: "c"},
					},
					"b": {
						{timestamp: time.Unix(99, 0), text: "b"},
						{timestamp: time.Unix(97, 0), text: "d"},
					},
				}
				f.client.Fake.PrependReactor("list", "pods",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, podList("some-ns", "some-name", 1), nil
					},
				)
			},
			assert: func(t *testing.T, output []string, err error) {
				testutil.AssertErrorsEqual(t, nil, err)
				testutil.AssertEqual(t, "output", []string{"a", "b", "c", "d"}, output)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakes{
				client:        fake.NewSimpleClientset(),
				reverseReader: &fakeReverseReader{},
			}

			if tc.setup != nil {
				tc.setup(t, f)
			}

			logsReader := NewLogsReader(f.client)
			logsReader.reverseReader = f.reverseReader

			output, err := logsReader.Recent(context.Background(), tc.obj, tc.lines)

			if tc.assert != nil {
				tc.assert(t, output, err)
			}
		})
	}
}

func TestLogsReader_Follow(t *testing.T) {
	t.Parallel()

	now := time.Now()

	type fakes struct {
		client        *fake.Clientset
		reverseReader *fakeReverseReader

		// streamLogs can be replaced in setup function. If it is not, then a
		// NOP function will be used.
		streamLogs streamLogsFunc
		ctx        context.Context
	}

	testCases := []struct {
		name   string
		lines  int
		obj    Object
		setup  func(t *testing.T, f *fakes)
		assert func(t *testing.T, output <-chan string, errs <-chan error)
	}{
		{
			name: "watching pod fails",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				f.client.Fake.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(nil, errors.New("some-error")))
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Ensure we eventually get an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					t.Fatal("expected an error")
				case err := <-errs:
					testutil.AssertErrorsEqual(t, errors.New("failed to watch Pods: some-error"), err)
				}
			},
		},
		{
			name: "streamLogs fails",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					return errors.New("some-error")
				}

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watcherWithPods("some-ns", "some-name", 1)
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Ensure we eventually get an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					t.Fatal("expected an error")
				case err := <-errs:
					testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
				}
			},
		},
		{
			name: "reads logs from every non-istio container",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {

				var (
					mu         sync.Mutex
					containers []string
				)

				t.Cleanup(func() {
					mu.Lock()
					defer mu.Unlock()

					// Ensure we got all the containers we wanted.
					// Sort them to make it easier to assert against.
					sort.Strings(containers)
					testutil.AssertEqual(t, "containers", []string{"a", "b"}, containers)
				})

				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					mu.Lock()
					containers = append(containers, opts.Container)
					mu.Unlock()

					// Ensure we used the passed time.
					testutil.AssertEqual(t, "SinceTime", now, opts.SinceTime.Time)

					// Write some log lines.
					for i := 0; i < 2; i++ {
						f(logLine{text: fmt.Sprintf("%s %d", opts.Container, i)})
					}

					return nil
				}

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watcherWithPods("some-ns", "some-name", 1)
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				assertOnLines(t, []string{"a 0", "a 1", "b 0", "b 1"}, output, errs)
			},
		},
		{
			name: "listens for modified Pods",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					// Write some log lines.
					for i := 0; i < 2; i++ {
						f(logLine{text: fmt.Sprintf("%s %d", opts.Container, i)})
					}
					return nil
				}

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watch.NewFake()
					go func() {
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							pod.Status.Phase = corev1.PodSucceeded
							watcher.Modify(&pod)
						}
					}()
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				assertOnLines(t, []string{"a 0", "a 1", "b 0", "b 1"}, output, errs)
			},
		},
		{
			name: "ignores non-ready Pods",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					t.Fatal("should not be invoked")
					return nil
				}

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watch.NewFake()
					go func() {
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							pod.Status.Phase = corev1.PodUnknown
							watcher.Add(&pod)
						}
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							pod.Status.Phase = corev1.PodPending
							watcher.Add(&pod)
						}
					}()
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Watch for an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					return
				case err := <-errs:
					t.Fatal(err)
				}
			},
		},
		{
			name: "can re-add pod after it has been deleted",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				var count int64
				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					atomic.AddInt64(&count, 1)
					return nil
				}

				t.Cleanup(func() {
					testutil.AssertEqual(t, "count", int64(4), atomic.LoadInt64(&count))
				})

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watch.NewFake()
					go func() {
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							watcher.Add(&pod)
						}
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							watcher.Delete(&pod)
						}
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							watcher.Add(&pod)
						}
					}()
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Watch for an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					return
				case err := <-errs:
					t.Fatal(err)
				}
			},
		},
		{
			name: "stops reading from Pod that has been deleted",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {

				// The way we're going to orchestrate this test is as follows:
				// 1. Send an add event for a pod. The delete event will be
				// sent AFTER the streamLogs function is invoked.
				// 2. StreamLogs will wait (for a little while) for the given
				// context to be canceled.

				var wg sync.WaitGroup
				wg.Add(2)
				var count int64

				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					atomic.AddInt64(&count, 1)
					wg.Done()

					// Eventually, this context should be closed.
					timer := time.NewTimer(250 * time.Millisecond)
					select {
					case <-timer.C:
						t.Fatalf("expected context to be canceled")
					case <-ctx.Done():
						// Success!
					}

					return nil
				}

				t.Cleanup(func() {
					if atomic.LoadInt64(&count) == 0 {
						t.Fatal("expected streamLogs to be invoked")
					}
				})

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watch.NewFake()
					go func() {
						for _, pod := range podList("some-ns", "some-name", 1).Items {
							watcher.Add(&pod)
						}

						// Wait for the streamLogs function to be invoked.
						wg.Wait()

						for _, pod := range podList("some-ns", "some-name", 1).Items {
							watcher.Delete(&pod)
						}
					}()
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Watch for an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					return
				case err := <-errs:
					t.Fatal(err)
				}
			},
		},
		{
			name: "output channels are closed when context is canceled",
			obj: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-ns",
					Name:      "some-name",
				},
			},
			setup: func(t *testing.T, f *fakes) {
				// Setup a canceled context.
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				f.ctx = ctx

				f.streamLogs = func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					return nil
				}

				f.client.Fake.PrependWatchReactor("pods", func(action ktesting.Action) (bool, watch.Interface, error) {
					watcher := watcherWithPods("some-ns", "some-name", 0)
					return true, watcher, nil
				})
			},
			assert: func(t *testing.T, output <-chan string, errs <-chan error) {
				// Watch for an error.
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-timer.C:
					return
				case _, notClosed := <-output:
					testutil.AssertFalse(t, "not closed", notClosed)
				case _, notClosed := <-errs:
					testutil.AssertFalse(t, "not closed", notClosed)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakes{
				client:        fake.NewSimpleClientset(),
				reverseReader: &fakeReverseReader{},
				ctx:           context.Background(),

				// Set a default streamLogs function.
				streamLogs: func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					// NOP
					return nil
				},
			}

			if tc.setup != nil {
				tc.setup(t, f)
			}

			logsReader := NewLogsReader(f.client)
			logsReader.streamLogs = f.streamLogs
			logsReader.reverseReader = f.reverseReader

			output, errs := logsReader.Follow(f.ctx, tc.obj, now)
			if tc.assert != nil {
				tc.assert(t, output, errs)
			}
		})
	}
}

func podList(ns, name string, n int) *corev1.PodList {
	list := &corev1.PodList{}
	for i := 0; i < n; i++ {
		list.Items = append(list.Items, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", i),
				Namespace: ns,
				Labels: map[string]string{
					"app.kubernetes.io/component":  "app-server",
					"app.kubernetes.io/managed-by": "kf",
					"app.kubernetes.io/name":       name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "istio-proxy",
					},
					{
						Name: "istio-init",
					},
				},
			},
		})
	}

	return list
}

func assertOnLines(t *testing.T, expected []string, output <-chan string, errs <-chan error) {
	t.Helper()

	// Read from the output channels. If there is an error, fail
	// the test. If there is a log line, then keep track of it.
	// Once the timer lapses, we will assert against what we
	// have.
	lines := []string{}
	timer := time.NewTimer(250 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			// Times up, lets see what we have.
			// Sort the output to make it easier to assert
			// against.
			sort.Strings(lines)
			sort.Strings(expected)

			testutil.AssertEqual(t, "lines", expected, lines)
			return
		case line := <-output:
			lines = append(lines, line)
		case err := <-errs:
			t.Fatal(err)
		}
	}
}

func watcherWithPods(ns, name string, n int) *watch.FakeWatcher {
	watcher := watch.NewFake()
	go func() {
		for _, pod := range podList("some-ns", "some-name", n).Items {
			pod.Status.Phase = corev1.PodRunning
			watcher.Add(&pod)
		}
	}()
	return watcher
}

func TestStreamLogs(t *testing.T) {
	t.Parallel()

	// XXX: Stream logs is only tested via integration tests. The reason is,
	// the client-go function to fetch logs (`GetLogs()`) is simple... It just
	// creates a rest.Request and then you execute it. Kubernetes fakes
	// weren't setup to allow for this. Therefore, setting up fakes to make
	// this unit-testable would be manual and hand-spun. Therefore, it makes
	// more sense to allow for integration tests to cover it.
}

type fakeReverseReader struct {
	err error
	m   map[string][]logLine
}

func (f *fakeReverseReader) Read(
	ctx context.Context,
	pod *corev1.Pod,
	container string,
	q enqueuer,
) error {
	// XXX: Give each stream's go-routine a moment to be scheduled. This
	// ensures we don't have a race condition. Its likely we can coordinate
	// this better.
	time.Sleep(100 * time.Millisecond)

	for _, ll := range f.m[container] {
		q.Enqueue(ctx, ll)
	}

	return f.err
}
