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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

var currentPort int = 8082

// TestIntegration_Routes creates a route via `create-route`, verifies it with
// `routes`, deletes it via `delete-route` and then verifies again.
func TestIntegration_Routes(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := os.Getenv(integration.SpaceDomainEnvVar)
		path := "some-path"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_RoutesCustomDomain pushes an App and creates a custom domain route via `create-route`, then maps the App to the route.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 200 OK
// with `proxy-route`.
func TestIntegration_RoutesCustomDomain(t *testing.T) {
	// Connection to BareMetal using Tailoribird tunnel and kf proxy doesn't work with custom domain, see b/195313679.
	if os.Getenv(integration.IsBareMetalEnvVar) == "true" {
		t.Skip()
	}
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := fmt.Sprintf("integration-routes-%d", time.Now().UnixNano())

		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld"))
		defer kf.Delete(ctx, appName, "--async")

		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := fmt.Sprintf("some-domain-%d.com", time.Now().UnixNano())
		path := "some-path"

		// Configure the space to allow the test domain.
		kf.ConfigureSpace(ctx, "append-domain", domain)

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, true)

		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			url := addr + "/" + path
			resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusOK)
			defer resp.Body.Close()
			defer respCancel()
			integration.Logf(t, "testing for 200")
		})

		kf.UnmapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, false)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_UnmappedRoute creates a route via `create-route` that is not mapped to an App.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 404
// with `proxy-route`.
func TestIntegration_UnmappedRoute(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := os.Getenv(integration.SpaceDomainEnvVar)
		path := "mypath"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			url := addr + "/" + path
			resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusNotFound)
			defer resp.Body.Close()
			defer respCancel()
			integration.Logf(t, "testing for 404")
		})

		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_MapRoute pushes an App and creates a route via `create-route`, then maps the App to the route.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 200 OK
// with `proxy-route`.
func TestIntegration_MapRoute(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := fmt.Sprintf("integration-routes-%d", time.Now().UnixNano())

		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld"))

		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := os.Getenv(integration.SpaceDomainEnvVar)
		path := "mypath"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, true)

		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			url := addr + "/" + path
			resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusOK)
			defer resp.Body.Close()
			defer respCancel()
			integration.Logf(t, "testing for 200")
		})

		kf.UnmapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_InternalRoute pushes an App and creates a internal Route via `create-route`, then maps the App to the Route.
// The test verifies that the route exists with `routes`.
func TestIntegration_InternalRoute(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := fmt.Sprintf("integration-routes-%d", time.Now().UnixNano())

		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld"))
		defer kf.Delete(ctx, appName, "--async")

		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := integration.InternalDefaultDomain
		path := "mypath"

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, true)

		kf.UnmapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_MultipleAppsPerRoute pushes two Apps and creates a route via `create-route`,
