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

package logs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	time "time"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestMutexWriter_CopyFrom(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	mw := &LineWriter{
		Writer: buf,
	}

	b := time.Now()
	a := b.Add(-time.Second)
	c := b.Add(time.Second)

	lastTime, err := mw.CopyFrom(ioutil.NopCloser(strings.NewReader(
		strings.Join(
			[]string{
				fmt.Sprintf("%s b", b.Format(time.RFC3339)),
				fmt.Sprintf("%s a", a.Format(time.RFC3339)),
				fmt.Sprintf("%s c", c.Format(time.RFC3339)),
			},
			"\n",
		),
	)))
	testutil.AssertErrorsEqual(t, nil, err)
	testutil.AssertEqual(t, "value", "b\na\nc\n", buf.String())
	testutil.AssertEqual(t, "lastTime", c.Unix(), lastTime.Unix())
}

func TestMutexWriter_cutoff(t *testing.T) {
	t.Parallel()

	// Assert that the MutexWriter stops writing to the underlying writer
	// after a certain number of lines have been written.
	buf := &bytes.Buffer{}
	mw := &LineWriter{
		Writer:        buf,
		NumberOfLines: 2,
	}
	mw.CopyFrom(ioutil.NopCloser(strings.NewReader("foo\nbar\nbaz\n")))

	testutil.AssertEqual(t, "value", "foo\nbar\n", buf.String())
}

func TestMutexWriter_race(t *testing.T) {
	t.Parallel()

	// This test doesn't have any direct assertions. Instead it lets the race
	// detector have an opportunity to look for problems and fail the test.

	mw := &LineWriter{
		Writer: &bytes.Buffer{},
	}

	go func() {
		for i := 0; i < 10000; i++ {
			mw.Write([]byte(fmt.Sprintf("%d", i)))
		}
	}()

	buf := &bytes.Buffer{}
	for i := 0; i < 10000; i++ {
		buf.Write([]byte(fmt.Sprintf("%d", i)))
	}
	mw.CopyFrom(ioutil.NopCloser(buf))
}
