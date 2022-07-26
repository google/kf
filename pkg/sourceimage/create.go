// Copyright 2020 Google LLC
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

package sourceimage

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	ignore "github.com/sabhiram/go-gitignore"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// FileFilter returns true if the file should be included in the result and
// false otherwise.
type FileFilter func(path string) bool

// Keychain returns a Keychain that is used to push an image to a remote
// registry. It will first attempt to use the local docker configuration
// (found in $HOME/.docker/config.json). If that fails, it will defer to the
// Google Keychain which uses gcloud to fetch a token (used for GCR and AR).
func Keychain() authn.Keychain {
	return authn.NewMultiKeychain(
		// The auth ordering determines how the CLI/server will attempt to
		// fetch credentials.
		authn.DefaultKeychain,
		google.Keychain,
	)
}

// PushImage uploads the given v1.Image to a registry incorporating the
// provided string into the image's repository name.  Returns the digest
// of the published image.
func PushImage(imageName string, image v1.Image, retryOnErr bool) (name.Reference, error) {
	reference, err := name.NewTag(imageName, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("invalid image name %q: %v", imageName, err)
	}

	if err := retry.OnError(
		wait.Backoff{
			Steps:    120,
			Duration: 1 * time.Second,
			Factor:   1,
			Jitter:   0.1,
		},
		func(err error) bool {
			return retryOnErr
		},
		func() error {
			return remote.Write(
				reference,
				image,
				remote.WithAuthFromKeychain(Keychain()),
			)
		},
	); err != nil {
		return nil, fmt.Errorf("error publishing image: %v", err)
	}

	hash, err := image.Digest()
	if err != nil {
		return nil, err
	}

	return name.NewDigest(fmt.Sprintf("%s@%s", reference.Repository, hash))
}

// PackageFile converts the tar file into a container image.
func PackageFile(path string) (v1.Image, error) {
	dataLayer, err := tarball.LayerFromFile(path)
	if err != nil {
		return nil, err
	}

	image, err := mutate.AppendLayers(empty.Image, dataLayer)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// PackageSourceDirectory converts all the source code in sourcePath into a
// container image.
func PackageSourceDirectory(sourcePath string, filter FileFilter) (v1.Image, error) {
	buffer := &bytes.Buffer{}
	if err := PackageSourceTar(buffer, sourcePath, filter); err != nil {
		return nil, err
	}

	dataLayer, err := tarball.LayerFromReader(buffer)
	if err != nil {
		return nil, err
	}

	image, err := mutate.AppendLayers(empty.Image, dataLayer)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func isZippedFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// http.DetectContentType only considers up to 512 bytes:
	// http://golang.org/pkg/net/http/#DetectContentType
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return false, err
	}

	return http.DetectContentType(buff) == "application/zip", nil
}

// PackageSourceTar recurses through sourcePath and writes source files to dest
// that match the given filter.
//
// NOTE: file permissions in the created TAR are fixed to 0755 to prevent
// issues in the build system or Windows and files in the TAR will ALWAYS have
// forward-slashed paths.
func PackageSourceTar(dest io.Writer, sourcePath string, filter FileFilter) error {
	stat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	// If the value is a file, we want to replace the filter with one that only
	// accepts the original file specified. This will ignore additional filters
	// assuming the user knows they want the file if it's explicitly specified.

	// If the value is a file, it must be a compressed zip archive.
	// In this case, the values of the archive are passed through.
	if !stat.IsDir() {
		return packageZipToTar(dest, sourcePath)
	}

	return packageDirectoryToTar(dest, sourcePath, filter)
}

func packageZipToTar(dest io.Writer, sourcePath string) error {
	tarWriter := tar.NewWriter(dest)
	defer tarWriter.Close()

	isZipped, err := isZippedFile(sourcePath)
	if err != nil {
		return err
	}
	if !isZipped {
		return fmt.Errorf("%s is not a zip archive", sourcePath)
	}

	r, err := zip.OpenReader(sourcePath)
	if err != nil {
		return nil
	}
	defer r.Close()

	for _, f := range r.File {
		if err := func() error {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			if err := addFileToTar(tarWriter, f.Name, f.FileInfo().IsDir(), rc, f.FileInfo().Size()); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

func packageDirectoryToTar(dest io.Writer, sourcePath string, filter FileFilter) error {
	if filter == nil {
		return errors.New("filter must not be nil")
	}

	tarWriter := tar.NewWriter(dest)
	defer tarWriter.Close()

	return filepath.Walk(sourcePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %s", err)
		}

		// Chase symlinks.
		info, err = os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("stat error: %s", err)
		}

		relativePath := filepath.ToSlash(TrimPathPrefix(filePath, sourcePath))
		if !filter(relativePath) {
			return nil
		}

		fd, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("couldn't open %q: %s", filePath, err)
		}
		defer fd.Close()

		return addFileToTar(tarWriter, relativePath, info.Mode().IsDir(), fd, info.Size())
	})
}

func addFileToTar(tarWriter *tar.Writer, filePath string, isDir bool, reader io.Reader, size int64) error {

	tarPath := path.Join(DefaultSourcePath, filePath)

	if isDir {
		return tarWriter.WriteHeader(&tar.Header{
			Name:     tarPath, // Directories get trailing slashes in tars
			Typeflag: tar.TypeDir,
			Mode:     0755, // full permissions for owner, list/traverse for others
		})
	}

	if err := tarWriter.WriteHeader(&tar.Header{
		Name:     tarPath,
		Typeflag: tar.TypeReg,
		Mode:     0755, // full permissions for owner, read/execute for others
		Size:     size,
	}); err != nil {
		return fmt.Errorf("couldn't write tar header: %s", err)
	}

	if _, err := io.Copy(tarWriter, reader); err != nil {
		return fmt.Errorf("couldn't copy %q: %s", filePath, err)
	}

	return nil
}

// BuildIgnoreFilter returns a FileFilter that ignores files similar to CF.
func BuildIgnoreFilter(srcPath string) (FileFilter, error) {
	ignoreFiles := []string{
		".kfignore",
		".cfignore",
	}

	var defaultIgnoreLines = []string{
		".cfignore",
		".kfignore",
		"/manifest.yml",
		".gitignore",
		".git",
		".hg",
		".svn",
		"_darcs",
		".DS_Store",
	}

	var (
		gitignore *ignore.GitIgnore
		err       error
	)
	for _, ignoreFile := range ignoreFiles {
		gitignore, err = ignore.CompileIgnoreFileAndLines(
			filepath.Join(srcPath, ignoreFile),
			defaultIgnoreLines...,
		)
		if err != nil {
			// Just move on.
			continue
		}

		break
	}

	if gitignore == nil {
		gitignore, err = ignore.CompileIgnoreLines(defaultIgnoreLines...)
		if err != nil {
			return nil, err
		}
	}

	return func(path string) bool {
		return !gitignore.MatchesPath(path)
	}, nil
}
