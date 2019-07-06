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

package cli

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/cloudfoundry/jvm-application-cnb/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

var (
	beans   = regexp.MustCompile("beans[\\s]*{")
	logback = regexp.MustCompile(fmt.Sprintf(".*ch%[1]sqos%[1]slogback%[1]s.*.groovy", string(filepath.Separator)))
	pogo    = regexp.MustCompile("class [\\w]+[\\s\\w]*{")
)

// Command represents a Spring Boot CLI Command.
type Command struct {
	groovyFiles groovyFiles
	layer       layers.Layer
	layers      layers.Layers
}

// Contribute makes the contribution to launch.
func (c Command) Contribute() error {
	if err := c.layer.Contribute(c.groovyFiles, func(layer layers.Layer) error {
		return layer.AppendLaunchEnv("GROOVY_FILES", " %s", strings.Join(c.groovyFiles, " "))
	}, layers.Launch); err != nil {
		return err
	}

	command := "spring run -cp $CLASSPATH $GROOVY_FILES"

	return c.layers.WriteApplicationMetadata(layers.Metadata{
		Processes: layers.Processes{
			{"spring-boot-cli", command},
			{"task", command},
			{"web", command},
		},
	})
}

type groovyFiles []string

func (g groovyFiles) Identity() (string, string) {
	return "Groovy Files", fmt.Sprintf("(%d files)", len(g))
}

// NewCommand creates a new Command instance.
func NewCommand(build build.Build) (Command, bool, error) {
	_, ok := build.BuildPlan[jvmapplication.Dependency]
	if !ok {
		return Command{}, false, nil
	}

	candidates, err := candidates(build.Application.Root)
	if err != nil {
		return Command{}, false, err
	}

	if len(candidates) == 0 || !all(candidates, func(candidate string) bool {
		b, err := ioutil.ReadFile(candidate)
		if err != nil {
			return true // invalid files do not count against analysis
		}

		if !utf8.Valid(b) {
			return true // invalid files do not count against analysis
		}

		s := string(b)

		return pogo.MatchString(s) || beans.MatchString(s)
	}) {
		return Command{}, false, nil
	}

	return Command{
		groovyFiles(candidates),
		build.Layers.Layer("command"),
		build.Layers,
	}, true, nil
}

func all(candidates []string, predicate func(candidate string) bool) bool {
	for _, c := range candidates {
		if !predicate(c) {
			return false
		}
	}

	return true
}

func candidates(root string) ([]string, error) {
	c, err := helper.FindFiles(root, regexp.MustCompile(`.+\.groovy`))
	if err != nil {
		return nil, err
	}

	i := 0
	for _, s := range c {
		if !logback.MatchString(s) {
			c[i] = s
			i++
		}
	}

	return c[:i], nil
}
