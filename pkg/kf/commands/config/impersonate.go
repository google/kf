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

package config

import (
	"net/http"

	"k8s.io/client-go/transport"
)

// impersonatingRoundTripper adds impersonation headers to HTTP requests based
// on KfParams.
type impersonatingRoundTripper struct {
	params *KfParams
	inner  http.RoundTripper
}

var _ http.RoundTripper = (*impersonatingRoundTripper)(nil)

// NewImpersonatingRoundTripperWrapper returns a WrapperFunc that wraps a
// transport with an optional RoundTripper that adds impersonation headers
func NewImpersonatingRoundTripperWrapper(params *KfParams) transport.WrapperFunc {
	return func(in http.RoundTripper) http.RoundTripper {
		return &impersonatingRoundTripper{params, in}
	}
}

// RoundTrip implements http.RoundTripper.
func (rt *impersonatingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	delegate := rt.inner
	impersonate := rt.params.Impersonate

	if len(impersonate.UserName) > 0 ||
		len(impersonate.Groups) > 0 ||
		len(impersonate.Extra) > 0 {

		delegate = transport.NewImpersonatingRoundTripper(impersonate, delegate)
	}

	return delegate.RoundTrip(req)
}
