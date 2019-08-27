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

// WrapperFunc wraps an http.RoundTripper when a new transport is created for a
// client, allowing per connection behavior to be injected.
type WrapperFunc func(http.RoundTripper) http.RoundTripper

// LoggingRoundTripperWrapper returns a WrapperFunc that logs values to stderr
// if params.LogHTTP is true.
func LoggingRoundTripperWrapper(params *KfParams) WrapperFunc {
	return func(in http.RoundTripper) http.RoundTripper {
		return NewLoggingRoundTripper(params, in)
	}
}

// NewLoggingRoundTripper creates a new logger that logs to stderr that wraps
// an inner RoundTripper.
func NewLoggingRoundTripper(params *KfParams, wrapped http.RoundTripper) http.RoundTripper {
	return NewLoggingRoundTripperWithStream(params, wrapped, os.Stderr)
}

// NewLoggingRoundTripperWithStream creates a new logger that logs to the given stream that wraps
// an inner RoundTripper.
func NewLoggingRoundTripperWithStream(params *KfParams, wrapped http.RoundTripper, out io.Writer) http.RoundTripper {
	return &LoggingRoundTripper{
		params: params,
		inner:  wrapped,
		out:    out,
	}
}

// LoggingRoundTripper logs HTTP requests.
type LoggingRoundTripper struct {
	params *KfParams
	inner  http.RoundTripper
	out    io.Writer
}

var _ http.RoundTripper = (*LoggingRoundTripper)(nil)

func isSensitiveHeader(header string) bool {
	sensitiveHeaderSet := map[string]bool{
		"Authorization":       true,
		"WWW-Authenticate":    true,
		"Cookie":              true,
		"Proxy-Authorization": true,
	}

	return sensitiveHeaderSet[header]
}

// RoundTrip implements http.RoundTripper.
func (t *LoggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if !t.params.LogHTTP {
		return t.inner.RoundTrip(r)
	}

	w := t.out

	reqCopy := sanitizeRequest(r)
	reqBytes, _ := httputil.DumpRequestOut(reqCopy, true)
	fmt.Fprintln(textio.NewPrefixWriter(w, "Request  > "), string(reqBytes))

	// Copy the body from the copy back over to the main request. The body is
	// consumed as part of httputil.DumpRequestOut but is replaced with an
	// identical stream from memory.
	r.Body = reqCopy.Body

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

func sanitizeRequest(r *http.Request) (out *http.Request) {
	reqCopy := *r
	reqCopy.Header = http.Header{}
	for k, v := range r.Header {
		if isSensitiveHeader(k) {
			v = []string{"[REDACTED]"}
		}

		reqCopy.Header[k] = v
	}

	return &reqCopy
}
