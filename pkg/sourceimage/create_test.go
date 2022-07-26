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
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestPushImage(t *testing.T) {
	t.Parallel()

	// Set up logger so we don't pollute stdout
	buf := &bytes.Buffer{}
	reg := registry.New(registry.Logger(log.New(buf, "", 0)))
	s := httptest.NewServer(reg)
	defer s.Close()

	r := strings.TrimPrefix(s.URL, "http://") + "/some-image"
	tag, err := name.NewTag(r)
	testutil.AssertNil(t, "tag err", err)

	image, err := random.Image(1024, 5)
	testutil.AssertNil(t, "random.Image err", err)

	reference, err := PushImage(tag.String(), image, false)
	testutil.AssertNil(t, "PushImage err", err)

	// Assert the reference is a digest
	_, err = name.NewDigest(reference.String())
	testutil.AssertNil(t, "name.NewDigest err", err)

	// Print the HTTP info if the test fails
	t.Log(buf.String())
}

func TestPushImage_retry(t *testing.T) {
	t.Parallel()

	// Set up logger so we don't pollute stdout
	buf := &bytes.Buffer{}
	reg := registry.New(registry.Logger(log.New(buf, "", 0)))

	// Wrap the registry handler so that it first returns errors and asserts
	// that PushImage retries.
	var count int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt64(&count, 1) < 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		reg.ServeHTTP(w, r)
	}))
	defer s.Close()

	r := strings.TrimPrefix(s.URL, "http://") + "/some-image"
	tag, err := name.NewTag(r)
	testutil.AssertNil(t, "tag err", err)

	image, err := random.Image(1024, 5)
	testutil.AssertNil(t, "random.Image err", err)

	_, err = PushImage(tag.String(), image, true)
	testutil.AssertNil(t, "PushImage err", err)

	// Print the HTTP info if the test fails
	t.Log(buf.String())
}

func TestPushImage_retryFails(t *testing.T) {
	t.Parallel()

	var count int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer s.Close()

	r := strings.TrimPrefix(s.URL, "http://") + "/some-image"
	tag, err := name.NewTag(r)
	testutil.AssertNil(t, "tag err", err)

	image, err := random.Image(1024, 5)
	testutil.AssertNil(t, "random.Image err", err)

	_, err = PushImage(tag.String(), image, true)
	testutil.AssertNotNil(t, "PushImage err", err)

	testutil.AssertEqual(t, "retry count", int64(120), atomic.LoadInt64(&count))
}

func TestPackageSourceTar(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		path   string
		filter FileFilter

		expectErr error
	}{
		"include everything": {
			path: filepath.Join("testdata", "source"),
			filter: func(_ string) bool {
				return true
			},
		},
		"exclude everything": {
			path: filepath.Join("testdata", "source"),
			filter: func(_ string) bool {
				return false
			},
		},
		"partial exclude": {
			path: filepath.Join("testdata", "source"),
			filter: func(path string) bool {
				return !strings.Contains(path, "exclude")
			},
		},
		"package a non zip file": {
			path: filepath.Join("testdata", "source", "include"),
			filter: func(_ string) bool {
				return true
			},
			expectErr: errors.New("testdata/source/include is not a zip archive"),
		},
		"package a zip file": {
			path: filepath.Join("testdata", "include.zip"),
			filter: func(_ string) bool {
				return true
			},
		},
		"package a jar file": {
			path: filepath.Join("testdata", "include.jar"),
			filter: func(_ string) bool {
				return true
			},
		},
		"file does not exist": {
			path: filepath.Join("testdata", "dne"),
			filter: func(_ string) bool {
				return true
			},
			expectErr: fmt.Errorf("stat %s: no such file or directory", filepath.Join("testdata", "dne")),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := PackageSourceTar(buf, tc.path, tc.filter)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)

			listBuf := &bytes.Buffer{}
			tarReader := tar.NewReader(buf)
			ListTar(listBuf, tarReader)
			testutil.AssertGolden(t, "tar-list", listBuf.Bytes())
		})
	}
}

func TestPackageFile(t *testing.T) {
	image, err := PackageFile(filepath.Join("testdata", "include.jar"))
	testutil.AssertNil(t, "PackageFile err", err)

	layers, err := image.Layers()
	testutil.AssertNil(t, "image.Layers err", err)
	testutil.AssertEqual(t, "layer count", 1, len(layers))
}

func TestPackageSourceDirectory(t *testing.T) {
	image, err := PackageSourceDirectory(filepath.Join("testdata", "source"), func(_ string) bool {
		return true
	})
	testutil.AssertNil(t, "PackageSourceDirectory err", err)

	layers, err := image.Layers()
	testutil.AssertNil(t, "image.Layers err", err)
	testutil.AssertEqual(t, "layer count", 1, len(layers))
}

func TestBuildIgnoreFilter(t *testing.T) {

	// The ignore files in the testdata directories ignore .o files.
	cases := map[string]struct {
		srcPath       string
		selectedFiles []string
		ignoredFiles  []string
		expectErr     error
	}{
		"cfignore": {
			srcPath:       "testdata/dockerfile-app",
			selectedFiles: []string{"Dockerfile"},
			ignoredFiles:  []string{"garbage.o"},
		},
		"kfignore": {
			srcPath:      "testdata/example-app",
			ignoredFiles: []string{"garbage.o"},
		},
		"cfignore itself": {
			srcPath:       "testdata/dockerfile-app",
			selectedFiles: []string{"Dockerfile"},
			ignoredFiles:  []string{".cfignore"},
		},
		"kfignore itself": {
			srcPath:      "testdata/example-app",
			ignoredFiles: []string{".kfignore"},
		},
		"default ignores": {
			srcPath:      "testdata/example-app",
			ignoredFiles: []string{"manifest.yml"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualFilter, err := BuildIgnoreFilter(tc.srcPath)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)

			for _, file := range tc.selectedFiles {
				result := actualFilter(file)
				testutil.AssertTrue(t, fmt.Sprint("result ", file), result)
			}

			for _, file := range tc.ignoredFiles {
				result := actualFilter(file)
				testutil.AssertFalse(t, fmt.Sprint("result ", file), result)
			}

			if tc.expectErr != nil {
				return
			}

			// Ensure the the filter works as expected with the tar packaging system
			buf := &bytes.Buffer{}
			err = PackageSourceTar(buf, tc.srcPath, actualFilter)
			testutil.AssertNil(t, "PackageSourceTar error", err)

			listBuf := &bytes.Buffer{}
			tarReader := tar.NewReader(buf)
			err = ListTar(listBuf, tarReader)
			testutil.AssertNil(t, "ListTar error", err)

			testutil.AssertGolden(t, "uploaded files", listBuf.Bytes())
		})
	}
}
