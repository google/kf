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
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/google/kf/pkg/kf/manifest"
	. "github.com/google/kf/pkg/kf/testutil"
)

const (
	appTimeout = 90 * time.Second
)

// TestIntegration_Push pushes the echo app, lists it to ensure it can find a
// domain, uses the proxy command and then posts to it. It finally deletes the
// app.
func TestIntegration_Push(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-push-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		)
		defer kf.Delete(ctx, appName)
		checkEchoApp(ctx, t, kf, appName, 8080, ExpectedAddr(appName, ""))
	})
}

// TestIntegration_Push_update pushes the echo app, uses the proxy command and
// then posts to it. It then updates the app to the helloworld app by pushing
// to the same app name. It finally deletes the app.
func TestIntegration_Push_update(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-push-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		)
		defer kf.Delete(ctx, appName)
		checkEchoApp(ctx, t, kf, appName, 8087, ExpectedAddr(appName, ""))
		checkHelloWorldApp(ctx, t, kf, appName, 8088, ExpectedAddr(appName, ""))
	})
}

// TestIntegration_Push_docker pushes the echo app via a prebuilt docker
// image, lists it to ensure it can find a domain, uses the proxy command and
// then posts to it. It finally deletes the app.
func TestIntegration_Push_docker(t *testing.T) {
	checkClusterStatus(t)
	t.Parallel()
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-push-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--docker-image=gcr.io/kf-releases/echo-app",
		)
		defer kf.Delete(ctx, appName)
		checkEchoApp(ctx, t, kf, appName, 8086, ExpectedAddr(appName, ""))
	})
}

// TestIntegration_StopStart pushes the echo app and uses the proxy command
// and then posts to it. It posts to it to ensure the app is ready. It then
// stops it and ensures it can no longer reach the app. It then starts it and
// tries posting to it again. It finally deletes the app.
func TestIntegration_StopStart(t *testing.T) {
	t.Skip("This test is slow and flaky")
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-push-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		)
		defer kf.Delete(ctx, appName)

		// Hit the app via the proxy. This makes sure the app is handling
		// traffic as expected and ensures the proxy works. We use the proxy
		// for two reasons:
		// 1. Test the proxy.
		// 2. Tests work even if a domain isn't setup.
		Logf(t, "hitting echo app to ensure it's working...")

		// TODO(#441): Use port 0 so that we don't have to worry about port
		// collisions.
		go kf.Proxy(ctx, appName, 8085)

		{
			resp, respCancel := RetryPost(ctx, t, "http://localhost:8085", appTimeout, http.StatusOK, "testing")
			defer resp.Body.Close()
			defer respCancel()
			Logf(t, "done hitting echo app to ensure it's working.")
		}

		Logf(t, "stoping echo app...")
		kf.Stop(ctx, appName)
		Logf(t, "done stopping echo app.")

		{
			Logf(t, "hitting echo app to ensure it's NOT working...")
			resp, respCancel := RetryPost(ctx, t, "http://localhost:8085", appTimeout, http.StatusNotFound, "testing")
			defer resp.Body.Close()
			defer respCancel()
			Logf(t, "done hitting echo app to ensure it's NOT working.")
		}

		Logf(t, "starting echo app...")
		kf.Start(ctx, appName)
		Logf(t, "done starting echo app.")

		{
			Logf(t, "hitting echo app to ensure it's working...")
			resp, respCancel := RetryPost(ctx, t, "http://localhost:8085", 5*time.Minute, http.StatusOK, "testing")
			defer resp.Body.Close()
			defer respCancel()
			Logf(t, "done hitting echo app to ensure it's working.")
		}
	})
}

