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

package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/v2/pkg/dockerutil"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	"github.com/google/kf/v2/pkg/sourceimage"
)

const (
	// RunAcceptanceTestsEnv is the environment variable that when set to
	// 'true', will instruct the test suite to run the acceptance tests.
	RunAcceptanceTestsEnv = "RUN_ACCEPTANCE"

	// CacheAcceptanceTestsEnv is the environment variable that when set to
	// 'true', will instruct the test suite to cache the application
	// directories (after being downloaded and setup) to ACCEPTANCE_IMAGE. It
	// is invalid to not have the ACCEPTANCE_IMAGE set when
	// CACHE_ACCEPTANCE_IMAGE is set to 'true'.
	CacheAcceptanceTestsEnv = "CACHE_ACCEPTANCE_IMAGE"

	// AcceptanceTestsImageEnv is the environment variable that when set, is
	// used as either the destination (when CACHE_ACCEPTANCE_IMAGE is set to
	// 'true'), or the location of the setup source code for acceptance tests.
	AcceptanceTestsImageEnv = "ACCEPTANCE_IMAGE"
)

// SourceCode is the key of the map provided to the RunTest. It describes the
// source code the tests should download.
type SourceCode struct {
	// Name is used in t.Run(...) and as the directory name when cloning the
	// repo.
	Name string

	// Repo is the URL for git to download (e.g.,
	// https://github.com/someorg/someproject).
	Repo string

	// Path is the directory of the App source code.
	Path string

	// Setup is invoked after the repo is cloned. It can be nil.
	Setup func(*testing.T)
}

// KfAcceptanceTest is a test ran by RunTest.
type KfAcceptanceTest func(ctx context.Context, t *testing.T, kf *integration.Kf, appPath string)

// RunTest is similar to RunKfTest but for acceptance tests instead
// of integration tests. It will invoke the test for each app.
func RunTest(t *testing.T, apps []SourceCode, test KfAcceptanceTest) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	integration.CancelOnSignal(ctx, cancel, t)

	if ShouldSkipAcceptance(t) {
		return
	}

	// We fetch the source code early (before compiling or checking the
	// cluster) in case the operation is only to save a copy of the source
	// code to the cache.
	appSourceLocations := fetchApps(ctx, t, apps)

	ctx = integration.ContextWithStackdriverOutput(ctx)

	kfPath := integration.CompileKf(ctx, t)

	// Ensure the cluster is healthy
	integration.CheckCluster(ctx, integration.NewKf(t, kfPath))

	for name, dir := range appSourceLocations {
		t.Run(name, func(t *testing.T) {
			// We want a new kf for each one so that it uses the
			// correct 't'.
			kf := integration.NewKf(t, kfPath)

			// Create the space
			spaceName, _ := kf.CreateSpace(ctx)

			ctx = integration.ContextWithSpace(ctx, spaceName)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Wait for space to become ready
			testutil.AssertRetrySucceeds(ctx, t, -1, time.Second, func() error {
				for _, s := range kf.Spaces(ctx) {
					if strings.HasPrefix(s, spaceName) &&
						// Ensure space is marked "Ready True"
						regexp.MustCompile(`\sTrue\s`).MatchString(s) {
						return nil
					}
				}
				return fmt.Errorf("%s-> did not find space %s", t.Name(), spaceName)
			})
			test(ctx, t, kf, dir)
		})
	}
}

