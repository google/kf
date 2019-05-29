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

package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
)

// TestIntegration_Doctor runs the doctor command. It ensures the cluster the
// tests are running against is in good shape.
func TestIntegration_Doctor(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		kf.Doctor(ctx)
	})
}

// TestIntegration_Push pushes the echo app (via --built-in command), lists it
// to ensure it can find a domain, uses the proxy command and then posts to
// it. It finally deletes the app.
func TestIntegration_Push(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-echo-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName, "", map[string]string{
			"--built-in":           "",
			"--path":               filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
			"--container-registry": fmt.Sprintf("gcr.io/%s", GCPProjectID()),
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

// TestIntegration_Push_manifest pushes the manifest app, using a manifest.yml
// file. The app is identical to the echo app, and this fact is used to also
// test manifest file environment variables. It finally deletes the app.
func TestIntegration_Push_manifest(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := "manifest-app"

		// Push an app with a manifest file.
		kf.Push(ctx, appName, "./samples/apps/manifest", map[string]string{
			"--container-registry": fmt.Sprintf("gcr.io/%s", GCPProjectID()),
		})
		defer kf.Delete(ctx, appName)

		// List the apps and make sure we can find a domain.
		Logf(t, "ensuring app has domain...")
		apps := kf.Apps(ctx)
		if apps[appName].Domain == "" {
			t.Fatalf("empty domain")
		}
		Logf(t, "done ensuring manifest-app has domain.")

		// TODO: Use port 0 so that we don't have to worry about port
		// collisions. This doesn't work yet:
		// https://github.com/poy/kf/issues/46
		go kf.Proxy(ctx, appName, 8080)

		// Check manifest file environment variables by curling the app
		checkVars(ctx, t, kf, appName, 8080, map[string]string{
			"WHATNOW": "BROWNCOW",
		}, nil)
	})
}

// TestIntegration_Delete pushes an app and then deletes it. It then makes
// sure it is marked as "Deleting".
func TestIntegration_Delete(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-echo-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// simplies replies with the same body that was posted.
		kf.Push(ctx, appName, "", map[string]string{
			"--built-in":           "",
			"--path":               filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
			"--container-registry": fmt.Sprintf("gcr.io/%s", GCPProjectID()),
		})

		// This is only in place for cleanup if the test fails.
		defer kf.Delete(ctx, appName)

		// List the apps and make sure we can find the app.
		Logf(t, "ensuring app is there...")
		_, ok := kf.Apps(ctx)[appName]
		AssertEqual(t, "app presense", true, ok)
		Logf(t, "done ensuring app is there.")

		// Delete the app.
		kf.Delete(ctx, appName)

		// Make sure the app is "deleting"
		// List the apps and make sure we can find the app.
		Logf(t, "ensuring app is deleting...")
		app := kf.Apps(ctx)[appName]
		AssertEqual(t, "reason", "Deleting", app.Reason)
		Logf(t, "done ensuring app is deleting.")
	})
}

// TestIntegration_Envs pushes the env sample app. It sets two variables while
// pushing, and another via SetEnv. It then reads them via Env. It then unsets
// one via Unset and reads them again via Env.
func TestIntegration_Envs(t *testing.T) {
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-envs-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the envs app which
		// returns the set environment variables via JSON. Set two environment
		// variables (ENV1 and ENV2).
		kf.Push(ctx, appName, "", map[string]string{
			"--built-in":           "",
			"--path":               filepath.Join(RootDir(ctx, t), "./samples/apps/envs"),
			"--container-registry": fmt.Sprintf("gcr.io/%s", GCPProjectID()),
			"--env":                "ENV1=VALUE1",
			"--env=ENV2=VALUE2":    "",
		})

		// This is only in place for cleanup if the test fails.
		defer kf.Delete(ctx, appName)

		// TODO: Use port 0 so that we don't have to worry about port
		// collisions. This doesn't work yet:
		// https://github.com/poy/kf/issues/46
		go kf.Proxy(ctx, appName, 8081)

		checkVars(ctx, t, kf, appName, 8081, map[string]string{
			"ENV1": "VALUE1", // Set on push
			"ENV2": "VALUE2", // Set on push
		}, nil)

		// Unset the environment variables ENV1.
		kf.UnsetEnv(ctx, appName, "ENV1")

		// Overwrite ENV2 via set-env
		kf.SetEnv(ctx, appName, "ENV2", "OVERWRITE2")

		checkVars(ctx, t, kf, appName, 8081, map[string]string{
			"ENV2": "OVERWRITE2", // Set on push and overwritten via set-env
		}, []string{"ENV1"})
	})
}

// CheckVars uses Env and the output of the app to ensure the expected
// variables. It will retry for 5 seconds if the environment variables
// aren't returning the correct values. This is to give the
// enventually consistent system time to catch-up.
func checkVars(ctx context.Context, t *testing.T, kf *Kf, appName string, proxyPort int, expectedVars map[string]string, absentVars []string) {

	ctx, _ = context.WithTimeout(ctx, 90*time.Second)
	var success bool
	for !success {
		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled before reaching sucessful check")
		default:
		}

		// List the environment variables and ensure they are all
		// present.
		envs := kf.Env(ctx, appName)

		// The envs app will return all its environment variables as
		// JSON. This checks to make sure everything is ACTUALLY being
		// set from the app's perspective.
		Logf(t, "hitting app %s to check the envs...", appName)
		resp := RetryPost(ctx, t, fmt.Sprintf("http://localhost:%d", proxyPort), 5*time.Second, nil)
		defer resp.Body.Close()
		AssertEqual(t, "status code", http.StatusOK, resp.StatusCode)
		var appEnvs map[string]string
		AssertNil(t, "json", json.NewDecoder(resp.Body).Decode(&appEnvs))
		Logf(t, "done hitting %s app to check the envs.", appName)

		// Check all the environment variables.
		success = true
		for k, v := range expectedVars {
			if envs[k] != v {
				Logf(t, "wrong: %s != %s", envs[k], v)
				success = false
				break
			}
			if appEnvs[k] != v {
				Logf(t, "wrong: %s != %s", appEnvs[k], v)
				success = false
				break
			}
		}

		// Ensure all of the absentVars are NOT there (used to test
		// unset-env).
		for _, k := range absentVars {
			if _, ok := envs[k]; ok {
				Logf(t, "wrong: %s still is present", k)
				success = false
				break
			}
			if _, ok := appEnvs[k]; ok {
				Logf(t, "wrong: %s still is present", k)
				success = false
				break
			}
		}

		// No need to bombard it.
		if !success {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
