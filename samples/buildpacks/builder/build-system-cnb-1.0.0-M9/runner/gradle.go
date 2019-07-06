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
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/build-system-cnb/buildsystem"
	"github.com/cloudfoundry/libcfbuildpack/build"
)

// NewRunner creates a new Gradle Runner instance.
func NewGradleRunner(build build.Build, buildSystem buildsystem.BuildSystem) Runner {
	return NewRunner(build, gradleBuiltArtifactProvider, buildSystem.Executable(), "-x", "test", "build")
}

func gradleBuiltArtifactProvider(application application.Application) (string, error) {
	target, ok := os.LookupEnv("BP_BUILT_ARTIFACT")
	if !ok {
		target = filepath.Join("build", "libs", "*.[jw]ar")
	}

	candidates, err := filepath.Glob(filepath.Join(application.Root, target))
	if err != nil {
		return "", err
	}

	if len(candidates) != 1 {
		return "", fmt.Errorf("unable to find built artifact in %s, candidates: %s", target, candidates)
	}

	return candidates[0], nil
}
