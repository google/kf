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

package routes_test

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/google/kf/pkg/kf/testutil"
)

// TestIntegration_Routes creates a route via `create-route`, verifies it with
// `routes`, deletes it via `delete-route` and then verifies again.
func TestIntegration_Routes(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := "example.com"
		path := "some-path"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_UnmappedRoute creates a route via `create-route` that is not mapped to an app.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 503
// with `proxy-route`.
func TestIntegration_UnmappedRoute(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := "example.com"
		path := "mypath"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		go kf.ProxyRoute(ctx, routeHost, 8083)

		{
			resp, respCancel := RetryGet(ctx, t, "http://localhost:8083/"+path, 90*time.Second, http.StatusServiceUnavailable)
			defer resp.Body.Close()
			defer respCancel()
			Logf(t, "testing for 503")
		}

		kf.DeleteRoute(ctx, "example.com", "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_MapRoute pushes an app and creates a route via `create-route`, then maps the app to the route.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 200 OK
// with `proxy-route`.
func TestIntegration_MapRoute(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-routes-%d", time.Now().UnixNano())

		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/helloworld"),
		)
		defer kf.Delete(ctx, appName)

		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := "example.com"
		path := "mypath"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		go kf.ProxyRoute(ctx, routeHost, 8083)

		{
			resp, respCancel := RetryGet(ctx, t, "http://localhost:8083/"+path, 90*time.Second, http.StatusOK)
			defer resp.Body.Close()
			defer respCancel()
			Logf(t, "testing for 200")
		}

		kf.UnmapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

func findRoute(ctx context.Context, t *testing.T, kf *Kf, hostname, domain, path string, shouldFind bool) {
	// TODO (#699): Stop using panics for flow control
	RetryOnPanic(ctx, t, func() {
		var found bool
		for _, line := range kf.Routes(ctx) {
			expected := fmt.Sprintf("%s %s /%s", hostname, domain, path)
			actual := strings.Join(strings.Fields(line), " ")
			if expected == actual {
				found = true
				break
			}
		}

		if shouldFind != found {
			// We'll panic so we can use our retry logic
			panic(fmt.Errorf("Wanted %v, got %v", shouldFind, found))
		}
	})
}
