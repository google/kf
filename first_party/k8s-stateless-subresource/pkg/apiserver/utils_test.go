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

package apiserver_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/rawhttp"
	"knative.dev/pkg/injection"

	// The following imports provide authentication helpers for GCP and OIDC.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	cleanupMu sync.Mutex
	cleanup   = func() {
		// Starts as a NOP. This function will be replaced if the tests get
		// far enough.
	}
)

func TestMain(m *testing.M) {
	result := m.Run()

	cleanupMu.Lock()
	c := cleanup
	cleanupMu.Unlock()
	c()

	os.Exit(result)
}

func end2endTest(t *testing.T, f func(ctx context.Context)) {
	t.Parallel()

	if testing.Short() {
		t.Skip("end2End tests aren't ran when --short is provided")
	}

	// Ensure `ko` is installed. If not, skip the test.
	if !koInPath() {
		t.Skip("tests require ko to be installed\nInstall https://github.com/google/ko and then rerun the tests")
	}

	if os.Getenv("KO_DOCKER_REPO") == "" {
		t.Skip("KO_DOCKER_REPO must be set to run the tests")
	}

	// Setup contexts.
	ctx := context.Background()
	ctx = withTesting(ctx, t)
	ctx = withRestClient(ctx)

	// Install the sample deployment and register the API service.
	installSample(ctx)

	// Run the test.
	f(ctx)
}

type testingKey struct{}

func withTesting(ctx context.Context, t *testing.T) context.Context {
	return context.WithValue(ctx, testingKey{}, t)
}

func testingFromContext(ctx context.Context) *testing.T {
	return ctx.Value(testingKey{}).(*testing.T)
}

type restClientKey struct{}

func withRestClient(ctx context.Context) context.Context {
	t := testingFromContext(ctx)
	t.Helper()

	config, err := injection.GetRESTConfig("", "")
	if err != nil {
		t.Fatalf("failed to setup REST config: %v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("failed to setup REST config: %v", err)
	}

	return context.WithValue(ctx, restClientKey{}, k8sClient.CoreV1().RESTClient())
}

func restClientFromContext(ctx context.Context) *rest.RESTClient {
	return ctx.Value(restClientKey{}).(*rest.RESTClient)
}

var installSampleOnce sync.Once

func installSample(ctx context.Context) {
	installSampleOnce.Do(func() {
		// The sample exists in the sample directory at the root of the repo.
		sampleDir := filepath.Join(rootDir(ctx), "sample", "config")

		// Set the cleanup. We want to ensure that we attempt to delete the
		// sample even if it doesn't install correctly.
		cleanupMu.Lock()
		cleanup = func() {
			koDelete(ctx, sampleDir)
		}
		cleanupMu.Unlock()

		koApply(ctx, sampleDir)
	})
}

func koApply(ctx context.Context, dir string) {
	t := testingFromContext(ctx)
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ko", "apply", "-f", dir)
	t.Logf("running %s", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to apply resources via ko: %v", err)
	}
}

func koDelete(ctx context.Context, dir string) {
	// We can't use testing.T in this function as it is used for cleanup.

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ko", "delete", "--wait=false", "--ignore-not-found", "-f", dir)
	log.Printf("running %s", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Just give a headsup.
		log.Printf("failed to delete resources (this is likely OK): %v", err)
	}
}

func rootDir(ctx context.Context) string {
	t := testingFromContext(ctx)
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "env", "GOMOD")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get Stdout pipe for RootDir: %s", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command to get RootDir %s", err)
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Fatalf("failed to read Stdout pipe for RootDir: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("failed to get RootDir %s", err)
	}

	return filepath.Dir(strings.TrimSpace(string(data)))
}

func koInPath() bool {
	_, err := exec.LookPath("ko")
	return err == nil
}

// get is basically doing `kubectl get --raw {path}`. It has a little retry
// logic baked in to help avoid common race conditions (e.g., server is up but
// not ready).
func get(ctx context.Context, path string, f func(map[string]interface{}) bool) {
	t := testingFromContext(ctx)
	t.Helper()

	client := restClientFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var output string
	var err error

	for ctx.Err() == nil {
		for ctx.Err() == nil {
			ioStreams, _, out, errOut := genericclioptions.NewTestIOStreams()
			t.Logf("GETting %s", path)
			if err = rawhttp.RawGet(
				client,
				ioStreams,
				path,
			); err != nil {
				// Looks like it failed. If there is anything useful on out or
				// err go ahead and show it. Then retry.
				t.Logf("failed to make GET request (retrying): %v", err)
				if out.String() != "" {
					t.Logf("(out = %q)", out.String())
				}
				if errOut.String() != "" {
					t.Logf("(err = %q)", errOut.String())
				}
				time.Sleep(250 * time.Millisecond)
				continue
			}
			output = out.String()
		}

		if err != nil {
			t.Fatalf("failed make request: %v", err)
		}

		// Decode the data from JSON for the function.
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(output), &m); err != nil {
			t.Fatalf("failed to unmarshal data (output=%s): %v", output, err)
		}

		if f(m) {
			// Move on.
			return
		}

		// Wait a moment, then retry.
		time.Sleep(250 * time.Millisecond)
	}
}
