/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/google-stackdriver-cnb/java"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(101)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(101)
	}

	if code, err := d(detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func d(detect detect.Detect) (int, error) {
	bp := buildplan.BuildPlan{}

	if _, ok := detect.BuildPlan[jvmapplication.Dependency]; ok {
		if detect.Services.HasService("google-stackdriver-debugger", "PrivateKeyData") {
			bp[java.DebuggerDependency] = detect.BuildPlan[java.DebuggerDependency]
		}

		if detect.Services.HasService("google-stackdriver-profiler", "PrivateKeyData") {
			bp[java.ProfilerDependency] = detect.BuildPlan[java.ProfilerDependency]
		}
	}

	if reflect.DeepEqual(bp, buildplan.BuildPlan{}) {
		return detect.Fail(), nil
	}

	return detect.Pass(bp)
}
