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

package describe

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/segmentio/textio"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// TabbedWriter indents all tabbed output to be aligned.
func TabbedWriter(w io.Writer, f func(io.Writer)) {
	out := new(tabwriter.Writer)
	out.Init(w, 0, 8, 2, ' ', 0)

	f(out)

	out.Flush()
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// IndentWriter creates a new writer that indents all lines passing through it
// by two spaces.
func IndentWriter(w io.Writer, f func(io.Writer)) {
	iw := textio.NewPrefixWriter(w, "  ")
	defer iw.Flush()

	f(iw)
}

// SectionWriter writes a section heading with the given name then calls f with
// a tab aligning indenting writer to format the contents of the section.
func SectionWriter(w io.Writer, name string, f func(io.Writer)) {
	buf := &bytes.Buffer{}

	TabbedWriter(buf, func(w io.Writer) {
		IndentWriter(w, f)
	})

	if len(buf.Bytes()) == 0 {
		fmt.Fprintf(w, "%s: <empty>\n", name)
	} else {
		fmt.Fprintf(w, "%s:\n", name)
		w.Write(buf.Bytes())
	}
}
