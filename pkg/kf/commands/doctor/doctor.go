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
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	failColor = color.New(color.FgHiRed, color.Bold)
	passColor = color.New(color.FgHiGreen, color.Bold)

	testsNotRunByDefault = []string{"operator"}
)

// DoctorTest represents a single high-level test. These should be roughly at
// the granularity of objects e.g. Apps, Service Brokers, Namespaces, and Clusters.
type DoctorTest struct {
	Name string
	Test doctor.Diagnosable
}

// ComponentCloser is a function that can bind a Component to a specific object
// for testing.
type ComponentCloser interface {
	Close(component troubleshooter.Component, resourceName string) doctor.Diagnosable
}

// NewDoctorCommand creates a new doctor command for the given tests.
// The tests will be executed in the order they appear in the list as long as
// they're requested by the user.
func NewDoctorCommand(
	p *config.KfParams,
	tests []DoctorTest,
	objectTests []troubleshooter.Component,
	closer ComponentCloser,
) *cobra.Command {
	componentTestMap := make(map[string]DoctorTest)
	for _, t := range tests {
		componentTestMap[t.Name] = t
	}
	componentSet := sets.StringKeySet(componentTestMap)
	componentSet.Delete(testsNotRunByDefault...)

	componentTestNames := componentSet.List()

	objectTestMap := make(map[string]troubleshooter.Component)
	for _, t := range objectTests {
		objectTestMap[strings.ToLower(t.Type.FriendlyName())] = t
	}
	objectTestNames := sets.StringKeySet(objectTestMap).List()

	var (
		retries int
		delay   time.Duration
	)

	doctorCmd := &cobra.Command{
		Use:   "doctor [(COMPONENT|TYPE/NAME)...]",
		Short: "Run validation tests against one or more components.",
		Example: `
		# Run doctor against all components.
		kf doctor
		# Run doctor against server-side components.
		kf doctor cluster
		# Run doctor for a Kf App named my-app.
		kf doctor app/my-app
		# Run doctor for a Kf Service named my-service.
		kf doctor serviceinstance/my-service
		# Run doctor for a Kf Binding named my-binding.
		kf doctor serviceinstancebinding/my-binding
		# Run doctor for the Kf Operator.
		kf doctor operator
		`,
		Long: fmt.Sprintf(`Doctor runs tests on one or more components or objects to
		validate their desired status.

		If no arguments are supplied, then all component tests are ran.
		If one or more arguments are suplied then only the test for those
		components or objects are ran.

		Possible components are:

		* %s

		Possible object types are:

		* %s
		`,
			strings.Join(componentTestNames, "\n		* "),
			strings.Join(objectTestNames, "\n		* ")),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Run all component tests if nothing is specified.
			if len(args) == 0 {
				args = componentTestNames
			}

			var testsToRun []DoctorTest

			for _, arg := range args {
				argSplit := strings.Split(arg, "/")
				switch len(argSplit) {
				case 1:
					name := argSplit[0]
					test, ok := componentTestMap[name]
					if !ok {
						return fmt.Errorf(
							"unknown component %q, supported components: %s",
							name,
							strings.Join(componentTestNames, "|"),
						)
					}
					testsToRun = append(testsToRun, test)

				case 2:
					objType, objName := argSplit[0], argSplit[1]
					test, ok := objectTestMap[objType]
					if !ok {
						return fmt.Errorf(
							"unknown object type %q, supported object types: %s",
							objType,
							strings.Join(objectTestNames, "|"),
						)
					}

					testsToRun = append(testsToRun, DoctorTest{
						Name: objType,
						Test: closer.Close(test, objName),
					})

				default:
					return fmt.Errorf("malformed argument: %q", arg)
				}
			}

			return retry(retries, delay, func() error {
				return runDoctor(cmd.Context(), cmd.OutOrStdout(), testsToRun)
			})
		},
	}

	doctorCmd.Flags().DurationVar(&delay, "delay", 5*time.Second, "Set the delay between executions.")
	doctorCmd.Flags().IntVar(&retries, "retries", 1, "Number of times to retry doctor if it isn't successful.")

	return doctorCmd
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

func runDoctor(ctx context.Context, w io.Writer, desiredTests []DoctorTest) error {
	d := doctor.NewDiagnostic("doctor", w)
	for _, dt := range desiredTests {
		d.GatedRun(ctx, dt.Name, dt.Test.Diagnose)
	}

	// Report
	if d.Failed() {
		failColor.Fprintln(w, "FAIL")
		return errors.New("environment failed checks")
	}

	passColor.Fprintln(w, "PASS")
	return nil
}
