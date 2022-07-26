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
	"sync"
)

type queue struct {
	c chan logLine

	mu   sync.Mutex
	head *logLine
}

func newQueue(size int) *queue {
	return &queue{
		c: make(chan logLine, size),
	}
}

// Peek implements dequeuer.
func (q *queue) Peek() *logLine {
	q.mu.Lock()
	defer q.mu.Unlock()

	// First check to see if we have anything saved in head. If so, return it.
	if q.head != nil {
		head := q.head
		return head
	}

	// Looks like head is empty, so set it.
	select {
	case ll := <-q.c:
		q.head = &ll
	default:
		q.head = nil
	}

	return q.head
}

// Dequeue implements dequeuer.
func (q *queue) Dequeue() *logLine {
	// First check to see if we have anything saved in head. If so, clear it
	// and reteurn it.
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.head != nil {
		head := q.head
		q.head = nil
		return head
	}

	select {
	case ll := <-q.c:
		return &ll
	default:
		return nil
	}
}

func (q *queue) Empty() bool {
	// First check to see if we have anything saved in head. If so, we're not
	// empty.
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head != nil {
		return false
	}

	return len(q.c) == 0
}

// Enqueue implements enqueuer.
func (q *queue) Enqueue(ctx context.Context, ll logLine) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.c <- ll:
		return nil
	}
}

type enqueuer interface {
	// Enqueue enqueues a log line into the queue. It will block if the queue
	// is full unless an entry is dequeued or the context is cancelled. It
	// will return an error if the context is cancelled, otherwise it will
	// return nil.
	Enqueue(context.Context, logLine) error
}

type dequeuer interface {
	// Dequeue dequeues from the queue. If it is empty, then it returns nil.
	// Therefore, it does not block.
	Dequeue() *logLine

	// Peek returns the next log line (without removing it).
	Peek() *logLine
}

type queues []*queue

func (qs queues) Empty() bool {
	for _, q := range qs {
		if !q.Empty() {
			return false
		}
	}

	return true
}
