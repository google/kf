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
	"k8s.io/client-go/kubernetes"
)

func TestWindowedReader(t *testing.T) {
	t.Parallel()

	now := time.Now()
	start := now.Add(-5 * time.Second)
	end := now

	testCases := []struct {
		name      string
		returnErr error
		logLines  []logLine
		assert    func(t *testing.T, ll []logLine, err error)
	}{
		{
			name: "truncates logs to window",
			logLines: []logLine{
				{text: "a", timestamp: start},
				{text: "b", timestamp: start.Add(time.Second)},
				// This logLine should not be included as it is outside the
				// window.
				{text: "c", timestamp: start.Add(6 * time.Second)},
			},
			assert: func(t *testing.T, ll []logLine, err error) {
				testutil.AssertErrorsEqual(t, nil, err)
				testutil.AssertEqual(t, "len(ll)", 2, len(ll))
			},
		},
		{
			name: "sort the logs in ascending order",
			logLines: []logLine{
				{text: "b", timestamp: start.Add(time.Second)},
				{text: "a", timestamp: start},
				{text: "c", timestamp: start.Add(2 * time.Second)},
			},
			assert: func(t *testing.T, ll []logLine, err error) {
				testutil.AssertErrorsEqual(t, nil, err)

				names := []string{}
				for _, l := range ll {
					names = append(names, l.text)
				}

				testutil.AssertEqual(t, "names", []string{"a", "b", "c"}, names)
			},
		},
		{
			name:      "returns the error",
			returnErr: errors.New("some-error"),
			logLines: []logLine{
				{text: "a", timestamp: start},
				{text: "b", timestamp: start.Add(time.Second)},
				{text: "c", timestamp: start.Add(2 * time.Second)},
			},
			assert: func(t *testing.T, ll []logLine, err error) {
				testutil.AssertErrorsEqual(t, errors.New("some-error"), err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &windowedLogReader{
				streamLogs: func(
					ctx context.Context,
					client kubernetes.Interface,
					pod *corev1.Pod,
					opts corev1.PodLogOptions,
					f func(logLine) bool,
				) error {
					testutil.AssertEqual(t, "container", opts.Container, "some-container")
					testutil.AssertTrue(t, "timestamps", opts.Timestamps)
					testutil.AssertEqual(t, "SinceTime", opts.SinceTime.Time, start)

					for _, ll := range tc.logLines {
						f(ll)
					}

					return tc.returnErr
				},
			}

			logLines, err := r.Read(
				context.Background(),
				start,
				end,
				&corev1.Pod{},
				"some-container",
			)

			if tc.assert != nil {
				tc.assert(t, logLines, err)
			}
		})
	}
}
