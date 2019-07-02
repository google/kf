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
	"fmt"
	"sort"
	"strings"

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/doctor"

	"github.com/spf13/cobra"
)

// DoctorTest represents a single high-level test. These should be roughly at
// the granularity of objects e.g. Apps, Service Brokers, Namespaces, and Clusters.
type DoctorTest struct {
	Name string
	Test doctor.Diagnosable
}

// NewDoctorCommand creates a new doctor command for the given tests.
// The tests will be executed in the order they appear in the list as long as
// they're requested by the user.
func NewDoctorCommand(p *config.KfParams, tests []DoctorTest) *cobra.Command {
	var knownTestNames []string
	for _, t := range tests {
		knownTestNames = append(knownTestNames, t.Name)
	}
	sort.Strings(knownTestNames)

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

			toTest := make(map[string]bool)
			for _, arg := range args {
				toTest[arg] = true
			}

			d := doctor.NewDiagnostic("doctor", cmd.OutOrStdout())
			for _, dt := range tests {
				if _, ok := toTest[dt.Name]; !ok {
					continue
				}

				d.GatedRun(dt.Name, dt.Test.Diagnose)
			}

			// Report
			if d.Failed() {
				fmt.Fprintln(cmd.OutOrStdout(), "FAIL")
				return errors.New("environment failed checks")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "PASS")
			return nil
		},
	}

	return doctorCmd
}
