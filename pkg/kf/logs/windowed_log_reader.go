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
	"sort"
	time "time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type windowReader interface {
	// Read reads logs within a time window [start, end).
	Read(
		ctx context.Context,
		start time.Time,
		end time.Time,
		pod *corev1.Pod,
		container string,
	) ([]logLine, error)
}

// windowedLogReader reads blocks of time from a container on a Pod.
type windowedLogReader struct {
	client kubernetes.Interface

	// XXX: This function is written in a way that it can be replaced for
	// testing purposes ONLY. It should NEVER be replaced for production use.
	// There does not yet exist a solid way to test the internal functions
	// outside of an integration test. Therefore the unit tests needed a way,
	// and that way is to replace this function. It defaults to streamLogs.
	streamLogs streamLogsFunc
}

func newWindowedLogReader(client kubernetes.Interface) windowReader {
	return &windowedLogReader{
		client:     client,
		streamLogs: streamLogs,
	}
}

// Read implements windowReader.
func (r *windowedLogReader) Read(
	ctx context.Context,
	start time.Time,
	end time.Time,
	pod *corev1.Pod,
	container string,
) ([]logLine, error) {

	// We don't set TailLines because normally we'll be reading from multiple
	// pods and therefore we don't know how many log lines to read from each
	// pod.
	opts := corev1.PodLogOptions{
		Container:  container,
		Timestamps: true,
		SinceTime:  &metav1.Time{Time: start},
	}

	var result logLines
	if err := r.streamLogs(ctx, r.client, pod, opts, func(ll logLine) bool {
		if ll.timestamp.UnixNano() >= end.UnixNano() {
			// Outside the given time window, bail.
			return false
		}

		result = append(result, ll)
		return true
	}); err != nil {
		return nil, err
	}

	// Sort the logs in ascending order.
	sort.Sort(result)

	return result, nil
}
