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
	"testing"
	time "time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReverseReader(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx       context.Context
		pod       *corev1.Pod
		container string
		enqueuer  *queue
	}

	testCases := []struct {
		name   string
		setup  func(*args, *fakeWindowReader)
		assert func(*testing.T, error, *fakeWindowReader, *queue)
	}{
		{
			name: "respects context",
			setup: func(a *args, f *fakeWindowReader) {
				// Setup a done context.
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				a.ctx = ctx

				// This error should NOT be returned.
				f.err = errors.New("some-error")
			},
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				// The error should be about the context because it should
				// have never actually invoked the window reader.
				testutil.AssertErrorsEqual(t, errors.New("context canceled"), err)
			},
		},
		{
			name: "doesn't read before Pod was created",
			setup: func(a *args, f *fakeWindowReader) {
				// Setup a creation time in the future, meaning the Pod
				// shouldn't be read from.
				a.pod.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now().Add(time.Hour)}

				// This error should NOT be returned.
				f.err = errors.New("some-error")
			},
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				// The error should be nil because it should have never
				// actually invoked the window reader.
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
		{
			name: "window reader returns an error",
			setup: func(a *args, f *fakeWindowReader) {
				f.err = errors.New("some-error")
			},
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
		{
			name: "retries on empty results",
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				// This will eventually bail because the context we give it
				// expires.
				testutil.AssertErrorsEqual(t, errors.New("context deadline exceeded"), err)
				testutil.AssertTrue(t, "count>1", f.count > 1)

				// It should have walked the time back 5 seconds for each
				// try. We'll have to do an approximation for to assert if
				// this is true as time has obviously passed since the call.
				sinceStart := time.Since(f.start)
				sinceStartDelta := sinceStart - 5*time.Duration(f.count)*time.Second
				testutil.AssertTrue(t, "start is N*5s ago", sinceStartDelta < time.Second)
			},
		},
		{
			name: "time window extends 5 seconds each retry",
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				delta := f.end.Sub(f.start)
				testutil.AssertEqual(t, "delta", 5*time.Second*time.Duration(f.count), delta)
			},
		},
		{
			name: "data is enqueued in reverse",
			setup: func(a *args, f *fakeWindowReader) {
				// Return something so that it tries again with a new time
				// window.
				f.logLines = []logLine{
					{text: "a"},
					{text: "b"},
					{text: "c"},
				}
			},
			assert: func(t *testing.T, err error, f *fakeWindowReader, q *queue) {
				// This will eventually bail because the context we give it
				// expires.
				testutil.AssertErrorsEqual(t, errors.New("context deadline exceeded"), err)

				testutil.AssertEqual(t, "logLine.text", "c", q.Dequeue().text)
				testutil.AssertEqual(t, "logLine.text", "b", q.Dequeue().text)
				testutil.AssertEqual(t, "logLine.text", "a", q.Dequeue().text)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given each one of these test cases might take half a second, we
			// should try to do them all in parallel. This implies we need to
			// shadow tc so we don't get closure issues.
			tc := tc
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			f := &fakeWindowReader{}
			r := reverseLogReader{windowReader: f}
			a := args{
				ctx:       ctx,
				pod:       &corev1.Pod{},
				container: "some-container",
				enqueuer:  newQueue(100),
			}

			if tc.setup != nil {
				tc.setup(&a, f)
			}

			err := r.Read(a.ctx, a.pod, a.container, a.enqueuer)

			if tc.assert != nil {
				tc.assert(t, err, f, a.enqueuer)
			}
		})
	}
}

type fakeWindowReader struct {
	count int

	start     time.Time
	end       time.Time
	pod       *corev1.Pod
	container string

	logLines []logLine
	err      error
}

func (f *fakeWindowReader) Read(
	ctx context.Context,
	start time.Time,
	end time.Time,
	pod *corev1.Pod,
	container string,
) ([]logLine, error) {
	f.count++

	f.start = start
	f.end = end
	f.pod = pod
	f.container = container

	return f.logLines, f.err
}
