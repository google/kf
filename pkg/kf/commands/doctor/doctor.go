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

package doctor

import (
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/doctor"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	failColor = color.New(color.FgHiRed, color.Bold)
	passColor = color.New(color.FgHiGreen, color.Bold)
)

// Test represents a single high-level doctor test. These should be roughly at
// the granularity of objects e.g. Apps, Service Brokers, Namespaces, and Clusters.
type Test struct {
	Name string
	Test doctor.Diagnosable
}

// NewDoctorCommand creates a new doctor command for the given tests.
// The tests will be executed in the order they appear in the list as long as
// they're requested by the user.
func NewDoctorCommand(p *config.KfParams, tests []Test) *cobra.Command {
	var knownTestNames []string
	for _, t := range tests {
		knownTestNames = append(knownTestNames, t.Name)
	}
	sort.Strings(knownTestNames)

	var (
		retries int
		delay   time.Duration
	)

	doctorCmd := &cobra.Command{
		Use:     "doctor [COMPONENT...]",
		Short:   "Doctor runs validation tests against one or more components",
		Example: `  kf doctor cluster`,
		Long: `Doctor runs tests one or more components to validate them.

		If no arguments are supplied, then all tests are run.
		If one or more arguments are suplied then only those components are run.

		Possible components are: ` + strings.Join(knownTestNames, ", "),
		ValidArgs: knownTestNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cobra.OnlyValidArgs(cmd, args); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			// Run everything if nothing is specified
			if len(args) == 0 {
				args = knownTestNames
			}

			matchingTests := testsMatching(args, tests)

			return retry(retries, delay, func() error {
				return runDoctor(cmd.OutOrStdout(), matchingTests)
			})
		},
	}

	doctorCmd.Flags().DurationVar(&delay, "delay", 5*time.Second, "Set the delay between executions")
	doctorCmd.Flags().IntVar(&retries, "retries", 1, "Number of times to retry doctor if it isn't successful")

	return doctorCmd
}

// testsMatching returns all tests with names in the given set.
// if no tests are specified as desired, all get run.
func testsMatching(desired []string, knownTests []Test) []Test {
	if len(desired) == 0 {
		return knownTests
	}

	desiredSet := sets.NewString(desired...)

	var out []Test
	for _, dt := range knownTests {
		if desiredSet.Has(dt.Name) {
			out = append(out, dt)
		}
	}

	return out
}

func retry(times int, delay time.Duration, callback func() error) error {
	// Retries n-1 times and ignores failures. Only return early on success.
	for i := 1; i < times; i++ {
		if err := callback(); err == nil {
			return nil
		}

		time.Sleep(delay)
	}

	// The last error (if any) is the one to keep.
	return callback()
}

func runDoctor(w io.Writer, desiredTests []Test) error {
	d := doctor.NewDiagnostic("doctor", w)
	for _, dt := range desiredTests {
		d.GatedRun(dt.Name, dt.Test.Diagnose)
	}

	// Report
	if d.Failed() {
		if _, err := failColor.Fprintln(w, "FAIL"); err != nil {
			return err
		}
		return errors.New("environment failed checks")
	}

	if _, err := passColor.Fprintln(w, "PASS"); err != nil {
		return err
	}
	return nil
}
