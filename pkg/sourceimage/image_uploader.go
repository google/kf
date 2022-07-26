// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sourceimage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	v1alpha1lister "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/config"
	"knative.dev/pkg/logging"
)

// ImageUploader is the logic for the service API for data.builds.appdevexperience.dev.
// It composes a container image and publishes it for the given tar data.
type ImageUploader struct {
	spaceLister v1alpha1lister.SpaceLister
	buildLister v1alpha1lister.BuildLister
	imagePusher func(path, imageName string) (name.Reference, error)
	cfg         *config.Defaults
}

// NewImageUploader returns a new ImageUploader.
func NewImageUploader(
	cfg *config.Defaults,
	buildLister v1alpha1lister.BuildLister,
	spaceLister v1alpha1lister.SpaceLister,
	imagePusher func(path, imageName string) (name.Reference, error),
) *ImageUploader {
	return &ImageUploader{
		buildLister: buildLister,
		spaceLister: spaceLister,
		imagePusher: imagePusher,
		cfg:         cfg,
	}
}

// Upload composes a container image and publishes the given data. It will
// publish to the configured container registry.
func (u *ImageUploader) Upload(
	ctx context.Context,
	namespace string,
	name string,
	r io.Reader,
) (name.Reference, error) {
	// Fetch the Space to lookup the container registry.
	space, err := u.spaceLister.Get(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to find Space: %v", err)
	}
	// Fetch the Build to lookup the spec.
	build, err := u.
		buildLister.
		Builds(namespace).
		Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find Build: %v", err)
	}

	// Ensure the build is still pending. Otherwise, it is considered
	// immutable.
	if v1alpha1.IsStatusFinal(build.Status.Status) {
		return nil, errors.New("Build is not pending")
	}

	// Save the data to a temp file and take the checksum to ensure they
	// match.
	cleanup, tmpFile, err := u.save(ctx, r, build)

	// Always invoke the cleanup, even if there is an error.
	defer cleanup()

	if err != nil {
		return nil, err
	}

	imageURI := destinationImageName(build, space)
	// Push the image.
	imageRef, err := u.imagePusher(tmpFile, imageURI)
	if err != nil {
		return nil, fmt.Errorf("failed to build and push image: %v", err)
	}

	// Success!
	return imageRef, nil
}

func destinationImageName(source *v1alpha1.Build, space *v1alpha1.Space) string {
	registry := space.Status.BuildConfig.ContainerRegistry

	// Use underscores because those aren't permitted in k8s names so you can't
	// cause accidental conflicts.

	return path.Join(registry, fmt.Sprintf("app_%s_%s:%s", source.Namespace, source.Name, source.UID))
}

func (u *ImageUploader) save(
	ctx context.Context,
	r io.Reader,
	build *v1alpha1.Build,
) (cleanup func(), filePath string, err error) {
	logger := logging.FromContext(ctx)

	// Create a temp file to save the data to.
	cleanup, tmpFile, err := createTempFile(ctx)
	if err != nil {
		return cleanup, "", err
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			logger.Warnf("failed to close temp file for SourcePackge %s: %v", build.Name, err)
		}
	}()

	// Save the data to the file.
	if _, err := io.Copy(tmpFile, r); err != nil && err != io.EOF {
		return cleanup, "", fmt.Errorf("failed to save data: %v", err)
	}

	// Success!
	return cleanup, tmpFile.Name(), nil
}

func createTempFile(
	ctx context.Context,
) (cleanup func(), tmpFile *os.File, err error) {
	logger := logging.FromContext(ctx)

	// cleanup will be replaced when it has a file to delete.
	cleanup = func() {}

	// Create a temporary file to save the data to.
	tmpFile, err = ioutil.TempFile("", "")
	if err != nil {
		return cleanup, nil, fmt.Errorf("failed to create temp file: %v", err)
	}

	cleanup = func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			logger.Warnf("failed to delete temp file %q: %v", tmpFile.Name(), err)
		}
	}

	// Success!
	return cleanup, tmpFile, nil
}