// TestIntegration_Push_manifest pushes the manifest app, using a manifest.yml
// file. The app is identical to the env app, and this fact is used to also
// test manifest file environment variables. It finally deletes the app.
func TestIntegration_Push_manifest(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		currentTime := time.Now().UnixNano()
		appName := fmt.Sprintf("integration-manifest-%d", currentTime)
		appPath := filepath.Join(RootDir(ctx, t), "samples", "apps", "manifest")

		// Create a custom manifest file for this test.
		newManifestFile, manifestCleanup, err := copyManifest(appName, appPath, currentTime)
		AssertNil(t, "app manifest copy error", err)
		defer manifestCleanup()

		// Push an app with a manifest file.
		kf.Push(ctx, appName,
			"--path", appPath,
			"--manifest", newManifestFile,
		)
		defer kf.Delete(ctx, appName)

		checkEnvApp(ctx, t, kf, appName, 8082, map[string]string{
			"WHATNOW": "BROWNCOW",
		}, ExpectedAddr(appName, ""))
	})
}

// copyManifest copies the manifest.yml file in a given appPath.
// The copy is edited such that the 1st app is renamed with the given appName.
// The filename of the new manifest is returned, along with a cleanup function.
func copyManifest(appName, appPath string, currentTime int64) (string, func(), error) {
	manifestPath := filepath.Join(appPath, "manifest.yml")
	manifestBytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return "", nil, err
	}

	var manifest manifest.Manifest
	err = yaml.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return "", nil, err
	}

	if len(manifest.Applications) < 1 {
		return "", nil, fmt.Errorf("No applications in manifest file %s", manifestPath)
	}

	manifest.Applications[0].Name = appName
	var newManifestBytes []byte
	newManifestBytes, err = yaml.Marshal(manifest)
	if err != nil {
		return "", nil, err
	}

	newManifestFile := filepath.Join(appPath, fmt.Sprintf("manifest-integration-%d.yml", currentTime))
	err = ioutil.WriteFile(newManifestFile, newManifestBytes, 0644)
	if err != nil {
		return "", nil, err
	}

	return newManifestFile, func() {
		os.Remove(newManifestFile)
	}, nil
}

// TestIntegration_Delete pushes an app and then deletes it. It then makes
// sure it is marked as "Deleting".
func TestIntegration_Delete(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-delete-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// simplies replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		)

		// This is only in place for cleanup if the test fails.
		defer kf.Delete(ctx, appName)

		// List the apps and make sure we can find the app.
		Logf(t, "ensuring app is there...")
		_, ok := kf.Apps(ctx)[appName]
		AssertEqual(t, "app presence", true, ok)
		Logf(t, "done ensuring app is there.")

		// Delete the app.
		kf.Delete(ctx, appName)

		// Make sure the app is "deleting"
		// List the apps and make sure we can find the app.
		Logf(t, "ensuring app is deleting...")
		app := kf.Apps(ctx)[appName]
		AssertEqual(t, "requested state", "deleting", app.RequestedState)
		Logf(t, "done ensuring app is deleting.")
	})
}

// TestIntegration_Envs pushes the env sample app. It sets two variables while
// pushing, and another via SetEnv. It then reads them via Env. It then unsets
// one via Unset and reads them again via Env.
func TestIntegration_Envs(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-envs-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the envs app which
		// returns the set environment variables via JSON. Set two environment
		// variables (ENV1 and ENV2).
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/envs"),
			"--env", "ENV1=VALUE1",
			"--env=ENV2=VALUE2",
		)

		// This is only in place for cleanup if the test fails.
		defer kf.Delete(ctx, appName)

		checkEnvApp(ctx, t, kf, appName, 8081, map[string]string{
			"ENV1": "VALUE1", // Set on push
			"ENV2": "VALUE2", // Set on push
		}, ExpectedAddr(appName, ""))

		t.Run("overwrite envs", func(t *testing.T) {
			t.Skip("this is flaky as knative isn't fast at updating the env values")
			// Unset the environment variables ENV1.
			RetryOnPanic(ctx, t, func() { kf.UnsetEnv(ctx, appName, "ENV1") })

			// Overwrite ENV2 via set-env
			RetryOnPanic(ctx, t, func() { kf.SetEnv(ctx, appName, "ENV2", "OVERWRITE2") })

			assertVars(ctx, t, kf, appName, 8081, map[string]string{
				"ENV2": "OVERWRITE2", // Set on push and overwritten via set-env
			}, []string{"ENV1"})
		})
	})
}

