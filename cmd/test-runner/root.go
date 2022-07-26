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

package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// NewRootCommand returns a root command for the test-runner.
func NewRootCommand() *cobra.Command {
	var (
		attempts  int
		timeout   time.Duration
		runRegexp string
	)

	cmd := &cobra.Command{
		Use:   "test-runner",
		Short: "test-runner runs go test but with retries",
		Args:  cobra.ExactArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			ctx = configlogging.SetupLogger(ctx, cmd.ErrOrStderr())
			cmd.SetContext(ctx)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx := cmd.Context()
			pkg := args[0]

			var (
				retryTests  []*testResult
				passedTests []*testResult
				failedTests []*testResult
			)

			defer func() {
				// Write report so we can easily tell what failed and what
				// passed.
				tw := tabwriter.NewWriter(cmd.OutOrStderr(), 0, 2, 1, ' ', tabwriter.AlignRight)

				// Sort the tests by name.
				allTests := append(append(passedTests, retryTests...), failedTests...)
				sort.Slice(allTests, func(i, j int) bool {
					trI, trJ := allTests[i], allTests[j]
					return trI.Name < trJ.Name
				})

				// Write the FAILS before the final PASS to show how many
				// retries the test took.
				for _, tr := range allTests {
					fmt.Fprint(tw, tr.Name+"\t")
					numTimesFailed := attempts - tr.RemainingFailAttemps
					fmt.Fprint(tw, strings.Repeat(failColor.Sprint("FAIL")+"\t", numTimesFailed))
					if tr.Pass {
						fmt.Fprint(tw, passColor.Sprint("PASS")+"\t")
					}
					fmt.Fprintln(tw)
				}
				tw.Flush()
				logging.FromContext(ctx).Infof("%d/%d tests passed", len(passedTests), len(passedTests)+len(retryTests)+len(failedTests))
			}()

			for ctx.Err() == nil {
				var err error
				var passed, failed []*testResult
				passed, retryTests, failed, err = runTests(ctx, pkg, runRegexp, retryTests, attempts, timeout)
				if err != nil {
					return err
				}

				// Keep track of the passed tests.
				passedTests = append(passedTests, passed...)
				failedTests = failed

				if len(failed) > 0 {
					// If we have failed tests, then we should short circuit.
					return errors.New("tests failed")
				}

				if len(retryTests) > 0 {
					continue
				}

				return nil
			}

			return ctx.Err()
		},
	}

	cmd.Flags().IntVar(&attempts, "attempts", 0, "the number of attempts after a failed test to try again")
	cmd.Flags().StringVar(&runRegexp, "run", "", "the run regexp to start with")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "the timeout for a test to take")

	return cmd
}
