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
//

package route

import (
	"fmt"
	"io"
)

// DefaultConfigLogger is a basic logger used in tests when a *testing.T type
// is not available (such as in Examples)
type DefaultConfigLogger struct {
	w io.Writer
}

// Infof does not print out to writer since doing so would modify the output in the Examples
func (l *DefaultConfigLogger) Infof(format string, v ...interface{}) {
	return
}

func (l *DefaultConfigLogger) Fatalf(format string, v ...interface{}) {
	fmt.Fprintf(l.w, format, v)
}

func (l *DefaultConfigLogger) Errorf(format string, v ...interface{}) {
	fmt.Fprintf(l.w, format, v)
}

// NewDefaultConfigLogger creates an instance of a DefaultConfigLogger
func NewDefaultConfigLogger(w io.Writer) *DefaultConfigLogger {
	return &DefaultConfigLogger{
		w: w,
	}
}
