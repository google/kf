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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// RouteServiceURL is an alias for the net/url parsing of the service URL.
type RouteServiceURL URL

// URL is a copy of url.URL. We have to duplicate this so the schema
// generartion tools are happy.
type URL struct {
	Scheme      string `json:"Scheme,omitempty"`
	Opaque      string `json:"Opaque,omitempty"`
	Host        string `json:"Host,omitempty"`
	Path        string `json:"Path,omitempty"`
	RawPath     string `json:"RawPath,omitempty"`
	ForceQuery  bool   `json:"ForceQuery,omitempty"`
	RawQuery    string `json:"RawQuery,omitempty"`
	Fragment    string `json:"Fragment,omitempty"`
	RawFragment string `json:"RawFragment,omitempty"`
}

func (u URL) toURL() *url.URL {
	return &url.URL{
		Scheme:      u.Scheme,
		Opaque:      u.Opaque,
		Host:        u.Host,
		Path:        u.Path,
		RawPath:     u.RawPath,
		ForceQuery:  u.ForceQuery,
		RawQuery:    u.RawQuery,
		Fragment:    u.Fragment,
		RawFragment: u.RawFragment,
	}
}

// Hostname returns the Host of the URL, stripping any valid port number if present.
func (u *RouteServiceURL) Hostname() string {
	if u == nil {
		return ""
	}
	return URL(*u).toURL().Hostname()
}

// Port returns the port part of the Host, without the leading colon.
// If the host doesn't contain a valid numberic port, Port returns an empty string.
func (u *RouteServiceURL) Port() string {
	if u == nil {
		return ""
	}
	return URL(*u).toURL().Port()
}

// String converts a URL into a valid URL string, using the net/url implementation to construct the string.
func (u *RouteServiceURL) String() string {
	if u == nil {
		return ""
	}
	return URL(*u).toURL().String()
}

// URL converts the RouteServiceURL into a url.URL from net/url.
func (u *RouteServiceURL) URL() *url.URL {
	if u == nil {
		return &url.URL{}
	}
	return URL(*u).toURL()
}

// ParseURL converts a raw string into a RouteServiceURL. It follows the same logic as Parse in net/url.
func ParseURL(rawURL string) (*RouteServiceURL, error) {
	if rawURL == "" {
		return &RouteServiceURL{}, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &RouteServiceURL{
		Scheme:      parsed.Scheme,
		Opaque:      parsed.Opaque,
		Host:        parsed.Host,
		Path:        parsed.Path,
		RawPath:     parsed.RawPath,
		ForceQuery:  parsed.ForceQuery,
		RawQuery:    parsed.RawQuery,
		Fragment:    parsed.Fragment,
		RawFragment: parsed.RawFragment,
	}, nil
}

// MarshalJSON is a custom JSON marshal implementation for RouteServiceURL.
// It encodes the URL as a readable string.
func (u RouteServiceURL) MarshalJSON() ([]byte, error) {
	b := fmt.Sprintf("%q", u.String())
	return []byte(b), nil
}

// UnmarshalJSON is a custom JSON unmarshal implementation for RouteServiceURL.
// It decodes the JSON into a RouteServiceURL.
func (u *RouteServiceURL) UnmarshalJSON(b []byte) error {
	var urlStr string
	if err := json.Unmarshal(b, &urlStr); err != nil {
		return err
	}
	if rsURL, err := ParseURL(urlStr); err != nil {
		return err
	} else if rsURL != nil {
		*u = *rsURL
	} else {
		*u = RouteServiceURL{}
	}

	return nil
}
