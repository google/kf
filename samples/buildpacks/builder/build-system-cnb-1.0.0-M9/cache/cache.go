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
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// Cache represents the location that a build system caches its downloaded artifacts for reuse.
type Cache struct {
	destination string
	layer       layers.Layer
	logger      logger.Logger
}

// Contribute links the cache layer to the destination if it does not already exist.
func (c Cache) Contribute() error {
	exists, err := helper.FileExists(c.destination)
	if err != nil {
		return err
	}

	if exists {
		c.logger.Debug("Cache destination already exists")
		return nil
	}

	c.logger.Body("Linking Cache to %s", c.destination)

	c.layer.Touch()

	c.logger.Debug("Creating cache directory %s", c.layer.Root)
	if err := os.MkdirAll(c.layer.Root, 0755); err != nil {
		return err
	}

	parent := filepath.Dir(c.destination)
	c.logger.Debug("Creating destination parent directory %s", parent)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	c.logger.Debug("Linking %s => %s", c.layer.Root, c.destination)
	if err := os.Symlink(c.layer.Root, c.destination); err != nil {
		return err
	}

	return c.layer.WriteMetadata(nil, layers.Cache)
}

// NewCache creates a new Cache instance.
func NewCache(build build.Build, destination string) (Cache, error) {
	return Cache{
		destination,
		build.Layers.Layer("build-system-cache"),
		build.Logger,
	}, nil
}