// then maps both Apps to the same route. The test verifies that hitting the route returns a 200 OK
// and that traffic to the route is roughly split according to the weights set on the two Apps.
// In this test, there should be about 3x more traffic sent to the hello App than the envs App.
func TestIntegration_MultipleAppsPerRoute(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		helloWorldApp := fmt.Sprintf("integration-hello-%d", time.Now().UnixNano())
		envsApp := fmt.Sprintf("integration-envs-%d", time.Now().UnixNano())

		kf.CachePush(ctx, helloWorldApp, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld"))
		defer kf.Delete(ctx, helloWorldApp, "--async")

		kf.CachePush(ctx, envsApp, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/envs"))
		defer kf.Delete(ctx, envsApp, "--async")

		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())
		domain := os.Getenv(integration.SpaceDomainEnvVar)
		path := "mypath"

		// Weigh the traffic so that the hello world App receives 3x the traffic as envs App
		helloWeight := 3
		envsWeight := 1

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, helloWorldApp, domain, "--hostname="+hostname, "--path="+path, "--weight="+strconv.Itoa(helloWeight))
		kf.MapRoute(ctx, envsApp, domain, "--hostname="+hostname, "--path="+path, "--weight="+strconv.Itoa(envsWeight))

		// Assert that each App lists the route
		findAppRoute(ctx, t, kf, hostname, domain, path, helloWorldApp, true)
		findAppRoute(ctx, t, kf, hostname, domain, path, envsApp, true)

		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			helloWorldCount := 0
			envsCount := 0
			url := addr + "/" + path
			for i := 0; i < 100; i++ {
				resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusOK)
				defer resp.Body.Close()
				defer respCancel()
				if resp.StatusCode == http.StatusOK {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {

					}
					bodyStr := string(body)
					if strings.Contains(bodyStr, "hello") {
						helloWorldCount++
					} else if strings.Contains(bodyStr, "VCAP_SERVICES") {
						envsCount++
					}
				}
			}

			integration.Logf(t, "number of requests routed to hello world App: %d", helloWorldCount)
			integration.Logf(t, "number of requests routed to envs App: %d", envsCount)

			// This method isn't great, but it's difficult to come up with a
			// deterministic test that won't cause false-positives due to the way
			// Istio does weighted routing.
			//
			// A better approach would be a T-test to tell if the sample is from the
			// expected distribution within some margin of error.
			testutil.AssertTrue(t, "route traffic went to helloWorld", helloWorldCount > 0)
			testutil.AssertTrue(t, "route traffic went to envsCount", envsCount > 0)
			testutil.AssertTrue(t, "helloWorld greater than envs", helloWorldCount > envsCount)

			// Unmap the hello world App from the route. Now all traffic should go to the envs App.
			kf.UnmapRoute(ctx, helloWorldApp, domain, "--hostname="+hostname, "--path="+path)

			// We have to wait for the VirtualServices to reconcile because Routes
			// don't reflect real statuses, just that the VS was triggered.
			time.Sleep(5 * time.Second)

			helloWorldCount = 0
			envsCount = 0

			for i := 0; i < 100; i++ {
				resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusOK)
				defer resp.Body.Close()
				defer respCancel()
				if resp.StatusCode == http.StatusOK {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {

					}
					bodyStr := string(body)
					if strings.Contains(bodyStr, "hello") {
						helloWorldCount++
					} else if strings.Contains(bodyStr, "VCAP_SERVICES") {
						envsCount++
					}
				}
			}
			integration.Logf(t, "number of requests routed to hello world App: %d", helloWorldCount)
			integration.Logf(t, "number of requests routed to envs App: %d", envsCount)

			testutil.AssertTrue(t, "route traffic directed to one App", envsCount == 100)

			kf.UnmapRoute(ctx, envsApp, domain, "--hostname="+hostname, "--path="+path)
			findAppRoute(ctx, t, kf, hostname, domain, path, envsApp, false)

			kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
			findRoute(ctx, t, kf, hostname, domain, path, false)
			findAppRoute(ctx, t, kf, hostname, domain, path, helloWorldApp, false)
		})
	})
}

// TestIntegration_WildcardRoute pushes an App and creates a wildcard route via `create-route`, then maps the App to the route.
// The test verifies that the route exists with `routes`, and checks that hitting the route returns a 200 OK
// with `proxy-route`.
func TestIntegration_WildcardRoute(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := fmt.Sprintf("integration-routes-%d", time.Now().UnixNano())

		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld"))
		defer kf.Delete(ctx, appName, "--async")

		hostname := "*"
		domain := os.Getenv(integration.SpaceDomainEnvVar)
		path := "mypath"

		// Configure the space to allow the test domain.
		kf.ConfigureSpace(ctx, "append-domain", domain)

		kf.CreateRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, true)

		kf.MapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, true)

		routeHost := fmt.Sprintf("made-up.%s", domain)
		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			url := addr + "/" + path
			resp, respCancel := integration.RetryGet(ctx, t, url, 90*time.Second, http.StatusOK)
			defer resp.Body.Close()
			defer respCancel()
			integration.Logf(t, "testing for 200")
		})

		kf.UnmapRoute(ctx, appName, domain, "--hostname="+hostname, "--path="+path)
		findAppRoute(ctx, t, kf, hostname, domain, path, appName, false)
		kf.DeleteRoute(ctx, domain, "--hostname="+hostname, "--path="+path)
		findRoute(ctx, t, kf, hostname, domain, path, false)
	})
}

