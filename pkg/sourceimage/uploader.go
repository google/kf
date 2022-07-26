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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	v1alpha1lister "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"knative.dev/pkg/logging"
)

// Uploader is the logic for the service API for upload.kf.dev. It composes a
// container image and publishes it for the given data.
type Uploader struct {
	spaceLister         v1alpha1lister.SpaceLister
	sourcePackageLister v1alpha1lister.SourcePackageLister
	imagePusher         func(path, imageName string) (name.Reference, error)
	statusUpdater       func(s *v1alpha1.SourcePackage) error
}

// NewUploader returns a new Uploader.
func NewUploader(
	spaceLister v1alpha1lister.SpaceLister,
	sourcePackageLister v1alpha1lister.SourcePackageLister,
	imagePusher func(path, imageName string) (name.Reference, error),
	statusUpdater func(s *v1alpha1.SourcePackage) error,
) *Uploader {
	return &Uploader{
		spaceLister:         spaceLister,
		sourcePackageLister: sourcePackageLister,
		imagePusher:         imagePusher,
		statusUpdater:       statusUpdater,
	}
}

// Upload composes a container image and publishes the given data. It will
// publish to the configured container registry found in the Space.
func (u *Uploader) Upload(
	ctx context.Context,
	spaceName string,
	sourcePackageName string,
	r io.Reader,
) (name.Reference, error) {
	// Fetch the Space to read the configured container registry.
	space, err := u.spaceLister.Get(spaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find Space: %v", err)
	}

	// Fetch the SourcePackage to lookup the spec.
	sourcePackage, err := u.
		sourcePackageLister.
		SourcePackages(spaceName).
		Get(sourcePackageName)
	if err != nil {
		return nil, fmt.Errorf("failed to find SourcePackage: %v", err)
	}

	// Ensure the sourcePackage is still pending. Otherwise, it is considered
	// immutable.
	if v1alpha1.IsStatusFinal(sourcePackage.Status.Status) {
		return nil, errors.New("SourcePackage is not pending")
	}

	// Save the data to a temp file and take the checksum to ensure they
	// match.
	cleanup, tmpFile, err := u.saveAndVerify(ctx, r, sourcePackage)

	// Always invoke the cleanup, even if there is an error.
	defer cleanup()

	if err != nil {
		return nil, err
	}

	// Build the image name using configured container registry and
	// SourcePackage UID.
	imageName := path.Join(
		space.Status.BuildConfig.ContainerRegistry,
		string(sourcePackage.UID),
	)

	// Build and push the data to the configured container registry.
	imageRef, err := u.imagePusher(tmpFile, imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to build and push image: %v", err)
	}

	// Propagate the spec to the SourcePackage status.
	sourcePackage = sourcePackage.DeepCopy()
	sourcePackage.Status.PropagateSpec(imageRef.Name(), sourcePackage.Spec)

	// Save the SourcePackage status.
	if err := u.statusUpdater(sourcePackage); err != nil {
		return nil, fmt.Errorf("failed to update SourcePackage status: %v", err)
	}

	// Success!
	return imageRef, nil
}

// saveAndVerify reads data from the given reader and saves it to temp file.
// While saving the file, it also calcultes the checksum (based on the
// configured hasher). It then compares the hash and size to the expected ones
// set by the user. A mismatch results in an error.
//
// The resulting cleanup function MUST be invoked (even when there is an
// error). This is responsible for cleaning up the temp file.
func (u *Uploader) saveAndVerify(
	ctx context.Context,
	r io.Reader,
	sourcePackage *v1alpha1.SourcePackage,
) (cleanup func(), filePath string, err error) {
	logger := logging.FromContext(ctx)

	// Create a temp file to save the data to.
	cleanup, tmpFile, err := u.createTempFile(ctx, sourcePackage)
	if err != nil {
		return cleanup, "", err
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			logger.Warnf("failed to close temp file for SourcePackge %s: %v", sourcePackage.Name, err)
		}
	}()

	// Calculate the checksum as we write to the file.
	// NOTE: The only checksum we currently support is sha256.
	var hasher hash.Hash
	switch sourcePackage.Spec.Checksum.Type {
	case v1alpha1.PackageChecksumSHA256Type:
		hasher = sha256.New()
	default:
		// Unknown Checksum type.
		return cleanup, "", fmt.Errorf("unknown checksum type: %s", sourcePackage.Spec.Checksum.Type)
	}
	tr := io.TeeReader(r, hasher)

	// Save the data to the file.
	if n, err := io.CopyN(tmpFile, tr, int64(sourcePackage.Spec.Size)); err != nil && err != io.EOF {
		return cleanup, "", fmt.Errorf("failed to save data: %v", err)
	} else if expected, actual := sourcePackage.Spec.Size, uint64(n); expected != actual {
		return cleanup, "", fmt.Errorf("expected %d bytes, got %d", expected, actual)
	}

	// Go requires an addressable space before slicing it (turning it into a
	// byte slice from an array). So save it to a variable.
	cz := hasher.Sum(nil)

	// Assert that the calculated checksum matches what was set in the spec.
	if actual, expected := hex.EncodeToString(cz[:]), sourcePackage.Spec.Checksum.Value; expected != actual {
		return cleanup, "", errors.New("checksum does not match expected")
	}

	// Success!
	return cleanup, tmpFile.Name(), nil
}

func (u *Uploader) createTempFile(
	ctx context.Context,
	sourcePackage *v1alpha1.SourcePackage,
) (cleanup func(), tmpFile *os.File, err error) {
	logger := logging.FromContext(ctx)

	// cleanup will be replaced when it has a file to delete.
	cleanup = func() {}

	// Create a temporary file to save the data to.
	tmpFile, err = ioutil.TempFile("", string(sourcePackage.UID))
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
