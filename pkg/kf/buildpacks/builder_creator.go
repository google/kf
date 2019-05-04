// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildpacks

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/buildpack/pack"
)

// BuilderCreator creates a new buildback builder.
type BuilderCreator interface {
	// Create creates and publishes a builder image.
	Create(dir, containerRegistry string) (string, error)
}

// builderCreator creates a new buildback builder. It should be created via
// NewCreateBuilder.
type builderCreator struct {
	f BuilderFactoryCreate
}

// BuilderFactory creates and publishes new builde image.
type BuilderFactoryCreate func(flags pack.CreateBuilderFlags) error

// NewBuilderCreator creates a new BuilderCreator.
func NewBuilderCreator(f BuilderFactoryCreate) BuilderCreator {
	return &builderCreator{
		f: f,
	}
}

// Create creates and publishes a builder image.
func (b *builderCreator) Create(dir, containerRegistry string) (string, error) {
	if dir == "" {
		return "", kf.ConfigErr{"dir must not be empty"}
	}
	if containerRegistry == "" {
		return "", kf.ConfigErr{"containerRegistry must not be empty"}
	}

	if filepath.Base(dir) != "builder.toml" {
		dir = filepath.Join(dir, "builder.toml")
	}

	imageName := path.Join(containerRegistry, fmt.Sprintf("buildpack-builder:%d", time.Now().UnixNano()))
	if err := b.f(pack.CreateBuilderFlags{
		Publish:         true,
		BuilderTomlPath: dir,
		RepoName:        imageName,
	}); err != nil {
		return "", err
	}

	return imageName, nil
}
