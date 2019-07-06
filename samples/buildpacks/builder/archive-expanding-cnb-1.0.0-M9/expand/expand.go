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

package expand

import (
	"fmt"
	"os"
	"strings"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

const (
	// Archive indicates the archive that should be expanded.
	Archive = "archive"

	// Dependency indicates that an application is an archive to be expanded.
	Dependency = "archive-expanding"
)

// Expand represents the information about an archive to expand.
type Expand struct {
	application application.Application
	archive     string
	logger      logger.Logger
}

// Contribute makes the contribution to the application.
func (e Expand) Contribute() error {
	e.logger.Body("Expanding %s to %s", e.archive, e.application.Root)

	switch {
	case strings.HasSuffix(e.archive, ".jar"),
		strings.HasSuffix(e.archive, ".war"),
		strings.HasSuffix(e.archive, ".zip"):

		if err := helper.ExtractZip(e.archive, e.application.Root, 0); err != nil {
			return err
		}
	case strings.HasSuffix(e.archive, ".tar.gz"),
		strings.HasSuffix(e.archive, ".tgz"):

		if err := helper.ExtractTarGz(e.archive, e.application.Root, 0); err != nil {
			return err
		}
	case strings.HasSuffix(e.archive, ".tar"):

		if err := helper.ExtractTar(e.archive, e.application.Root, 0); err != nil {
			return err
		}
	}

	e.logger.Body("Removing  %s", e.archive)
	return os.Remove(e.archive)
}

// NewExpand creates a new Expand instance.
func NewExpand(build build.Build) (Expand, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return Expand{}, false, nil
	}

	a, ok := bp.Metadata[Archive].(string)
	if !ok {
		return Expand{}, false, fmt.Errorf("unable to determine archive path")
	}

	return Expand{
		build.Application,
		a,
		build.Logger,
	}, true, nil
}
