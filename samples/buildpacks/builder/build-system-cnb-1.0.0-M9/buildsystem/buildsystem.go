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

package buildsystem

import (
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// BuildSystem represents the build system distribution contributed by the buildpack.
type BuildSystem struct {
	contributor  layers.DependencyLayerContributor
	distribution string
	layer        layers.DependencyLayer
	logger       logger.Logger
	wrapper      string
}

// Contribute makes the contribution to the cache layer.
func (b BuildSystem) Contribute() error {
	if b.hasWrapper() {
		b.logger.Body("Using wrapper")
		return nil
	}

	return b.layer.Contribute(b.contributor, layers.Cache)
}

// Executable returns the path to the executable that should be used.  Will be the wrapper if it exists, the contributed
// build system distribution otherwise.
func (b BuildSystem) Executable() string {
	if b.hasWrapper() {
		return b.wrapper
	}

	return b.distribution
}

func (b BuildSystem) hasWrapper() bool {
	exists, err := helper.FileExists(b.wrapper)
	if err != nil {
		return false
	}

	return exists
}
