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

package servicebindings_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/acceptance"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

func setupNFSApp() acceptance.SourceCode {
	return acceptance.SourceCode{
		Name: "cf-volume-services-acceptance-tests",
		Repo: "https://github.com/cloudfoundry/cf-volume-services-acceptance-tests.git",
	}
}

func TestAcceptance_NFS_WriteSuccessful(t *testing.T) {
	t.Parallel()
	acceptance.RunTest(
		t,
		[]acceptance.SourceCode{
			setupNFSApp(),
		},
		func(ctx context.Context, t *testing.T, kf *integration.Kf, appPath string) {
			// We only need a subset of the repo.
			manifestPath := path.Join(appPath, "assets", "pora")
			t.Parallel()
			kf.Push(ctx, "",
				"--path", manifestPath,
				"--manifest", path.Join(manifestPath, "manifest.yml"),
			)

			// There will only be this app in the space.
			info := kf.Apps(ctx)
			if l := len(info); l != 1 {
				t.Fatalf("expected there to be one app, but there were %d", l)
			}

			for name := range info {

				kf.WithProxy(ctx, name, func(addr string) {
					serviceInstanceName := "nfs-instance"
					kf.CreateVolumeService(
						ctx,
						serviceInstanceName,
						//"-c", "{\"share\":\"10.52.220.162/share\", \"capacity\":\"1Gi\"}",
						"-c", "{\"share\":\"10.126.178.2/test\", \"capacity\":\"1Gi\"}",
					)

					kf.BindService(
						ctx,
						name,
						serviceInstanceName,
						"-c", "{\"mount\":\"/mount1\", \"gid\":\"3000\", \"uid\":\"3000\"}",
					)

					serviceInstanceName2 := "nfs-instance2"
					kf.CreateVolumeService(
						ctx,
						serviceInstanceName2,
						//"-c", "{\"share\":\"10.52.220.162/share\", \"capacity\":\"1Gi\"}",
						"-c", "{\"share\":\"10.126.178.2/share\", \"capacity\":\"1Gi\"}",
					)

					kf.BindService(
						ctx,
						name,
						serviceInstanceName2,
						"-c", "{\"mount\":\"/mount2\", \"gid\":\"3000\", \"uid\":\"3000\"}",
					)

					u, err := url.Parse(addr)
					if err != nil {
						t.Fatalf("unable to parse proxy url %v", addr)
					}
					u.Path = "write"
					response, clean := integration.RetryGet(ctx, t, u.String(), 1*time.Minute, http.StatusOK)
					expectedBody :=
						`Hello Persistent World!
Append text!
`
					body, err := ioutil.ReadAll(response.Body)
					if err != nil {
						t.Fatal(err)
					}

					testutil.AssertEqual(t, "response body", expectedBody, string(body))
					clean()
				})
			}
		})
}
