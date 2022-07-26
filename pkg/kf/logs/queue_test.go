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
	"testing"
	time "time"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(q *queue)
		assert func(t *testing.T, q *queue)
	}{
		{
			name: "Dequeue returns nil on empty queue",
			assert: func(t *testing.T, q *queue) {
				if q.Dequeue() != nil {
					t.Fatal("expected nil")
				}
				testutil.AssertTrue(t, "empty", q.Empty())
			},
		},
		{
			name: "Dequeue returns log lines",
			setup: func(q *queue) {
				q.Enqueue(context.Background(), logLine{text: "a"})
				q.Enqueue(context.Background(), logLine{text: "b"})
				q.Enqueue(context.Background(), logLine{text: "c"})
			},
			assert: func(t *testing.T, q *queue) {
				testutil.AssertFalse(t, "empty", q.Empty())
				testutil.AssertEqual(t, "text", "a", q.Dequeue().text)
				testutil.AssertEqual(t, "text", "b", q.Dequeue().text)
				testutil.AssertEqual(t, "text", "c", q.Dequeue().text)
				if q.Dequeue() != nil {
					t.Fatal("expected nil")
				}
				testutil.AssertTrue(t, "empty", q.Empty())
			},
		},
		{
			name: "Peek returns log line without dequeuing it",
			setup: func(q *queue) {
				q.Enqueue(context.Background(), logLine{text: "a"})
				q.Enqueue(context.Background(), logLine{text: "b"})
			},
			assert: func(t *testing.T, q *queue) {
				testutil.AssertFalse(t, "empty", q.Empty())
				testutil.AssertEqual(t, "text", "a", q.Peek().text)
				testutil.AssertFalse(t, "empty", q.Empty())

				testutil.AssertEqual(t, "text", "a", q.Dequeue().text)
				testutil.AssertEqual(t, "text", "b", q.Dequeue().text)
				if q.Dequeue() != nil {
					t.Fatal("expected nil")
				}
				testutil.AssertTrue(t, "empty", q.Empty())
			},
		},
		{
			name: "Enqueue blocks if the queue is full",
			setup: func(q *queue) {
				q.Enqueue(context.Background(), logLine{text: "a"})
				q.Enqueue(context.Background(), logLine{text: "b"})
				q.Enqueue(context.Background(), logLine{text: "c"})
				q.Enqueue(context.Background(), logLine{text: "d"})
				q.Enqueue(context.Background(), logLine{text: "e"})
			},
			assert: func(t *testing.T, q *queue) {
				ctx, done := context.WithCancel(context.Background())
				go func() {
					q.Enqueue(context.Background(), logLine{text: "f"})
					defer done()
				}()

				// Make sure it blocks for a moment.
				{
					timer := time.NewTimer(250 * time.Millisecond)
					select {
					case <-timer.C:
						// Worked!
					case <-ctx.Done():
						// This shouldn't have happened...
						t.Fatal("expected Enqueue to block")
					}
				}

				// Read from the queue to unblock the Enqueue.
				testutil.AssertEqual(t, "text", "a", q.Dequeue().text)
				{
					timer := time.NewTimer(250 * time.Millisecond)
					select {
					case <-timer.C:
						// This shouldn't have happened...
						t.Fatal("expected Enqueue to unblock")
					case <-ctx.Done():
						// Worked!
					}
				}
			},
		},
		{
			name: "Enqueue respects the given context",
			setup: func(q *queue) {
				q.Enqueue(context.Background(), logLine{text: "a"})
				q.Enqueue(context.Background(), logLine{text: "b"})
				q.Enqueue(context.Background(), logLine{text: "c"})
				q.Enqueue(context.Background(), logLine{text: "d"})
				q.Enqueue(context.Background(), logLine{text: "e"})
			},
			assert: func(t *testing.T, q *queue) {
				ctx, done := context.WithCancel(context.Background())
				go func() {
					// Setup a context and cancel it. This should cause the
					// Enqueue method to exit even though the queue is full.
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					q.Enqueue(ctx, logLine{text: "f"})
					defer done()
				}()

				// Nothing should be blocking.
				{
					timer := time.NewTimer(250 * time.Millisecond)
					select {
					case <-timer.C:
						// This shouldn't have happened...
						t.Fatal("expected Enqueue to unblock")
					case <-ctx.Done():
						// Worked!
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q := newQueue(5)

			if tc.setup != nil {
				tc.setup(q)
			}

			if tc.assert != nil {
				tc.assert(t, q)
			}
		})
	}
}

func TestQueues_Empty(t *testing.T) {
	t.Parallel()

	queues := queues{
		newQueue(5),
		newQueue(5),
	}

	testutil.AssertTrue(t, "empty", queues.Empty())
}

func TestQueues_Empty_notEmpty(t *testing.T) {
	t.Parallel()

	queues := queues{
		newQueue(5),
		newQueue(5),
	}

	queues[0].Enqueue(context.Background(), logLine{})

	testutil.AssertFalse(t, "empty", queues.Empty())
}
