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

package java

import (
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

// Credentials represents the google-stackdriver-credentials helper application.
type Credentials struct {
	buildpack buildpack.Buildpack
	layer     layers.HelperLayer
}

// Contributes makes the contribution to launch
func (c Credentials) Contribute() error {
	return c.layer.Contribute(func(artifact string, layer layers.HelperLayer) error {
		if err := helper.CopyFile(artifact, filepath.Join(layer.Root, "bin", "google-stackdriver-credentials")); err != nil {
			return err
		}

		return layer.WriteProfile("google-stackdriver-credentials", `printf "Configuring Google Stackdriver Credentials\n"

google-stackdriver-credentials %[1]s
export GOOGLE_APPLICATION_CREDENTIALS=%[1]s
`, filepath.Join(layer.Root, "google-stackdriver-credentials.json"))
	}, layers.Launch)
}

// NewCredentials creates a new Credentials instance.
func NewCredentials(build build.Build) Credentials {
	return Credentials{build.Buildpack, build.Layers.HelperLayer("google-stackdriver-credentials", "Google Stackdriver Credentials")}
}
