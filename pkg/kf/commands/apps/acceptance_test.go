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

package apps

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
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

func setupSimpleChineseApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "simple-chinese-app",
		Repo: "http://github.com/cloudfoundry-samples/simple-chinese-app",
		Setup: func(t *testing.T) {
			man, err := manifest.New("simple-chinese-app")
			testutil.AssertNil(t, "manifest.New", err)
			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("manifest.yml", yamlData, os.ModePerm))
		},
	}
}

func setupSpringMusic() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "spring-music",
		Repo: "http://github.com/cloudfoundry-samples/spring-music",
		Setup: func(t *testing.T) {
			man, err := manifest.NewFromFile(context.Background(), "manifest.yml", nil)
			testutil.AssertNil(t, "manifest.NewFromFile", err)
			if l := len(man.Applications); l != 1 {
				t.Fatalf("expected one application, found %d", l)
			}

			man.Applications[0].Path = ""
			delete(man.Applications[0].Env, "JBP_CONFIG_SPRING_AUTO_RECONFIGURATION")
			man.Applications[0].Env["BP_AUTO_RECONFIGURATION_ENABLED"] = "false"
			man.Applications[0].Stack = "org.cloudfoundry.stacks.cflinuxfs3"

			yamlData, err := yaml.Marshal(man)
			testutil.AssertNil(t, "yaml.Marshal", err)
			testutil.AssertNil(t, "ioutil.WriteFile", ioutil.WriteFile("manifest.yml", yamlData, os.ModePerm))
		},
	}
}

func setupTestApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "test-app",
		Repo: "http://github.com/cloudfoundry-samples/test-app",
	}
}

func TestAcceptance_Get200(t *testing.T) {
	t.Parallel()
	acceptance.RunTest(
		t,
		[]acceptance.SourceCode{
			setupCfPhpInfo(),
			setupDotnetCoreHelloWorld(),
			setupSimpleChineseApp(),
			setupSpringMusic(),
			setupTestApp(),
		},
		func(ctx context.Context, t *testing.T, kf *integration.Kf, appPath string) {
			t.Parallel()
			kf.Push(ctx, "",
				"--path", appPath,
				"--manifest", path.Join(appPath, "manifest.yml"),
			)

			// There will only be this app in the space.
			info := kf.Apps(ctx)
			if l := len(info); l != 1 {
				t.Fatalf("expected there to be one app, but there were %d", l)
			}

			for name := range info {
				kf.WithProxy(ctx, name, func(addr string) {
					_, clean := integration.RetryGet(ctx, t, addr, 1*time.Minute, http.StatusOK)
					clean()
				})
			}
		})
}
