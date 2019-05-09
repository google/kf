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

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type prefixFilter struct {
	nw newlineWriter
}

func NewPrefixFilter(prefixes map[string]io.Writer, defaultWriter io.Writer) io.Writer {
	return &prefixFilter{
		nw: newlineWriter{
			dest: func(data []byte) (int, error) {
				for prefix, dest := range prefixes {
					result := bytes.SplitAfterN(data, []byte(prefix), 2)
					if len(result) != 2 {
						continue
					}

					return fmt.Fprintln(dest, string(result[1]))
				}

				return fmt.Fprintln(defaultWriter, string(data))
			},
		},
	}
}

func (f *prefixFilter) Write(data []byte) (int, error) {
	return f.nw.Write(data)
}

type newlineWriter struct {
	dest func([]byte) (int, error)
	buf  bytes.Buffer
}

func (w *newlineWriter) Write(data []byte) (int, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		// Check to see if we ended on a delimiter. If not, we need to save it
		// in memory and wait for the delimiter.
		if err := r.UnreadByte(); err != nil {
			return 0, err
		}
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if b != '\n' {
			// Did not end in delimiter, add to buffer and move on.
			w.buf.Write(data)
			break
		}
		// End end delimiter checking.

		if _, err := w.dest(append(w.buf.Bytes(), line...)); err != nil {
			return 0, err
		}
		w.buf.Reset()
	}

	return len(data), nil
}