// TestIntegration_Logs pushes the echo app, tails
// it's logs and then posts to it. It then waits for the expected logs. It
// finally deletes the app.
func TestIntegration_Logs(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-logs-%d", time.Now().UnixNano())

		// Push an app and then clean it up. This pushes the echo app which
		// replies with the same body that was posted.
		kf.Push(ctx, appName,
			"--path", filepath.Join(RootDir(ctx, t), "./samples/apps/echo"),
		)
		defer kf.Delete(ctx, appName)

		logOutput := kf.Logs(ctx, appName, "-n=30", "-f")

		// Hit the app via the proxy. This makes sure the app is handling
		// traffic as expected and ensures the proxy works. We use the proxy
		// for two reasons:
		// 1. Test the proxy.
		// 2. Tests work even if a domain isn't setup.

		// TODO(#441): Use port 0 so that we don't have to worry about port
		// collisions.
		go kf.Proxy(ctx, appName, 8083)

		// Write out an expected log line until the test dies. We do this more
		// than once because we can't guarantee much about logs.
		expectedLogLine := fmt.Sprintf("testing-%d", time.Now().UnixNano())
		for i := 0; i < 10; i++ {
			resp, respCancel := RetryPost(ctx, t, "http://localhost:8083", appTimeout, http.StatusOK, expectedLogLine)
			resp.Body.Close()
			respCancel()
		}

		// Wait around for the log to stream out. If it doesn't after a while,
		// fail the test.
		timer := time.NewTimer(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				t.Fatal("context cancelled")
			case <-timer.C:
				t.Fatal("timed out waiting for log line")
			case line := <-logOutput:
				if line == expectedLogLine {
					return
				}
			}
		}
	})
}

// TestIntegration_LogsNoContainer tests that the logs command exits in a
// reasonable amount of time when logging an application that doesn't have a
// container (scaled to 0).
func TestIntegration_LogsNoContainer(t *testing.T) {
	t.Parallel()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		appName := fmt.Sprintf("integration-logs-noc-%d", time.Now().UnixNano())

		output := kf.Logs(ctx, appName)

		timer := time.NewTimer(5 * time.Second)
		for {
			select {
			case <-timer.C:
				t.Fatal("expected kf logs to exit")
			case _, ok := <-output:
				if !ok {
					// Success
					return
				}

				// Not closed, command still going.
			}
		}
	})
}

