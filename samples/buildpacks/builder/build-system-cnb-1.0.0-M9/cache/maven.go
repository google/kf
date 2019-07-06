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

package cache

import (
	"os/user"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

// NewMavenCache creates a new Cache instance for Maven.
func NewMavenCache(build build.Build) (Cache, error) {
	u, err := user.Current()
	if err != nil {
		return Cache{}, err
	}

	destination := filepath.Join(u.HomeDir, ".m2")
	build.Logger.Debug(".m2 directory: %s", destination)

	return NewCache(build, destination)
}
