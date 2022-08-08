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

	"github.com/google/kf/v2/samples/apps/jsessionid-lb/pkg"
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
	logger.Printf("  Proxy service: %s://%s:%s\n", proxyScheme, proxyService, proxyPort)

	proxy := &httputil.ReverseProxy{
		FlushInterval: 1 * time.Second,
		ErrorLog:      logger,
		Director:      func(_ *http.Request) {},
	}

	address := fmt.Sprintf(":%s", proxyPort)
	logger.Printf("Listening on %s\n", address)

	lb := &pkg.LoadBalancer{
		SessionCookie: sessionCookie,
		StickyCookie:  stickyCookie,
		ProxyService:  proxyService,
		ProxyPort:     proxyPort,
		ProxyScheme:   proxyScheme,

		LookupIP:     net.LookupIP,
		ReverseProxy: proxy,
		ErrorLog:     logger,
		RandIntn:     rand.Intn,
	}

	http.ListenAndServe(address, lb)
}
