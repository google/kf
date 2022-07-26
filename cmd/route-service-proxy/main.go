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
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/google/kf/v2/pkg/reconciler/route/resources"
)

func main() {
	rsURLString := os.Getenv("ROUTE_SERVICE_URL")
	if rsURLString == "" {
		log.Fatal("Route service URL is not set")
	}

	routeServiceURL, err := url.Parse(rsURLString)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Add scheme to URL if it does not exist. Defaults to HTTP.
	// Regenerate and update the URL to have the correct Host. (A URL with an empty scheme has an empty Host.)
	if routeServiceURL.Scheme == "" {
		err := updateURLParts(routeServiceURL)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	log.Fatal(http.ListenAndServe(hostPort(), newProxy(routeServiceURL)))
}

func hostPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf(":%s", port)
}

// updateURLParts modifies a URL in place and returns an error if parsing the updated URL fails.
// It adds the HTTP scheme to the URL if it does not exist and re-parses the URL string to have the correct Host.
// This ensures that the URL is correctly set in the reverse proxy.
func updateURLParts(existingURL *url.URL) error {
	existingURL.Scheme = "http"
	updatedURL, err := url.Parse(existingURL.String())
	if err != nil {
		return err
	}
	*existingURL = *updatedURL
	return nil
}

// newProxy forwards the original request to the route service URL.
// The request is modified with an added X-CF-Forwarded-URL header.
func newProxy(url *url.URL) http.Handler {
	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Set X-CF-Forwarded-URL header to original route destination URL.
			// This header is set so that the route service can forward the request to the original destination.
			forwardedURL := *req.URL
			forwardedURL.Scheme = "http"
			forwardedURL.Host = req.Host
			req.Header[resources.CfForwardedURLHeader] = []string{forwardedURL.String()}

			// Direct the request to the route service.
			req.URL = url
			req.Host = url.Hostname()
		},
	}
	return reverseProxy
}
