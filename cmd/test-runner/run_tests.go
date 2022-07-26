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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"knative.dev/pkg/logging"
)

const (
	OutputAction   = "output"
	RunAction      = "run"
	PauseAction    = "pause"
	ContinueAction = "cont"
	FailAction     = "fail"
	PassAction     = "pass"
	SkipAction     = "skip"
)

var (
	passColor = color.New(color.FgGreen)
	failColor = color.New(color.FgRed)
)

type test struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
}

type testResult struct {
	Name                 string
	Package              string
	Start                time.Time
	RemainingFailAttemps int
	Pass                 bool
}

func (tr *testResult) fullName() string {
	return fmt.Sprintf("%s#%s", tr.Package, tr.Name)
}

// runTests runs the given tests and returns the failed tests that can be
// retried. If tests have exhausted their retries, it will return an error.
func runTests(
	ctx context.Context,
	pkg string,
	runRegexp string,
	previousTestResults []*testResult,
	attempts int,
	timeout time.Duration,
) (
	passed []*testResult,
	retry []*testResult,
	failed []*testResult,
	err error,
) {
	logger := logging.FromContext(ctx)
	m := make(map[string]*testResult)
	args := []string{"test", "-json"}
	args = append(args, pkg)
	args = append(args, "--timeout="+timeout.String())

	// Create the unified the --run argument.
	{
		var previousFailedTests []string
		for _, tr := range previousTestResults {
			m[tr.fullName()] = tr
			previousFailedTests = append(previousFailedTests, fmt.Sprintf("(^%s$)", tr.Name))
		}
		if len(previousTestResults) > 0 {
			args = append(args, "--run="+strings.Join(previousFailedTests, "|"))
		} else if runRegexp != "" {
			// We only want to set the "--run" flag this way if we don't have
			// tests to retry.
			args = append(args, "--run="+runRegexp)
		}
	}

	stdout, wait, err := runCommand(ctx, "go", args)
	if err != nil {
		return nil, nil, nil, err
	}

	testResults := consumeTests(ctx, stdout, m, attempts)

	if err := wait(); err != nil && len(testResults) == 0 {
		// We only want to fail if we didn't find any tests (likely because of
		// a compilation error).
		return nil, nil, nil, err
	}

	// Retry the failed tests.
	for _, tr := range testResults {
		if tr.Pass {
			logger.Infof("%s %s", tr.Name, passColor.Sprint("PASSED"))
			passed = append(passed, tr)
			continue
		}

		if tr.RemainingFailAttemps <= 0 {
			logger.Infof("%s %s", tr.Name, failColor.Sprint("FAILED"))
			failed = append(failed, tr)
			continue
		}

		retry = append(retry, tr)
	}

	return passed, retry, failed, nil
}

// parseTest JSON decodes the data. If the value fails, it does its best to
// create a test object from the line.
func parseTest(data string) test {
	var t test
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		// NOP, if something is logged outside of the test suite, then
		// it will likely not be formatted this way. Just do our best
		// to guess.
		t.Time = time.Now()
		t.Action = OutputAction
		t.Package = "UNKNOWN"
		t.Test = "UNKNOWN"
	}

	// Remove any extra whitespace.
	t.Output = strings.TrimSpace(t.Output)

	return t
}

// consumeTests reads from the channel and saves the results into the given
// map. If the test is unknown, it initializes it. It any log line to stderr
// prefixed with the associated package/test.
func consumeTests(
	ctx context.Context,
	stdout io.Reader,
	m map[string]*testResult,
	attempts int,
) map[string]*testResult {
	logger := logging.FromContext(ctx)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		t := parseTest(scanner.Text())
		if t.Test == "" {
			// We don't care about package actions.
			continue
		}

		testName := fmt.Sprintf("%s#%s", t.Package, t.Test)

		switch t.Action {
		case OutputAction:
			logger.Infof("%s %s", t.Time, t.Output)
		case RunAction:
			result := m[testName]
			if result == nil {
				result = &testResult{
					Name:                 t.Test,
					Package:              t.Package,
					RemainingFailAttemps: attempts,
				}
			}

			result.Start = t.Time
			m[testName] = result
			logger.Infof("%s Start", t.Time)
		case PauseAction:
			logger.Infof("%s Pause", t.Time)
		case ContinueAction:
			logger.Infof("%s Continue", t.Time)
		case SkipAction:
			logger.Infof("%s Skip", t.Time)
			testResult := m[testName]
			testResult.Pass = true
			m[testName] = testResult
		case FailAction:
			remainingFailAttemps := m[testName].RemainingFailAttemps
			logger.Infof("%s Fail after %v %v %v", t.Time, t.Time.Sub(m[testName].Start), t.Time, m[testName].Start)
			testResult := m[testName]
			testResult.RemainingFailAttemps = remainingFailAttemps - 1
			m[testName] = testResult
		case PassAction:
			logger.Infof("%s Pass after %v %v %v", t.Time, t.Time.Sub(m[testName].Start), t.Time, m[testName].Start)
			testResult := m[testName]
			testResult.Pass = true
			m[testName] = testResult
		default:
			logger.Fatalf("Unknown action: %s", t.Action)
		}
	}

	return m
}
