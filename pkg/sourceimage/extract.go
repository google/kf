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

package sourceimage

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

// ExtractImage extracts the files and directories from the given sourcePath in
// image to targetPath on the current system.
func ExtractImage(targetPath, sourcePath string, image v1.Image) error {
	tarReader := mutate.Extract(image)
	defer tarReader.Close()
	r := tar.NewReader(tarReader)
	return ExtractTar(targetPath, sourcePath, r)
}

// ExtractTar extracts the files and directories from the given sourcePath in
// the tarReader to the targetPath on the current system.
//
// Special files are skipped, and all files are converted to their absolute
// path equivalents to avoid traversal attacks.
func ExtractTar(targetPath, sourcePath string, tarReader *tar.Reader) error {
	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		default:
			if err := extract(targetPath, sourcePath, header, tarReader); err != nil {
				return err
			}
		}
	}
}

func copy(dest string, from io.Reader, info os.FileInfo) error {
	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, info.Mode())
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}

func extract(target, sourcePath string, header *tar.Header, r io.Reader) error {
	// certain archive creators skip leading slashses or use ./ as the prefix
	// for image contents in the root directory.
	absPath := cleanTarPath(header.Name)
	if !IsSubPath(absPath, sourcePath) {
		// Skip non-matching files
		return nil
	}

	outPath := filepath.Join(target, TrimPathPrefix(absPath, sourcePath))

	switch header.Typeflag {
	case tar.TypeDir:
		return ignoreExistsErr(os.MkdirAll(outPath, header.FileInfo().Mode()))
	case tar.TypeReg:
		// NOTE: Some archives don't put a corresponding directory type ahead
		// of the file. Therefore, we have to ensure the directory already
		// exists.
		ignoreExistsErr(os.MkdirAll(filepath.Dir(outPath), 0755))

		return copy(outPath, r, header.FileInfo())
	default:
		log.Println("skipping unsupported type:", header.Typeflag, "for file:", absPath)
		return nil
	}
}

func ignoreExistsErr(err error) error {
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func cleanTarPath(tarPath string) string {
	return filepath.Join(string(filepath.Separator), filepath.Clean(tarPath))
}
