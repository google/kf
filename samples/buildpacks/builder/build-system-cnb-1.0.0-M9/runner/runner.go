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

package runner

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/runner"
)

// Runner represents the behavior of running the build system command to build an application.
type Runner struct {
	application           application.Application
	args                  []string
	bin                   string
	builtArtifactProvider BuiltArtifactProvider
	layer                 layers.Layer
	logger                logger.Logger
	runner                runner.Runner
}

// Contributes builds the application from source code, removes the source code, and expands the built artifact to
// $APPLICATION_ROOT.
func (r Runner) Contribute() error {
	c, err := NewCompiledApplication(r.application, r.runner)
	if err != nil {
		return err
	}

	if err := r.layer.Contribute(c, func(layer layers.Layer) error {
		if err := os.RemoveAll(layer.Root); err != nil {
			return err
		}

		if err := r.runner.Run(r.bin, r.application.Root, r.args...); err != nil {
			return err
		}

		artifact, err := r.builtArtifactProvider(r.application)
		if err != nil {
			return err
		}

		r.logger.Debug("Copying %s to %s", artifact, r.cachedApplication())
		return helper.CopyFile(artifact, r.cachedApplication())
	}); err != nil {
		return err
	}

	r.logger.Header("Removing source code")
	if cs, err := ioutil.ReadDir(r.application.Root); err != nil {
		return err
	} else {
		for _, c := range cs {
			if err := os.RemoveAll(filepath.Join(r.application.Root, c.Name())); err != nil {
				return err
			}
		}
	}

	r.logger.Debug("Expanding %s to %s", r.cachedApplication(), r.application.Root)
	return helper.ExtractZip(r.cachedApplication(), r.application.Root, 0)
}

func (r Runner) cachedApplication() string {
	return filepath.Join(r.layer.Root, "application.zip")
}

// BuildArtifactProvider returns the location of the build artifact.
type BuiltArtifactProvider func(application application.Application) (string, error)

func NewRunner(build build.Build, builtArtifactProvider BuiltArtifactProvider, bin string, args ...string) Runner {
	return Runner{
		build.Application,
		args,
		bin,
		builtArtifactProvider,
		build.Layers.Layer("build-system-application"),
		build.Logger,
		build.Runner,
	}
}
