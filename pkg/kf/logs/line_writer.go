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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// LineWriter contains writer with a mutex instance. It is VERY opinionated on
// how it writes to the underlying writer. It counts lines and will STOP
// writing to the underlying writer if a limit is set.
type LineWriter struct {
	Writer io.Writer
	sync.Mutex

	// NumberOfLines is a VERY simple (hacky) way for us to ensure only the
	// set number of lines is written to stdout.
	NumberOfLines int
}

// Write implements io.Writer
func (mw *LineWriter) Write(data []byte) (int, error) {
	mw.Lock()
	defer mw.Unlock()

	totalLines := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		ll := parseLogLine(scanner.Text())

		fmt.Fprintln(mw.Writer, ll.text)
		totalLines++
		if mw.NumberOfLines != 0 && totalLines >= mw.NumberOfLines {
			break
		}
	}

	// We're going to lie about how much data we wrote. Either we changed the
	// newline, or we didn't actually write all the data. Either way, we're
	// going to suggest we wrote out the expected number of bytes.
	return len(data), scanner.Err()
}

// CopyFrom copies from s to Writer. It returns the latest timestamp written.
func (mw *LineWriter) CopyFrom(s io.Reader) (time.Time, error) {
	ltw := &latestTimeWriter{}
	if _, err := io.Copy(mw, io.TeeReader(s, ltw)); err != nil {
		if err == io.EOF {
			return ltw.latest, nil
		}

		return time.Time{}, err
	}

	return ltw.latest, nil
}

// latestTimeWriter simply keeps track of the latest timestamp. It never
// returns an error or even writes anything.
type latestTimeWriter struct {
	latest time.Time
}

// Write implements io.Writer
func (w *latestTimeWriter) Write(data []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		ll := parseLogLine(scanner.Text())

		// Replace the latest timestamp if this one is newer.
		if ll.timestamp.After(w.latest) {
			w.latest = ll.timestamp
		}
	}

	return len(data), nil
}
