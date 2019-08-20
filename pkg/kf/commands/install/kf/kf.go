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
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	. "github.com/google/kf/pkg/kf/commands/install/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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
		err := wait.ExponentialBackoff(
			wait.Backoff{
				Duration: time.Second,
				Steps:    10,
				Factor:   1,
			}, func() (bool, error) {
				Logf(ctx, "install "+yaml.name)
				if _, err := Kubectl(
					ctx,
					"apply",
					"--filename",
					yaml.yaml,
				); err != nil {
					Logf(ctx, "failed to install %s... Retrying", yaml.name)
					// Don't return the error. This will cause the
					// ExponentialBackoff to stop.
					return false, nil
				}

				return true, nil
			})

		if err != nil {
			return err
		}
	}

	// Wait for controller and webhook deployments to be ready
	if err := waitForKfDeployments(ctx); err != nil {
		return err
	}

	// Setup kf space
	if err := SetupSpace(ctx, containerRegistry); err != nil {
		return err
	}

	return nil
}

func waitForKfDeployments(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	for _, deploymentName := range []string{"controller", "webhook"} {
		Logf(ctx, "waiting for %s deployment to be available...", deploymentName)
		err := wait.ExponentialBackoff(
			wait.Backoff{
				Duration: 5 * time.Second,
				Steps:    10,
				Factor:   1.5,
			}, func() (bool, error) {
				output, err := Kubectl(
					ctx,
					"get",
					"deployments",
					deploymentName,
					"--namespace", v1alpha1.KfNamespace,
					"--output=json",
				)
				if err != nil {
					return false, err
				}
				deployment := appsv1.Deployment{}
				if err := json.NewDecoder(strings.NewReader(strings.Join(output, "\n"))).Decode(&deployment); err != nil {
					return false, err
				}

				for _, cond := range deployment.Status.Conditions {
					if cond.Type == appsv1.DeploymentAvailable {
						return cond.Status == corev1.ConditionTrue, nil
					}
				}
				return false, nil
			},
		)

		if err != nil {
			return err
		}
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
