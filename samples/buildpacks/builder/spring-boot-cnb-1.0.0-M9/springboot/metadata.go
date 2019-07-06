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

package springboot

import (
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/manifest"
)

// Metadata describes the application's metadata.
type Metadata struct {
	// Classes indicates the Spring-Boot-Classes of a Spring Boot application.
	Classes string `mapstructure:"classes" properties:"Spring-Boot-Classes,default=" toml:"classes"`

	// Classpath is the classpath of a Spring Boot application.
	ClassPath []string `mapstructure:"classpath" properties:",default=" toml:"classpath"`

	// Lib indicates the Spring-Boot-Lib of a Spring Boot application.
	Lib string `mapstructure:"lib" properties:"Spring-Boot-Lib,default=" toml:"lib"`

	// StartClass indicates the Start-Class of a Spring Boot application.
	StartClass string `mapstructure:"start-class" properties:"Start-Class,default=" toml:"start-class"`

	// Version indicates the Spring-Boot-Version of a Spring Boot application.
	Version string `mapstructure:"version" properties:"Spring-Boot-Version,default=" toml:"version"`
}

func (m Metadata) Identity() (string, string) {
	return "Spring Boot", m.Version
}

// NewMetadata creates a new Metadata returning false if Spring-Boot-Version is not defined.
func NewMetadata(application application.Application, logger logger.Logger) (Metadata, bool, error) {
	md := Metadata{}

	m, err := manifest.NewManifest(application, logger)
	if err != nil {
		return Metadata{}, false, err
	}

	if err := m.Decode(&md); err != nil {
		return Metadata{}, false, err
	}

	if md.Version == "" {
		return Metadata{}, false, nil
	}

	j, err := helper.FindFiles(application.Root, regexp.MustCompile(".*\\.jar$"))
	if err != nil {
		return Metadata{}, false, err
	}

	md.ClassPath = append(md.ClassPath, filepath.Join(application.Root, md.Classes))
	md.ClassPath = append(md.ClassPath, j...)
	return md, true, nil
}
