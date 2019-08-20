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

package config

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/segmentio/textio"
)

// NewLoggingRoundTripper creates a new logger that logs to stderr that wraps
// an inner RoundTripper.
func NewLoggingRoundTripper(wrapped http.RoundTripper) http.RoundTripper {
	return &LoggingRoundTripper{
		inner: wrapped,
		out:   os.Stderr,
	}
}

// LoggingRoundTripper logs HTTP requests.
type LoggingRoundTripper struct {
	inner http.RoundTripper
	out   io.Writer
}

var _ http.RoundTripper = (*LoggingRoundTripper)(nil)

func isSensitiveHeader(header string) bool {
	return (map[string]bool{
		"Authorization":       true,
		"WWW-Authenticate":    true,
		"Cookie":              true,
		"Proxy-Authorization": true,
	})[header]
}

// RoundTrip implements http.RoundTripper.
func (t *LoggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	w := t.out

	// Sanitize the request
	reqCopy := *r
	for k := range r.Header {
		if isSensitiveHeader(k) {
			reqCopy.Header.Set(k, "[REDACTED]")
		}
	}

	reqBytes, _ := httputil.DumpRequestOut(&reqCopy, true)
	fmt.Fprintln(textio.NewPrefixWriter(w, "Request  > "), string(reqBytes))

	// Make request & report output
	resp, err := t.inner.RoundTrip(r)
	if err != nil {
		fmt.Fprintln(textio.NewPrefixWriter(w, "ERROR    = "), err)
	} else {
		respBytes, _ := httputil.DumpResponse(resp, true)
		fmt.Fprintln(textio.NewPrefixWriter(w, "Response < "), string(respBytes))
	}

	return resp, err
}
