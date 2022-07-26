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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

type recordingTransport struct {
	requestDump string
}

func (rt *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, fmt.Errorf("dumping request: %v", err)
	}
	rt.requestDump = string(dump)
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

type errorTransport struct{}

func (d *errorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("This is an error")
}

func TestLoggingRoundTripper_RoundTrip_normal(t *testing.T) {
	out := &bytes.Buffer{}
	recorder := &recordingTransport{}
	transport := NewLoggingRoundTripperWithStream(&KfParams{LogHTTP: true}, recorder, out)

	body := "{this: is, the: body, of: [the, request]}"
	req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(body))
	nonRedacted := "this string will be logged"
	redacted := "this string will be redacted"
	req.Header.Add("non-redacted", nonRedacted)
	req.Header.Add("Authorization", redacted)

	_, e := transport.RoundTrip(req)
	testutil.AssertNil(t, "RoundTrip error", e)
	s := out.String()

	testutil.AssertContainsAll(t, s, []string{
		"Request  >",
		"Response <",
		body,
		nonRedacted,
	})
	if strings.Contains(s, redacted) {
		t.Fatal("written output contained redacted fields")
	}

	// the redacted fields SHOULD be passed to the downstream RoundTripper
	testutil.AssertContainsAll(t, recorder.requestDump, []string{
		body,
		nonRedacted,
		redacted,
	})
}

func TestLoggingRoundTripper_RoundTrip_error(t *testing.T) {
	out := &bytes.Buffer{}
	transport := NewLoggingRoundTripperWithStream(&KfParams{LogHTTP: true}, &errorTransport{}, out)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	transport.RoundTrip(req)
	s := out.String()

	testutil.AssertContainsAll(t, s, []string{
		"Request  >",
		"Error    =",
	})
}

func TestLoggingRoundTripper_RoundTrip_noLogging(t *testing.T) {
	out := &bytes.Buffer{}
	transport := NewLoggingRoundTripperWithStream(&KfParams{LogHTTP: false}, &recordingTransport{}, out)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	transport.RoundTrip(req)
	s := out.String()

	testutil.AssertEqual(t, "RoundTrip log", "", s)
}
