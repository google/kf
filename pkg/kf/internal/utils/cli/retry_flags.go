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

package utils

import (
	"time"

	"github.com/spf13/cobra"
)

// RetryFlags is a flag set for managing how a command retries attempts.
type RetryFlags struct {
	retries int
	delay   time.Duration
}

// Retry repeats the command the specified number of times
// with a delay.
func (flags *RetryFlags) Retry(callback func() error) error {
	// Retries n-1 times and ignores failures. Only return early on success.
	for i := 1; i < flags.retries; i++ {
		if err := callback(); err == nil {
			return nil
		}

		time.Sleep(flags.delay)
	}

	// The last error (if any) is the one to keep.
	return callback()
}

// Add adds the retry flags to the Cobra command.
func (flags *RetryFlags) Add(cmd *cobra.Command, retries int, delay time.Duration) {
	cmd.Flags().DurationVar(
		&flags.delay,
		"retry-delay",
		delay,
		"Set the delay between retries.",
	)

	cmd.Flags().IntVar(
		&flags.retries,
		"retries",
		retries,
		"Number of times to retry execution if the command isn't successful.",
	)
}

// AddRetryForK8sPropagation sets up a retrier suitable for handling
// K8s propagation delays. It's guaranteed to retry more than once with a delay
// of >= 1 second, but specific backoff algorithm, times, delays, or jitter
// are subject to change.
func (flags *RetryFlags) AddRetryForK8sPropagation(cmd *cobra.Command) {
	flags.Add(cmd, 5, 1*time.Second)
}
