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

package main_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
)

// TestIntegration_Push is VERY similar to pkg/kf/commands/apps
// TestIntegration_Push. It omits the --built-in and --container-registry
// flags.
func TestIntegration_Push(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-echo-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName, map[string]string{
			"--path": filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		})
		defer kf.Delete(ctx, appName)

		// List the apps and make sure we can find a domain.
		Logf(t, "ensuring app has domain...")
		apps := kf.Apps(ctx)
		if apps[appName].Domain == "" {
			t.Fatalf("empty domain")
		}
		Logf(t, "done ensuring app has domain.")

		// Hit the app via the proxy. This makes sure the app is handling
		// traffic as expected and ensures the proxy works. We use the proxy
		// for two reasons:
		// 1. Test the proxy.
		// 2. Tests work even if a domain isn't setup.
		Logf(t, "hitting echo app to ensure its working...")

		// TODO: Use port 0 so that we don't have to worry about port
		// collisions. This doesn't work yet:
		// https://github.com/poy/kf/issues/46
		go kf.Proxy(ctx, appName, 8080)
		resp := RetryPost(ctx, t, "http://localhost:8080", 5*time.Second, strings.NewReader("testing"))
		defer resp.Body.Close()
		AssertEqual(t, "status code", http.StatusOK, resp.StatusCode)
		data, err := ioutil.ReadAll(resp.Body)
		AssertNil(t, "body error", err)
		AssertEqual(t, "body", "testing", string(data))
		Logf(t, "done hitting echo app to ensure its working.")
	})
}
