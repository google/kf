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

package utils

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

// AsyncFlags is a flag set for managing whether or not a command runs
// asynchronously.
type AsyncFlags struct {
	async bool
}

// Add adds the async flag to the Cobra command.
func (flags *AsyncFlags) Add(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&flags.async,
		"async",
		"",
		false,
		"Do not wait for the action to complete on the server before returning.",
	)
}

// IsAsync returns true if the user wanted the operation to run asynchronously.
func (flags *AsyncFlags) IsAsync() bool {
	return flags.async
}

// IsSynchronous returns true if the user wants the operation to be completed
// synchronously.
func (flags *AsyncFlags) IsSynchronous() bool {
	return !flags.async
}

// AwaitAndLog waits for the action to be completed if the flag specifies the
// command should run synchronously. In either case, it will notify the user
// of the decision by logging to the writer whether it waited or not.
// If an error is returned by the callback the result will be an error,
// otherwise the error will be nil.
func (flags *AsyncFlags) AwaitAndLog(w io.Writer, action string, callback func() error) error {
	if flags.IsSynchronous() {
		fmt.Fprintf(w, "%s...\n", action)
		if err := callback(); err != nil {
			return err
		}

		fmt.Fprintln(w, "Success")
	} else {
		fmt.Fprintf(w, "%s asynchronously\n", action)
	}

	return nil
}

// WaitFor waits for the action to be completed (signified by the callback
// returning true) if the flag specifies the command should be run
// synchronously. In either case, it will notify the user of the decision by
// logging to the writer whether it waited or not. If an error is returned by
// the callback the result will be an error, otherwise the error will be nil.
func (flags *AsyncFlags) WaitFor(
	ctx context.Context,
	w io.Writer,
	action string,
	interval time.Duration,
	callback func() (bool, error),
) error {
	return flags.AwaitAndLog(w, action, func() error {
		tick := time.NewTicker(interval)
		defer tick.Stop()

		for {
			if done, err := callback(); err != nil {
				return err
			} else if done {
				return nil
			}

			select {
			case <-tick.C:
				// Continue waiting for callback.
			case <-ctx.Done():
				return fmt.Errorf("%s timed out", action)
			}
		}
	})
}

type AsyncIfStoppedFlags struct {
	AsyncFlags
	no_short_circuit_wait bool
}

// Add adds the async and async_if_stopped flags to the cobra command
func (flags *AsyncIfStoppedFlags) Add(cmd *cobra.Command) {
	flags.AsyncFlags.Add(cmd)

	cmd.Flags().BoolVarP(
		&flags.no_short_circuit_wait,
		"no-short-circuit-wait",
		"",
		false,
		"Allow the CLI to skip waiting if the mutation does not impact a running resource.",
	)
}

// IsAsyncIfStopped returns true if the user wanted the operation to run asynchronously if the app is stopped.
func (flags *AsyncIfStoppedFlags) IsAsyncIfStopped() bool {
	return !flags.no_short_circuit_wait
}

// AwaitANdLog waits for the application to be completed if the flags and the
// app's stopped status determine that the command should run synchronously.
// In either case, it will notify the user of the decision by logging to the
// writer whether it waited or not. If an error is returned by the callback the
// result will be an error, otherwise the error will be nil
func (flags *AsyncIfStoppedFlags) AwaitAndLog(stopped bool, w io.Writer, action string, callback func() error) error {
	if stopped && flags.IsAsyncIfStopped() {
		fmt.Fprintf(w, "%s asynchronously because app is stopped\n", action)
		return nil
	}

	return flags.AsyncFlags.AwaitAndLog(w, action, callback)
}
