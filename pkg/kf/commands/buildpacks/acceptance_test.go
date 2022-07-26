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

package buildpacks

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/acceptance"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	"sigs.k8s.io/yaml"
)

func setupCfPhpInfo() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "cf-ex-php-info",
		Repo: "http://github.com/cloudfoundry-samples/cf-ex-php-info",
	}
}

func setupDotnetCoreHelloWorld() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "dotnet-core-hello-world",
		Repo: "http://github.com/cloudfoundry-samples/dotnet-core-hello-world",
		Setup: func(t *testing.T) {
			csproj, err := ioutil.ReadFile("dotnet-core-hello-world.csproj")
			testutil.AssertNil(t, "ioutil.ReadFile", err)
			modified := strings.Replace(string(csproj), "netcoreapp2.1", "netcoreapp3.1", 1)
			err = ioutil.WriteFile("dotnet-core-hello-world.csproj", []byte(modified), os.ModeAppend)
			testutil.AssertNil(t, "ioutil.WriteFile", err)
		},
	}
}

func setupJavaSpringMusic() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "spring-music",
		Repo: "http://github.com/cloudfoundry-samples/spring-music",
		Setup: func(t *testing.T) {
			cmd := exec.CommandContext(context.Background(), "./gradlew", "clean", "assemble")
			out, err := cmd.CombinedOutput()
			testutil.AssertNil(t, "./gradlew clean assemble", err)
			integration.Logf(t, string(out))
		},
	}
}

func setupGoTestApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "test-go-app",
		Repo: "http://github.com/cloudfoundry-samples/test-app",
	}
}

func setupBinaryApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "binary-app",
		Repo: "http://github.com/cloudfoundry-samples/test-app",
		Path: "/bin",
		Setup: func(t *testing.T) {
			newpath := path.Join(".", "bin")
			err := os.MkdirAll(newpath, os.ModePerm)
			cmd := exec.CommandContext(context.Background(), "go", "build", "main.go")
			out, err := cmd.CombinedOutput()
			testutil.AssertNil(t, "go build main.go", err)
			integration.Logf(t, string(out))

			cmd = exec.CommandContext(context.Background(), "mv", "main", "bin/main")
			out, err = cmd.CombinedOutput()
			testutil.AssertNil(t, "mv main bin/main", err)
			integration.Logf(t, string(out))

			man, err := manifest.New("binary-app")
			man.Applications[0].Command = "./main"
			man.Applications[0].LegacyBuildpack = "binary_buildpack"
			testutil.AssertNil(t, "manifest.New", err)
			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("bin/manifest.yml", yamlData, os.ModePerm))
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("bin/Procfile", []byte("web: ./main"), os.ModePerm))
		},
	}
}

func setupNodeJsApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "nodejs-app",
		Repo: "https://github.com/cloudfoundry-samples/cf-sample-app-nodejs",
	}
}

func setupRubyApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "ruby-app",
		Repo: "https://github.com/cloudfoundry-samples/ruby-sample-app",
		Setup: func(t *testing.T) {
			man, err := manifest.New("ruby-app")
			testutil.AssertNil(t, "manifest.New", err)
			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("manifest.yml", yamlData, os.ModePerm))
		},
	}
}

func setupNginxApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "nginx-app",
		Repo: "https://github.com/cloudfoundry/nginx-buildpack",
		Path: "/fixtures/mainline",
		Setup: func(t *testing.T) {
			man, err := manifest.New("nginx-app")
			testutil.AssertNil(t, "manifest.New", err)
			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("fixtures/mainline/manifest.yml", yamlData, os.ModePerm))
		},
	}
}

func setupPythonApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "python-app",
		Repo: "https://github.com/benwilcock/buildpacks-python-demo",
		Setup: func(t *testing.T) {
			man, err := manifest.New("python-app")
			testutil.AssertNil(t, "manifest.New", err)
			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("manifest.yml", yamlData, os.ModePerm))
		},
	}
}

// TODO http://b/208709905
func setupStaticSite() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "static-app",
		Repo: "https://github.com/govau/cf-example-staticfile",
		Setup: func(t *testing.T) {
			man, err := manifest.NewFromFile(context.Background(), "manifest.yml", nil)
			testutil.AssertNil(t, "manifest.NewFromFile", err)
			if l := len(man.Applications); l != 1 {
				t.Fatalf("expected one application, found %d", l)
			}

			man.Applications[0].LegacyBuildpack = ""

			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("manifest.yml", yamlData, os.ModePerm))
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("Staticfile", []byte("root: web"), os.ModePerm))
		},
	}
}
func TestAcceptance_V2ToV3Stack(t *testing.T) {
	t.Parallel()
	acceptance.RunTest(
		t,
		[]acceptance.SourceCode{
			// setupCfPhpInfo(),
			setupDotnetCoreHelloWorld(),
			setupGoTestApp(),
			setupNodeJsApp(),
			// setupRubyApp(),
			setupNginxApp(),
			setupPythonApp(),
			// setupStaticSite(),
			setupBinaryApp(),
		},
		func(ctx context.Context, t *testing.T, kf *integration.Kf, appPath string) {
			t.Parallel()
			kf.Push(ctx, "",
				"--path", appPath,
				"--stack", "kf-v2-to-v3-shim",
				"--manifest", path.Join(appPath, "manifest.yml"),
			)

			// There will only be this App in the Space.
			info := kf.Apps(ctx)
			if l := len(info); l != 1 {
				t.Fatalf("expected there to be one App, but there were %d", l)
			}

			for name := range info {
				kf.WithProxy(ctx, name, func(addr string) {
					_, clean := integration.RetryGet(ctx, t, addr, 1*time.Minute, http.StatusOK)
					clean()
				})
			}
		})
}

// Test java App separately because V2 java buildpack requires source code in a jar file.
// the `--path=appPath` (appPath = `./`) argument used for other Apps will override
// the `path` field (path = `./build/libs/spring-music-1.0.jar`) in the manifest.yml file of
// http://github.com/cloudfoundry-samples/spring-music.
func TestAcceptance_V2ToV3Stack_java(t *testing.T) {
	t.Parallel()
	acceptance.RunTest(
		t,
		[]acceptance.SourceCode{
			setupJavaSpringMusic(),
		},
		func(ctx context.Context, t *testing.T, kf *integration.Kf, appPath string) {
			t.Parallel()
			kf.Push(ctx, "",
				"--stack", "kf-v2-to-v3-shim",
				"--manifest", path.Join(appPath, "manifest.yml"),
			)

			// There will only be this App in the Space.
			info := kf.Apps(ctx)
			if l := len(info); l != 1 {
				t.Fatalf("expected there to be one App, but there were %d", l)
			}

			for name := range info {
				kf.WithProxy(ctx, name, func(addr string) {
					_, clean := integration.RetryGet(ctx, t, addr, 1*time.Minute, http.StatusOK)
					clean()
				})
			}
		})
}
