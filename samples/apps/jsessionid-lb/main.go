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

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	// Indicates no backing services could be found.
	badGateway = "502 Bad Gateway -- no upstream hosts."
)

var (
	sessionCookie string
	stickyCookie  string
	proxyService  string
	proxyPort     string
	proxyScheme   string
	servingPort   string

	logger = log.New(os.Stderr, "proxy ", log.Lshortfile|log.Lmicroseconds|log.LUTC|log.Ldate)
)

func main() {
	rand.Seed(time.Now().Unix())

	servingPort = os.Getenv("PORT")
	sessionCookie = os.Getenv("SESSION_COOKIE")
	stickyCookie = os.Getenv("STICKY_COOKIE")
	proxyService = os.Getenv("PROXY_SERVICE")
	proxyPort = os.Getenv("PROXY_PORT")
	proxyScheme = os.Getenv("PROXY_SCHEME")

	if servingPort == "" {
		logger.Fatal("Environment variable PORT is required.")
		return
	}

	if proxyService == "" {
		logger.Fatal("Environment variable PROXY_SERVICE is required.")
		return
	}

	if proxyPort == "" {
		proxyPort = "8080"
	}

	if stickyCookie == "" {
		stickyCookie = "__VCAP_ID__"
	}

	if sessionCookie == "" {
		sessionCookie = "JSESSIONID"
	}

	if proxyScheme == "" {
		proxyScheme = "http"
	}

	logger.Println("Proxy Configuration")
	logger.Printf("  Session cookie: %q\n", sessionCookie)
	logger.Printf("  Sticky cookie: %q\n", stickyCookie)
	logger.Printf("  Proxy service: %s:%s\n", proxyService, proxyPort)

	proxy := &httputil.ReverseProxy{
		FlushInterval: 1 * time.Second,
		ErrorLog:      logger,
		Director:      func(_ *http.Request) {},
	}

	address := fmt.Sprintf(":%s", proxyPort)
	logger.Printf("Listening on %s\n", address)

	http.ListenAndServe(address, WrapProxy(proxy))
}

// responseAdatper
type responseAdapter struct {
	http.ResponseWriter

	proxiedAddr string
}

func (a *responseAdapter) WriteHeader(statusCode int) {
	// Append the sticky cookie the response has a session cookie set.

	for _, cookie := range (&http.Response{
		Header: a.ResponseWriter.Header(),
	}).Cookies() {
		if cookie.Name != sessionCookie {
			continue
		}

		// Copy the session cookie's properties for most values.
		http.SetCookie(
			a.ResponseWriter,
			&http.Cookie{
				Name:   stickyCookie,
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

func WrapProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s %s\n", r.Method, r.URL.Redacted())

		// Lookup the current set of backends using DNS service discovery.
		ips, err := net.LookupIP(proxyService)
		if err != nil {
			logger.Printf("Couldn't lookup backends: %v", err)
			http.Error(w, badGateway, http.StatusBadGateway)
			return
		}

		if len(ips) == 0 {
			logger.Println("No healthy backends")
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
		if c, err := r.Cookie(stickyCookie); err == nil {
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
			r.URL.Scheme = proxyScheme
		}
		r.URL.Host = fmt.Sprintf("%s:%s", destinationIP.String(), proxyPort)

		logger.Printf("- %d healthy backends, forwarding to %q\n", len(ips), r.URL.Redacted())

		rw := &responseAdapter{
			ResponseWriter: w,
			proxiedAddr:    destinationIP.String(),
		}

		// Forward on to the destination.
		next.ServeHTTP(rw, r)
	})
}
