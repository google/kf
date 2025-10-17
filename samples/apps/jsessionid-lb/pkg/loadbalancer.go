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

package pkg

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
)

const (
	// Indicates no backing services could be found.
	badGateway = "502 Bad Gateway -- no upstream hosts."
)

// responseAdatper ensures the sticky cookie is added before the HTTP
// response is written if the session cookie is added.
type responseAdapter struct {
	http.ResponseWriter

	proxiedAddr   string
	sessionCookie string
	stickyCookie  string
}

// Unwrap implements http.ResponseWriter to allow the immediate flushing for HTTP Streaming events
func (a *responseAdapter) Unwrap() http.ResponseWriter {
	return a.ResponseWriter
}

func (a *responseAdapter) WriteHeader(statusCode int) {
	// Append the sticky cookie the response has a session cookie set.

	for _, cookie := range (&http.Response{
		Header: a.ResponseWriter.Header(),
	}).Cookies() {
		if cookie.Name != a.sessionCookie {
			continue
		}

		// Copy the session cookie's properties for most values.
		http.SetCookie(
			a.ResponseWriter,
			&http.Cookie{
				Name:   a.stickyCookie,
				Value:  a.proxiedAddr,
				Path:   cookie.Path,
				MaxAge: cookie.MaxAge,
				// Match gorouter HttpOnly:
				// https://github.com/cloudfoundry/gorouter/blob/379860daa83a162ffe0b6039eafb7c8bfa1eaccf/proxy/round_tripper/proxy_round_tripper.go#L345
				HttpOnly: true,
				Secure:   cookie.Secure,
				SameSite: cookie.SameSite,
				Expires:  cookie.Expires,
			},
		)

		break
	}

	a.ResponseWriter.WriteHeader(statusCode)
}

type LoadBalancer struct {
	// SessionCookie holds the name of the cookie set by the proxied app
	// to indicate the start of a session.
	SessionCookie string
	// StickyCookie is the name of the cookie used to store the sticky session.
	StickyCookie string

	// Proxied application settings.
	ProxyService string
	ProxyPort    string
	ProxyScheme  string

	// ReverseProxy holds an httputil.ReverseProxy to forward requests.
	ReverseProxy http.Handler
	// LookupIP matches net.LookupIP.
	LookupIP func(string) ([]net.IP, error)
	// Logger to write errors to.
	ErrorLog *log.Logger
	// Gets a random number [0, n), matches rand.Intn
	RandIntn func(int) int
}

// ServeHTTP implements http.Handler.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.ErrorLog.Printf("%s %s\n", r.Method, r.URL.Redacted())

	// Lookup the current set of backends using DNS service discovery.
	ips, err := lb.LookupIP(lb.ProxyService)
	if err != nil {
		lb.ErrorLog.Printf("Couldn't lookup backends: %v", err)
		http.Error(w, badGateway, http.StatusBadGateway)
		return
	}

	if len(ips) == 0 {
		lb.ErrorLog.Println("No healthy backends")
		http.Error(w, badGateway, http.StatusBadGateway)
		return
	}

	// Pick random destination IP.
	destinationIP := ips[rand.Intn(len(ips))]

	// If the inbound request has a cookie indicating a different IP,
	// check to make sure it's still a valid back-end to prevent us from
	// becoming an open proxy.
	//
	// If none match, the previous random IP is used.
	if c, err := r.Cookie(lb.StickyCookie); err == nil {
		for _, ip := range ips {
			// If there's a match, forward the request.
			if c.Value == ip.String() {
				destinationIP = ip
				break
			}
		}
	}

	// Update the destination address, but not the HTTP Host header
	// to make the proxy transparent.
	if r.URL.Scheme == "" {
		r.URL.Scheme = lb.ProxyScheme
	}
	r.URL.Host = fmt.Sprintf("%s:%s", destinationIP.String(), lb.ProxyPort)

	lb.ErrorLog.Printf("- %d healthy backends, forwarding to %q\n", len(ips), r.URL.Redacted())

	rw := &responseAdapter{
		ResponseWriter: w,
		proxiedAddr:    destinationIP.String(),
		stickyCookie:   lb.StickyCookie,
		sessionCookie:  lb.SessionCookie,
	}

	// Forward on to the destination.
	lb.ReverseProxy.ServeHTTP(rw, r)
}
