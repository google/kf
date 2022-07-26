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

package dockerutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExampleReadConfig() {
	cfg, err := ReadConfig(filepath.Join("testdata", "credhelpers"))
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.CredentialHelpers)

	// Output: map[asia.gcr.io:gcloud eu.gcr.io:gcloud gcr.io:gcloud us.gcr.io:gcloud]
}

func ExampleDescribeConfig_credHelpers() {
	cfg, err := ReadConfig(filepath.Join("testdata", "credhelpers"))
	if err != nil {
		panic(err)
	}

	DescribeConfig(os.Stdout, cfg)

	// Output: Docker config:
	//   Auth:
	//     <none>
	//   Credential helpers:
	//     Registry     Helper
	//     asia.gcr.io  gcloud
	//     eu.gcr.io    gcloud
	//     gcr.io       gcloud
	//     us.gcr.io    gcloud
}

func ExampleDescribeConfig_customAuth() {
	cfg, err := ReadConfig(filepath.Join("testdata", "customauth"))
	if err != nil {
		panic(err)
	}

	DescribeConfig(os.Stdout, cfg)

	// Output: Docker config:
	//   Auth:
	//     Registry        Username   Email
	//     https://gcr.io  _json_key  not@val.id
	//   Credential helpers:
	//     <none>
}
