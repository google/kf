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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/archive-expanding-cnb/expand"
	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/detect"
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
	var c []string

	files, err := ioutil.ReadDir(detect.Application.Root)
	if err != nil {
		return -1, err
	}

	for _, f := range files {
		if regexp.MustCompile(".*\\.(jar|war|tar|tar\\.gz|tgz|zip)$").MatchString(f.Name()) {
			c = append(c, filepath.Join(detect.Application.Root, f.Name()))
		}
	}

	if c == nil || len(c) != 1 {
		return detect.Fail(), nil
	}

	bp := detect.BuildPlan[expand.Dependency]
	if bp.Metadata == nil {
		bp.Metadata = make(buildplan.Metadata)
	}
	bp.Metadata[expand.Archive] = c[0]

	return detect.Pass(buildplan.BuildPlan{
		expand.Dependency:         bp,
		jvmapplication.Dependency: detect.BuildPlan[jvmapplication.Dependency],
	})
}