// TestIntegration_RoutesWithPorts pushes an App and creates routes to it with
// ports.
func TestIntegration_RoutesWithPorts(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appSuffix := fmt.Sprintf("-%d", time.Now().UnixNano())
		appName := "multiple-ports" + appSuffix
		appPath := filepath.Join(integration.RootDir(ctx, t), "samples", "apps", "multiple-ports")

		// This can't be a CachePush because it needs the app-suffix flag and
		// a manifest.
		kf.Push(ctx, "multiple-ports",
			"--manifest", filepath.Join(appPath, "manifest.yaml"),
			"--path", appPath,
			"--app-suffix", appSuffix,
		)
		defer kf.Delete(ctx, appName, "--async")

		domain := os.Getenv(integration.SpaceDomainEnvVar)

		// Configure the space to allow the test domain.
		kf.ConfigureSpace(ctx, "append-domain", domain)

		// Map both ports to the same route so traffic will be split between them.
		hostname := "multiple-ports"
		kf.MapRoute(ctx, appName, domain, "--hostname", hostname, "--weight", "3")
		kf.MapRoute(ctx, appName, domain, "--hostname", hostname, "--destination-port", "8888", "--weight", "1")
		defer kf.DeleteRoute(ctx, domain, "--hostname", hostname)

		defaultCount := 0
		customCount := 0
		routeHost := fmt.Sprintf("%s.%s", hostname, domain)
		kf.WithProxyRoute(ctx, routeHost, func(addr string) {
			for i := 0; i < 100; i++ {
				resp, respCancel := integration.RetryGet(ctx, t, addr, 90*time.Second, http.StatusOK)
				defer resp.Body.Close()
				defer respCancel()
				if resp.StatusCode == http.StatusOK {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						continue
					}
					bodyStr := string(body)
					if strings.Contains(bodyStr, "8080") {
						defaultCount++
					} else if strings.Contains(bodyStr, "8888") {
						customCount++
					}
				}
			}
		})

		integration.Logf(t, "number of requests routed to default App: %d", defaultCount)
		integration.Logf(t, "number of requests routed to custom App: %d", customCount)

		// This method isn't great, but it's difficult to come up with a
		// deterministic test that won't cause false-positives due to the way
		// Istio does weighted routing.
		//
		// A better approach would be a T-test to tell if the sample is from the
		// expected distribution within some margin of error.
		testutil.AssertTrue(t, "route traffic went to default port", defaultCount > 0)
		testutil.AssertTrue(t, "route traffic went to custom port", customCount > 0)
		testutil.AssertTrue(t, "default gte 2 x custom", defaultCount >= 2*customCount)
	})
}

func findAppRoute(
	ctx context.Context,
	t *testing.T,
	kf *integration.Kf,
	hostname string,
	domain string,
	urlPath string,
	appName string,
	shouldFind bool,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	urlPath = path.Join("/", urlPath)

	// TODO (#699): Stop using panics for flow control
	testutil.Retry(ctx, -1, 100*time.Millisecond, func() error {
		t.Helper()

		var appInfo v1alpha1.App
		testutil.AssertNil(t, "json.Unmarshal",
			json.Unmarshal(kf.App(ctx, appName), &appInfo),
		)

		{
			// Status
			var found bool
			for _, route := range appInfo.Status.Routes {
				if route.Source.Hostname == hostname && route.Source.Domain == domain && route.Source.Path == urlPath {
					found = true
					break
				}
			}

			if shouldFind != found {
				// We'll panic so we can use our retry logic
				return fmt.Errorf("Status: Wanted %v, got %v", shouldFind, found)
			}
		}

		{
			// Spec
			var found bool
			for _, route := range appInfo.Spec.Routes {
				if route.Hostname == hostname && route.Domain == domain && route.Path == urlPath {
					found = true
					break
				}
			}

			if shouldFind != found {
				// We'll panic so we can use our retry logic
				return fmt.Errorf("Spec: Wanted %v, got %v", shouldFind, found)
			}
		}

		// Success!
		return nil
	})
}

func findRoute(ctx context.Context, t *testing.T, kf *integration.Kf, hostname, domain, path string, shouldFind bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	testutil.Retry(ctx, -1, 100*time.Millisecond, func() error {
		t.Helper()

		// Check `kf routes` for the route
		var found bool
		for _, line := range kf.Routes(ctx) {
			expected := fmt.Sprintf("%s %s /%s", hostname, domain, path)
			actual := strings.Join(strings.Fields(line), " ")
			if strings.Contains(actual, expected) {
				found = true
				break
			}
		}

		if shouldFind != found {
			// We'll panic so we can use our retry logic
			return fmt.Errorf("Wanted %v, got %v", shouldFind, found)
		}

		// Success!
		return nil
	})
}
