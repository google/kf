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
	time "time"

	corev1 "k8s.io/api/core/v1"
)

// reverseReader reads log lines from a pod in reverse.
//
// NOTE: Some notes to future developers who might want to optimize.
//
// The Kubernetes API allows one to read logs from either a specific time. For
// recent logs, we need to read several streams to combine them into a block
// of N length. We don't know ahead of time which Pods will have the desired
// (most recent) logs. Therefore we have to read from all of them.
//
// If the Kubernetes API allowed us to read N logs from a specific point in
// time, we could likely get rid of this. If we could read backwards from
// a specif point of time, again, we could get rid of this. However, given we
// can't, this is used to allow the outer algorithm to read backwards from
// several streams to pick and choose which logs it wants to collect.
type reverseReader interface {
	// Read reads logs from the given Pod container in reverse. It does this
	// by reading in small time windows.
	Read(
		ctx context.Context,
		pod *corev1.Pod,
		container string,
		q enqueuer,
	) error
}

// reverseLogReader reads logs from a Pod in reverse.
type reverseLogReader struct {
	windowReader windowReader
}

func newReverseLogReader(windowReader windowReader) reverseReader {
	return &reverseLogReader{
		windowReader: windowReader,
	}
}

// Read implements reverseReader.
func (r *reverseLogReader) Read(
	ctx context.Context,
	pod *corev1.Pod,
	container string,
	q enqueuer,
) error {
	// Set these as the same value for now. The loop ALWAYS adjusts next.
	next := time.Now()
	end := next

	// Loop until the context is done or we are looking for logs before the
	// Pod was created.
	for ctx.Err() == nil && next.UnixNano() >= pod.CreationTimestamp.UnixNano() {
		// Always walk back 5 seconds.
		next = next.Add(-5 * time.Second)

		// Read a time window of log lines for the container.
		lines, err := r.windowReader.Read(ctx, next, end, pod, container)
		if err != nil {
			return err
		}

		if len(lines) == 0 {
			// We didn't find anything in that block, wait a moment and try
			// again.
			timer := time.NewTimer(250 * time.Millisecond)
			select {
			case <-timer.C:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Set the end of the next window where this one started.
		end = lines[0].timestamp

		// Enqueue all the logs in reverse order.
		for i := range lines {
			// We don't actually care about the error here. The only reason it
			// would return one is because the context is done and therefore
			// we'll return that error anyways.
			_ = q.Enqueue(ctx, lines[len(lines)-i-1])
		}
	}

	return ctx.Err()
}
