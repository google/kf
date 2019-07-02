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

package utils_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestPrefixFilter(t *testing.T) {
	t.Parallel()

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	defaultWriter := &bytes.Buffer{}

	f := utils.NewPrefixFilter(map[string]io.Writer{
		"[prefix-1] ": buf1,
		"[prefix-2] ": buf2,
	},
		defaultWriter,
	)
	f.Write([]byte("[prefix-1] "))
	f.Write([]byte("data-0\n"))
	f.Write([]byte("[prefix-2] data-1\n"))
	f.Write([]byte("[other-prefix] [prefix-2] data-2\n[prefix-1] data-3\n"))
	f.Write([]byte("[other-prefix] data-4\n"))

	testutil.AssertEqual(t, "buf-1", "data-0\ndata-3\n", buf1.String())
	testutil.AssertEqual(t, "buf-2", "data-1\ndata-2\n", buf2.String())
	testutil.AssertEqual(t, "defaultWriter", "[other-prefix] data-4\n", defaultWriter.String())
}
