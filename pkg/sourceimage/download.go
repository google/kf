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
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	v1alpha1lister "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/pkg/errors"
)

// Download returns the extracted image from the specified SourcePackage.
func Download(sourcePackageLister v1alpha1lister.SourcePackageLister, namespace, sourcePackageName string) (io.ReadCloser, error) {
	sourcePackage, err := sourcePackageLister.SourcePackages(namespace).Get(sourcePackageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get source package")
	}

	// Ensure the sourcePackage was successfully uploaded
	if !v1alpha1.IsStatusFinal(sourcePackage.Status.Status) {
		return nil, errors.New("source package not yet uploaded")
	}

	imageName := sourcePackage.Status.Image
	imageRef, err := name.ParseReference(imageName, name.WeakValidation)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse image ref %q", imageName))
	}
	image, err := remote.Image(imageRef, remote.WithAuthFromKeychain(Keychain()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image")
	}
	return mutate.Extract(image), nil
}
