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

package executablejar

import (
	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/manifest"
)

// Metadata describes the metadata
type Metadata struct {
	// Classpath is the classpath of the executable JAR.
	ClassPath []string `mapstructure:"classpath" properties:",default=" toml:"classpath"`

	// MainClass is the Main-Class of the executable JAR.
	MainClass string `mapstructure:"main-class" properties:"Main-Class,default=" toml:"main-class"`
}

func (Metadata) Identity() (string, string) {
	return "Executable JAR", ""
}

// NewMetadata creates a new Metadata returning false if Main-Class is not defined.
func NewMetadata(application application.Application, logger logger.Logger) (Metadata, bool, error) {
	md := Metadata{}

	m, err := manifest.NewManifest(application, logger)
	if err != nil {
		return Metadata{}, false, err
	}

	if err := m.Decode(&md); err != nil {
		return Metadata{}, false, err
	}

	if md.MainClass == "" {
		return Metadata{}, false, nil
	}

	md.ClassPath = []string{application.Root}
	return md, true, nil
}
