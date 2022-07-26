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

package integration

import (
	context "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

const (
	appTimeout = 90 * time.Second
)

// CheckEchoApp checks the echo App.
func CheckEchoApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedRoutes ...string,
) {
	CheckApp(ctx, t, kf, appName, expectedRoutes, func(ctx context.Context, t *testing.T, addr string) {
		resp, respCancel := RetryPost(ctx, t, addr, appTimeout, http.StatusOK, "testing")
		defer resp.Body.Close()
		defer respCancel()
		data, err := ioutil.ReadAll(resp.Body)
		testutil.AssertNil(t, "body error", err)
		testutil.AssertEqual(t, "body", "testing", string(data))
		Logf(t, "done hitting echo App to ensure its working.")
	})
}

func CheckSigtermApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
) {
	go func() {
		time.Sleep(3 * time.Second)
		kf.Stop(ctx, appName)
	}()
	kf.VerifyLogOutput(ctx, appName, "SIGTERM received", 20*time.Second)
}

// CheckHelloWorldApp checks the hello world App.
func CheckHelloWorldApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedRoutes ...string,
) {
	CheckApp(ctx, t, kf, appName, expectedRoutes, func(ctx context.Context, t *testing.T, addr string) {
		// The helloworld App doesn't care if what verb you use (e.g., POST vs
		// GET), so we'll just use the RetryPost method so we can get the
		// retry logic.
		resp, respCancel := RetryPost(ctx, t, addr, appTimeout, http.StatusOK, "testing")
		defer resp.Body.Close()
		defer respCancel()

		data, err := ioutil.ReadAll(resp.Body)
		testutil.AssertNil(t, "body error", err)

		expectedBody := fmt.Sprintf("hello from %s!", appName)
		testutil.AssertEqual(t, "body", expectedBody, strings.TrimSpace(string(data)))
		Logf(t, "done hitting helloworld App to ensure its working.")
	})
}

// CheckEnvApp checks the env App.
func CheckEnvApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedVars map[string]string,
	expectedRoutes ...string,
) {
	CheckApp(ctx, t, kf, appName, expectedRoutes, func(ctx context.Context, t *testing.T, addr string) {
		AssertVars(ctx, t, kf, appName, expectedVars, nil)
	})
}

// CheckApp runs the given function against the given App.
func CheckApp(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedRoutes []string,
	assert func(ctx context.Context, t *testing.T, addr string),
) {
	// Get the App and make sure we have the correct route(s).
	Logf(t, "ensuring App's route...")
	appJSON := kf.App(ctx, appName)

	app := &v1alpha1.App{}
	if err := json.Unmarshal([]byte(appJSON), app); err != nil {
		t.Fatal("unmarshaling App:", err)
	}

	var urls []string
	for _, route := range app.Status.Routes {
		urls = append(urls, route.Source.String())
	}

	sort.Strings(urls)
	testutil.AssertEqual(t, "routes", expectedRoutes, urls)
	Logf(t, "done ensuring App's route.")

	// Hit the App via the proxy. This makes sure the App is handling
	// traffic as expected and ensures the proxy works. We use the proxy
	// for two reasons:
	// 1. Test the proxy.
	// 2. Tests work even if a domain isn't setup.
	Logf(t, "hitting %s App to ensure its working...", appName)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	kf.WithProxy(ctx, appName, func(url string) {
		assert(ctx, t, url)
	})
}

// AssertVars uses Env and the output of the App to ensure the expected
// variables. It will retry for a while if the environment variables aren't
// returning the correct values. This is to give the enventually consistent
// system time to catch-up.
func AssertVars(
	ctx context.Context,
	t *testing.T,
	kf *Kf,
	appName string,
	expectedVars map[string]string,
	absentVars []string,
) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	testutil.AssertRetrySucceeds(ctx, t, -1, 100*time.Millisecond, func() error {
		// List the environment variables and ensure they are all
		// present.
		envs := kf.Env(ctx, appName)

		var (
			appEnvs map[string]string
			err     error
		)
		kf.WithProxy(ctx, appName, func(url string) {
			// The envs App will return all its environment variables as
			// JSON. This checks to make sure everything is ACTUALLY being
			// set from the App's perspective.
			Logf(t, "hitting App %s to check the envs...", appName)
			resp, respCancel := RetryPost(ctx, t, url, appTimeout, http.StatusOK, "")
			defer resp.Body.Close()
			defer respCancel()
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("status code %d", resp.StatusCode)
				return
			}
			if err := json.NewDecoder(resp.Body).Decode(&appEnvs); err != nil {
				err = fmt.Errorf("failed to serialize envs: %v", err)
				return
			}
			Logf(t, "done hitting %s App to check the envs.", appName)
		})

		if err != nil {
			return err
		}

		// Check all the environment variables.
		for k, v := range expectedVars {
			if envs[k] != v {
				return fmt.Errorf("(kf env) wrong: %s != %s", envs[k], v)
			}
			if appEnvs[k] != v {
				return fmt.Errorf("(response) wrong: %s != %s", appEnvs[k], v)
			}
		}

		// Ensure all of the absentVars are NOT there (used to test
		// unset-env).
		for _, k := range absentVars {
			if _, ok := envs[k]; ok {
				return fmt.Errorf("(kf env) wrong: %s still is present", k)
			}
			if _, ok := appEnvs[k]; ok {
				return fmt.Errorf("(response) wrong: %s still is present", k)
			}
		}

		// Success!
		return nil
	})
}
