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

package kf

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	. "github.com/google/kf/pkg/kf/commands/install/util"
)

const (
	// KnativeBuildYAML holds the knative build yaml release URL
	KnativeBuildYAML = "https://github.com/knative/build/releases/download/v0.6.0/build.yaml"
	// KfNightlyBuildYAML holds the kf nightly build release URL
	KfNightlyBuildYAML = "https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/releases/release-latest.yaml"
)

// Install installs the necessary kf components and allows the user to create
// a space.
func Install(ctx context.Context, containerRegistry string) error {
	ctx = SetLogPrefix(ctx, "Install kf")

	// Install Service Catalog
	if err := installServiceCatalog(ctx); err != nil {
		return err
	}

	// kubectl apply various yaml files{
	for _, yaml := range []struct {
		name string
		yaml string
	}{
		{name: "Knative Build", yaml: KnativeBuildYAML},
		{name: "kf", yaml: KfNightlyBuildYAML},
	} {

		Logf(ctx, "install "+yaml.name)
		if _, err := Kubectl(
			ctx,
			"apply",
			"--filename",
			yaml.yaml,
		); err != nil {
			return err
		}
	}

	// Setup kf space
	if err := SetupSpace(ctx, containerRegistry); err != nil {
		return err
	}

	return nil
}

func installServiceCatalog(ctx context.Context) error {
	ctx = SetLogPrefix(ctx, "Service Catalog")
	Logf(ctx, "installing Service Catalog")
	Logf(ctx, "downloading service catalog templates")
	tempDir, err := ioutil.TempDir("", "kf-service-catalog")
	if err != nil {
		return err
	}
	defer func() {
		Logf(ctx, "cleaning up %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	tmpKfPath := path.Join(tempDir, "kf")

	if _, err := Git(
		ctx,
		"clone",
		"https://github.com/google/kf",
		tmpKfPath,
	); err != nil {
		return err
	}

	Logf(ctx, "applying templates")
	if _, err := Kubectl(
		ctx,
		"apply",
		"-R",
		"--filename", path.Join(tmpKfPath, "third_party/service-catalog/manifests/catalog/templates"),
	); err != nil {
		return err
	}

	return nil
}