// assertVars uses Env and the output of the app to ensure the expected
// variables. It will retry for a while if the environment variables aren't
// returning the correct values. This is to give the enventually consistent
// system time to catch-up.
func assertVars(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	proxyPort int,
	expectedVars map[string]string,
	absentVars []string,
) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var success bool
	for !success {
		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled before reaching successful check")
		default:
		}

		// List the environment variables and ensure they are all
		// present.
		envs := kf.Env(ctx, appName)

		// The envs app will return all its environment variables as
		// JSON. This checks to make sure everything is ACTUALLY being
		// set from the app's perspective.
		Logf(t, "hitting app %s to check the envs...", appName)
		resp, respCancel := RetryPost(ctx, t, fmt.Sprintf("http://localhost:%d", proxyPort), appTimeout, http.StatusOK, "")
		defer resp.Body.Close()
		defer respCancel()
		if resp.StatusCode != http.StatusOK {
			Logf(t, "status code %d", resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var appEnvs map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&appEnvs); err != nil {
			Logf(t, "err serializing envs: %s", err)
			Logf(t, "%s", appEnvs)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		Logf(t, "done hitting %s app to check the envs.", appName)

		// Check all the environment variables.
		success = true
		for k, v := range expectedVars {
			if envs[k] != v {
				Logf(t, "(kf env) wrong: %s != %s", envs[k], v)
				success = false
				break
			}
			if appEnvs[k] != v {
				Logf(t, "(response) wrong: %s != %s", appEnvs[k], v)
				success = false
				break
			}
		}

		// Ensure all of the absentVars are NOT there (used to test
		// unset-env).
		for _, k := range absentVars {
			if _, ok := envs[k]; ok {
				Logf(t, "(kf env) wrong: %s still is present", k)
				success = false
				break
			}
			if _, ok := appEnvs[k]; ok {
				Logf(t, "(response) wrong: %s still is present", k)
				success = false
				break
			}
		}

		// No need to bombard it.
		if !success {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if !success {
		t.Fatalf("unsuccessful in checking env vars")
	}
}

var checkOnce sync.Once

func checkClusterStatus(t *testing.T) {
	checkOnce.Do(func() {
		testIntegration_Doctor(t)
	})
}

// testIntegration_Doctor runs the doctor command. It ensures the cluster the
// tests are running against is in good shape.
func testIntegration_Doctor(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		kf.Doctor(ctx)
	})
}

func checkApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedRoutes []string,
	proxyPort int,
	assert func(ctx context.Context, t *testing.T, addr string),
) {
	// List the apps and make sure we have the correct route.
	Logf(t, "ensuring app's route...")
	apps := kf.Apps(ctx)

	sort.Strings(apps[appName].URLs)
	AssertEqual(t, "routes", expectedRoutes, apps[appName].URLs)
	Logf(t, "done ensuring app's route.")

	// Hit the app via the proxy. This makes sure the app is handling
	// traffic as expected and ensures the proxy works. We use the proxy
	// for two reasons:
	// 1. Test the proxy.
	// 2. Tests work even if a domain isn't setup.
	Logf(t, "hitting echo app to ensure its working...")

	// TODO(#46): Use port 0 so that we don't have to worry about port
	// collisions.
	go kf.Proxy(ctx, appName, proxyPort)
	assert(ctx, t, fmt.Sprintf("http://localhost:%d", proxyPort))
}

func checkEchoApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	proxyPort int,
	expectedRoutes ...string,
) {
	checkApp(ctx, t, kf, appName, expectedRoutes, proxyPort, func(ctx context.Context, t *testing.T, addr string) {
		resp, respCancel := RetryPost(ctx, t, addr, appTimeout, http.StatusOK, "testing")
		defer resp.Body.Close()
		defer respCancel()
		data, err := ioutil.ReadAll(resp.Body)
		AssertNil(t, "body error", err)
		AssertEqual(t, "body", "testing", string(data))
		Logf(t, "done hitting echo app to ensure its working.")
	})
}

func checkHelloWorldApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	proxyPort int,
	expectedRoutes ...string,
) {
	checkApp(ctx, t, kf, appName, expectedRoutes, proxyPort, func(ctx context.Context, t *testing.T, addr string) {
		// The helloworld app doesn't care if what verb you use (e.g., POST vs
		// GET), so we'll just use the RetryPost method so we can get the
		// retry logic.
		resp, respCancel := RetryPost(ctx, t, addr, appTimeout, http.StatusOK, "testing")
		defer resp.Body.Close()
		defer respCancel()
		Logf(t, "done hitting helloworld app to ensure its working.")
	})
}

func checkEnvApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	proxyPort int,
	expectedVars map[string]string,
	expectedRoutes ...string,
) {
	checkApp(ctx, t, kf, appName, expectedRoutes, proxyPort, func(ctx context.Context, t *testing.T, addr string) {
		assertVars(ctx, t, kf, appName, proxyPort, expectedVars, nil)
	})
}