func fetchApps(ctx context.Context, t *testing.T, apps []SourceCode) map[string]string {
	tmpDir := integration.CreateTempDir(ctx, t)
	cwd, err := os.Getwd()
	testutil.AssertNil(t, "os.Getwd", err)
	defer func() {
		testutil.AssertNil(t, "os.Chdir", os.Chdir(cwd))
	}()

	// Check to see if we should used cached image.
	if acceptanceImage := os.Getenv(AcceptanceTestsImageEnv); acceptanceImage != "" && os.Getenv(CacheAcceptanceTestsEnv) != "true" {
		return useCachedImage(t, tmpDir, acceptanceImage)
	}

	dirs := map[string]string{}
	for _, app := range apps {
		testutil.AssertNil(t, "os.Chdir err", os.Chdir(tmpDir))
		appDir := filepath.Join(tmpDir, app.Name)
		dirs[app.Name] = path.Join(appDir, app.Path)

		// Clone the app repo
		cmd := exec.CommandContext(ctx, "git", "clone", app.Repo, appDir)
		out, err := cmd.CombinedOutput()
		testutil.AssertNil(t, "git clone", err)
		integration.Logf(t, string(out))

		// Setup the app directory
		if app.Setup != nil {
			testutil.AssertNil(t, "os.Chdir err", os.Chdir(appDir))
			app.Setup(t)
		}
	}

	cacheDir(t, tmpDir)
	return dirs
}

func useCachedImage(t *testing.T, dir, acceptanceImage string) map[string]string {
	// Download and extract source
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		t.Fatalf("os.Mkdir err: %v", err)
	}

	dockerutil.DescribeDefaultConfig(os.Stdout)
	imageRef, err := name.ParseReference(acceptanceImage, name.WeakValidation)
	testutil.AssertNil(t, "name.ParseReference", err)

	image, err := remote.Image(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	testutil.AssertNil(t, "remote.Image", err)

	testutil.AssertNil(t, "sourceImage.ExtractImage",
		sourceimage.ExtractImage(dir, sourceimage.DefaultSourcePath, image),
	)

	files, err := ioutil.ReadDir(dir)
	testutil.AssertNil(t, "ioutil.ReadDir", err)

	m := map[string]string{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		m[file.Name()] = filepath.Join(dir, file.Name())
	}

	return m
}

func cacheDir(t *testing.T, dir string) {
	// At this point, everything is setup. So if the intent was to setup the
	// cache, then do so then fail. This is similar to when the golden rules
	// are updated.
	if os.Getenv(CacheAcceptanceTestsEnv) == "true" {
		acceptanceImage := os.Getenv(AcceptanceTestsImageEnv)
		if acceptanceImage == "" {
			t.Fatalf(
				"it is invalid to set %s to 'true' without setting %s",
				CacheAcceptanceTestsEnv,
				AcceptanceTestsImageEnv,
			)
		}

		image, err := sourceimage.PackageSourceDirectory(dir, func(string) bool { return true })
		testutil.AssertNil(t, "sourceimage.PackageSourceDirectory err", err)

		_, err = sourceimage.PushImage(acceptanceImage, image, false)
		testutil.AssertNil(t, "sourceimage.PushImage err", err)

		t.Fatalf(
			"%s is set to true. Not running tests. Run tests with: %s=false %s=%s %s=true",
			CacheAcceptanceTestsEnv,
			CacheAcceptanceTestsEnv,
			AcceptanceTestsImageEnv,
			acceptanceImage,
			RunAcceptanceTestsEnv,
		)
	}
}

// ShouldSkipAcceptance returns true if acceptance tests are being skipped.
func ShouldSkipAcceptance(t *testing.T) bool {
	t.Helper()

	if !strings.HasPrefix(t.Name(), "TestAcceptance_") {
		// We want to enforce a convention so scripts can single out
		// integration tests.
		t.Fatalf("Acceptance tests must have the name format of 'TestAcceptance_XXX`")
		return true
	}

	if testing.Short() {
		t.Skipf("Skipping %s because short tests were requested", t.Name())
		return true
	}

	if os.Getenv(RunAcceptanceTestsEnv) != "true" && os.Getenv(CacheAcceptanceTestsEnv) != "true" {
		t.Skipf("Skipping %s because neither %s or %s were not true", t.Name(), RunAcceptanceTestsEnv, CacheAcceptanceTestsEnv)
		return true
	}

	return false
}
