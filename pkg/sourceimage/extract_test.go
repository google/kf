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
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

type tarEntry struct {
	tar.Header
	contents []byte
}

func file(name, content string) tarEntry {
	return custom(name, tar.TypeReg, 0600, content)
}

func directory(name string) tarEntry {
	return custom(name, tar.TypeDir, 0755, "")
}

func custom(name string, typeFlag byte, mode int64, content string) tarEntry {
	contentBytes := []byte(content)
	return tarEntry{
		Header: tar.Header{
			Name:     name,
			Typeflag: typeFlag,
			Mode:     mode,
			Size:     int64(len(contentBytes)),
		},
		contents: contentBytes,
	}
}

func newTar(entries ...tarEntry) *bytes.Buffer {
	out := &bytes.Buffer{}
	w := tar.NewWriter(out)

	for _, entry := range entries {
		w.WriteHeader(&entry.Header)
		w.Write(entry.contents)
	}

	w.Flush()
	return out
}

func AssertFile(t *testing.T, base, file string, mode int64, contents []byte) {
	t.Helper()

	joined := filepath.Join(base, file)
	stat, err := os.Stat(joined)
	testutil.AssertNil(t, "stat err", err)
	testutil.AssertEqual(t, "mode", mode, int64(stat.Mode().Perm()))
	testutil.AssertEqual(t, "size", int64(len(contents)), stat.Size())

	actualContents, err := ioutil.ReadFile(joined)
	testutil.AssertNil(t, "read err", err)
	testutil.AssertEqual(t, "contents", contents, actualContents)
}

func TestExtractTar(t *testing.T) {
	tarBytes := newTar(
		directory("/home/foo"),
		file("/home/foo/readme.txt", "some-text"),
		custom("/home/foo/prog", tar.TypeReg, 0700, "ELF"),
		file("/home/foo/../../etc/passwd", "malicious"),
		file("home/foo/no-leading-slash", "no-leading-slash"),
		file("./home/foo/leading-dot", "leading-dot"),

		// NOTE: This file does not have a corresponding directory on purpose.
		// JARs occasionally have this behavior.
		file("./home/bar/no-dir", "no-dir"),
		directory("./home/bar"),
	)

	temp, err := ioutil.TempDir("", "")
	testutil.AssertNil(t, "temp dir err", err)
	defer os.RemoveAll(temp)

	r := tar.NewReader(tarBytes)
	tarErr := ExtractTar(temp, "/home", r)
	testutil.AssertNil(t, "tar err", tarErr)

	AssertFile(t, temp, "foo/readme.txt", 0600, []byte("some-text"))
	AssertFile(t, temp, "foo/prog", 0700, []byte("ELF"))
	AssertFile(t, temp, "foo/no-leading-slash", 0600, []byte("no-leading-slash"))
	AssertFile(t, temp, "foo/leading-dot", 0600, []byte("leading-dot"))
	AssertFile(t, temp, "bar/no-dir", 0600, []byte("no-dir"))
}
