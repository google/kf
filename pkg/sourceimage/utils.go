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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"
)

const (
	// DefaultSourcePath contains the path to the source code inside the image.
	DefaultSourcePath = "/var/run/kontext"
)

// IsSubPath returns whether or not the path is a child of the given path.
func IsSubPath(path, prefix string) bool {
	return TrimPathPrefix(path, prefix) != path
}

// TrimPathPrefix removes the full prefix from the given path if it is a sub
// path.
func TrimPathPrefix(path, prefix string) string {
	pathParts := strings.Split(filepath.ToSlash(path), "/")
	prefixParts := strings.Split(filepath.ToSlash(prefix), "/")

	if len(pathParts) < len(prefixParts) {
		return path
	}

	if reflect.DeepEqual(prefixParts, pathParts[:len(prefixParts)]) {
		return filepath.Join(pathParts[len(prefixParts):]...)
	}

	return path
}

// ListTar writes a ls -la style listing of the tar to the given writer.
func ListTar(w io.Writer, reader *tar.Reader) error {
	sw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.StripEscape)
	defer sw.Flush()

	fmt.Fprintln(sw, "MODE\tSIZE\tNAME")
	for {
		header, err := reader.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		default:
			fmt.Fprintf(sw, "%s\t%d\t%s\n", os.FileMode(header.Mode), header.Size, header.Name)
		}
	}
}
