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

package logs_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	time "time"

	"github.com/google/kf/pkg/kf/logs"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestTailer_Tail_invalid_input(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName string
		opts    []logs.TailOption
		wantErr error
	}{
		"empty app name": {
			appName: "",
			wantErr: errors.New("appName is empty"),
		},
		"negative number of lines": {
			appName: "some-app",
			opts: []logs.TailOption{
				logs.WithTailNumberLines(-1),
			},
			wantErr: errors.New("number of lines must be greater than or equal to 0"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			gotErr := logs.NewTailer(nil).Tail(context.Background(), tc.appName, nil, tc.opts...)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
		})
	}
}

const defaultAppName = "some-app"

func TestTailer_Tail(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		opts   []logs.TailOption
		setup  func(t *testing.T, cs *fake.Clientset) context.Context
		assert func(t *testing.T, buf *bytes.Buffer, err error)
	}{
		"default namespace": {
			opts: []logs.TailOption{
				logs.WithTailTimeout(0),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				cs.PrependWatchReactor("pods", namespaceWatchReactor(t, "default"))
				return context.Background()
			},
		},
		"custom namespace": {
			opts: []logs.TailOption{
				logs.WithTailTimeout(0),
				logs.WithTailNamespace("custom-namespace"),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				cs.PrependWatchReactor("pods", namespaceWatchReactor(t, "custom-namespace"))
				return context.Background()
			},
		},
		"watching pods fails": {
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				cs.PrependWatchReactor("pods", errorWatchReactor(t, errors.New("some-error")))
				return context.Background()
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to watch pods: some-error"), err)
			},
		},
		"non-pod event": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				watcher := watch.NewFake()
				cs.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))
				go watcher.Add(&metav1.Status{})
				return context.Background()
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{
					"[WARN] watched object is not pod\n",
				})
			},
		},
		"cancelled context": {
			opts: []logs.TailOption{
				logs.WithTailTimeout(time.Hour),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				watcher := watch.NewFake()
				cs.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"stopped watcher": {
			opts: []logs.TailOption{
				logs.WithTailTimeout(time.Hour),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				watcher := watch.NewFake()
				cs.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))
				watcher.Stop()
				return context.Background()
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		"uses label selector": {
			opts: []logs.TailOption{
				logs.WithTailTimeout(0),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				cs.PrependWatchReactor("pods", labelSelectorWatchReactor(t, "serving.knative.dev/service="+defaultAppName))
				return context.Background()
			},
		},
		"writes logs about deleted pod": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			setup: func(t *testing.T, cs *fake.Clientset) context.Context {
				watcher := watch.NewFake()
				cs.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))
				go watcher.Delete(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultAppName,
					},
				})
				return context.Background()
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{
					fmt.Sprintf("Pod 'default/%s' is deleted\n", defaultAppName),
				})
			},
		},
		"getting pod fails": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			// We're going to setup a watcher, but the pod doesn't exist
			// so Pods(namespace).Get() will return an error.
			setup: whenAddEvent(nil),
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{
					fmt.Sprintf(`[WARN] failed to get Pod '%s': pods "%s" not found`, defaultAppName, defaultAppName),
				})
			},
		},
		"pod is deleted": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			setup: whenAddEvent(
				whenPodAdded(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultAppName,
						DeletionTimestamp: &metav1.Time{
							Time: time.Now(),
						},
					},
				}, nil),
			),
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{
					fmt.Sprintf("[INFO] Pod 'default/%s' is terminated\n", defaultAppName),
				})
			},
		},
		"pod is not running": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			setup: whenAddEvent(
				whenPodAdded(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultAppName,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				}, nil),
			),
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{
					fmt.Sprintf("[INFO] Pod 'default/%s' is not running\n", defaultAppName),
				})
			},
		},
		"write logs from pod": {
			opts: []logs.TailOption{
				// This helps the test move a little faster.
				logs.WithTailTimeout(250 * time.Millisecond),
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				t.Skip("https://github.com/kubernetes/kubernetes/issues/84203")
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.assert == nil {
				tc.assert = func(t *testing.T, buf *bytes.Buffer, err error) {
					testutil.AssertNil(t, "err", err)
				}
			}

			if tc.setup == nil {
				tc.setup = func(*testing.T, *fake.Clientset) context.Context {
					return context.Background()
				}
			}

			fakeClient := fake.NewSimpleClientset()
			ctx := tc.setup(t, fakeClient)

			// We need to use a MutexWriter because the Tailer writes to the
			// writer on different go routines than what we are reading from.
			// We need to ensure that we can read safely while not upsetting
			// the race detector.
			mw := &logs.MutexWriter{Writer: &bytes.Buffer{}}
			gotErr := logs.NewTailer(fakeClient).Tail(ctx, defaultAppName, mw, tc.opts...)

			buf := &bytes.Buffer{}
			mw.Lock()
			buf.Write(mw.Writer.(*bytes.Buffer).Bytes())
			mw.Unlock()
			tc.assert(t, buf, gotErr)
		})
	}
}

func namespaceWatchReactor(t *testing.T, namespace string) ktesting.WatchReactionFunc {
	t.Helper()
	return func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		if namespace != action.GetNamespace() {
			t.Errorf("%s: expected namespace %q, got %q", t.Name(), namespace, action.GetNamespace())
		}

		return false, nil, nil
	}
}

func labelSelectorWatchReactor(t *testing.T, selector string) ktesting.WatchReactionFunc {
	t.Helper()
	return func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		actualLabel := action.(ktesting.WatchActionImpl).WatchRestrictions.Labels.String()

		if actualLabel != selector {
			t.Errorf("%s: expected label selector %q, got %q", t.Name(), selector, actualLabel)
		}

		return false, nil, nil
	}
}

func errorWatchReactor(t *testing.T, err error) ktesting.WatchReactionFunc {
	t.Helper()
	return func(action ktesting.Action) (bool, watch.Interface, error) {
		t.Helper()
		return true, nil, err
	}
}

func whenAddEvent(f func(*testing.T, *fake.Clientset)) func(*testing.T, *fake.Clientset) context.Context {
	return func(t *testing.T, cs *fake.Clientset) context.Context {
		watcher := watch.NewFake()
		cs.PrependWatchReactor("pods", ktesting.DefaultWatchReactor(watcher, nil))
		go watcher.Add(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultAppName,
			},
		})

		if f != nil {
			f(t, cs)
		}

		return context.Background()
	}
}

func whenPodAdded(pod *corev1.Pod, f func(*testing.T, *fake.Clientset)) func(*testing.T, *fake.Clientset) {
	return func(t *testing.T, cs *fake.Clientset) {
		cs.CoreV1().Pods("default").Create(pod)
		if f != nil {
			f(t, cs)
		}
	}
}
