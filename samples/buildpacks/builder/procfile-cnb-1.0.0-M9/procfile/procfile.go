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

package procfile

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// Dependency indicates that an application runs on the JVM.
const Dependency = "procfile"

var pattern = regexp.MustCompile("^([A-Za-z0-9_-]+):\\s*(.+)$")

type Procfile struct {
	layers    layers.Layers
	logger    logger.Logger
	processes map[string]string
}

// Contribute makes the contribution to launch
func (p Procfile) Contribute() error {
	p.logger.Body("Contributing Procfile process types")

	var processes []layers.Process
	for t, c := range p.processes {
		processes = append(processes, layers.Process{
			Type:    t,
			Command: c,
		})
	}

	return p.layers.WriteApplicationMetadata(layers.Metadata{Processes: processes})
}

// NewProcfile creates a new Procfile instance.  OK is true if the build plan contains "procfile" dependency.
func NewProcfile(build build.Build) (Procfile, bool) {
	p, ok := build.BuildPlan[Dependency]
	if !ok {
		return Procfile{}, false
	}

	processes := make(map[string]string)
	for t, c := range p.Metadata {
		s, ok := c.(string)
		if !ok {
			build.Logger.Debug("Ignoring type '%s' because command '%s' is not a string", t, c)
			continue
		}

		processes[t] = s
	}

	return Procfile{
		build.Layers,
		build.Logger,
		processes,
	}, true
}

// ParseProcfile returns the contents of a procfile and true if the application contains a Procfile file, otherwise
// false.
func ParseProcfile(application application.Application, logger logger.Logger) (map[string]string, bool, error) {
	procfile := filepath.Join(application.Root, "Procfile")

	exists, err := helper.FileExists(procfile)
	if err != nil {
		return nil, false, err
	}

	if !exists {
		return nil, false, nil
	}

	file, err := os.OpenFile(procfile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	p := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := pattern.FindStringSubmatch(scanner.Text())
		if len(parts) > 0 {
			p[parts[1]] = parts[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, false, err
	}

	logger.Debug("Procfile: %s", p)
	return p, true, nil
}
