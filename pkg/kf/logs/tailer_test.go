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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	fakeapps "github.com/google/kf/pkg/kf/apps/fake"
	"github.com/google/kf/pkg/kf/logs"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestTailer_Tail(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		appName        string
		opts           []logs.TailOption
		assert         func(t *testing.T, buf *bytes.Buffer, err error)
		eventType      watch.EventType
		pod            *v1.Pod
		expectedOutput string
		watchErr       error
	}{
		"default namespace": {
			appName: "some-app",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
			},
		},
		"custom namespace": {
			appName: "some-app",
			opts: []logs.TailOption{
				logs.WithTailNamespace("custom-namespace"),
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
			},
		},
		"empty app name": {
			appName: "",
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("appName is empty"), err)
			},
		},
		"negative number of lines": {
			appName: "some-app",
			opts: []logs.TailOption{
				logs.WithTailNumberLines(-1),
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("number of lines must be greater than or equal to 0"), err)
			},
		},
		"watching pods fails": {
			appName:  "some-app",
			watchErr: errors.New("some-error"),
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to watch pods: some-error"), err)
			},
		},
		"uses service selector": {
			appName: "some-app",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
			},
		},
		"writes logs to the writer": {
			appName: "some-app",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
			},
		},
		"writes logs about pending pod": {
			appName:   "some-app",
			eventType: watch.Added,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{"Pod 'default/some-app-pod1' is not running\n"})
			},
		},
		"writes logs about deleted pod": {
			appName:   "some-app",
			eventType: watch.Deleted,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-app-pod1",
				},
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{"Pod 'default/some-app-pod1' is deleted\n"})
			},
		},
		"writes logs about terminated pod": {
			appName:   "some-app",
			eventType: watch.Added,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "some-app-pod1",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertNil(t, "err", err)
				testutil.AssertContainsAll(t, buf.String(), []string{"Pod 'default/some-app-pod1' is terminated\n"})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.assert == nil {
				tc.assert = func(t *testing.T, buf *bytes.Buffer, err error) {
					testutil.AssertNil(t, "err", err)
				}
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakeWatcher := fakeapps.NewFakeWatcher(ctrl)
			fakeWatcher.
				EXPECT().
				ResultChan().
				DoAndReturn(func() <-chan watch.Event {
					if len(tc.eventType) != 0 && tc.pod != nil {
						return createUpdatedEvent(watch.Event{
							Type:   tc.eventType,
							Object: tc.pod,
						})
					} else {
						return nil
					}

				}).
				AnyTimes()

			// Ensure Stop is invoked to clean up resources.
			fakeWatcher.
				EXPECT().
				Stop().
				AnyTimes()

			fakeClient := &fake.FakeCoreV1{
				Fake: &ktesting.Fake{},
			}

			fakeClient.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fakeWatcher, tc.watchErr
			}))

			if tc.pod != nil {
				fakeClient.AddReactor("get", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
					return true, tc.pod, nil
				})
			}

			buf := &bytes.Buffer{}
			gotErr := logs.NewTailer(fakeClient).Tail(context.Background(), tc.appName, buf, tc.opts...)

			tc.assert(t, buf, gotErr)
		})
	}
}

func createUpdatedEvent(es watch.Event) <-chan watch.Event {
	c := make(chan watch.Event, 1)
	defer close(c)
	c <- es
	return c
}
