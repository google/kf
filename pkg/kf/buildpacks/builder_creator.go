package buildpacks

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/buildpack/pack"
)

// BuilderCreator creates a new buildback builder. It should be created via
// NewCreateBuilder.
type BuilderCreator struct {
	f BuilderFactoryCreate
}

// BuilderFactory creates and publishes new builde image.
type BuilderFactoryCreate func(flags pack.CreateBuilderFlags) error

// NewBuilderCreator creates a new BuilderCreator.
func NewBuilderCreator(f BuilderFactoryCreate) *BuilderCreator {
	return &BuilderCreator{
		f: f,
	}
}

// Create creates and publishes a builder image.
func (b *BuilderCreator) Create(dir, containerRegistry string) (string, error) {
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
