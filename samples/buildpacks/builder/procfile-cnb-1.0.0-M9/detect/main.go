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

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/procfile-cnb/procfile"
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
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
	p, ok, err := procfile.ParseProcfile(detect.Application, detect.Logger)
	if err != nil {
		return detect.Error(102), err
	}

	if !ok {
		return detect.Fail(), nil
	}

	bp := detect.BuildPlan[procfile.Dependency]
	if bp.Metadata == nil {
		bp.Metadata = make(buildplan.Metadata)
	}

	for t, c := range p {
		bp.Metadata[t] = c
	}

	return detect.Pass(buildplan.BuildPlan{procfile.Dependency: bp})
}
