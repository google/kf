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
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	. "github.com/google/kf/pkg/kf/commands/install/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// KnativeBuildYAML holds the knative build yaml release URL
	KnativeBuildYAML = "https://github.com/knative/build/releases/download/v0.7.0/build.yaml"
	// KfNightlyBuildYAML holds the kf nightly build release URL
	KfNightlyBuildYAML = "https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/release.yaml"
)

// Install installs the necessary kf components and allows the user to create
// a space.
func Install(ctx context.Context, containerRegistry string) error {
	ctx = SetLogPrefix(ctx, "Install kf")

	// Install Service Catalog
	if err := installServiceCatalog(ctx); err != nil {
		return err
	}

	// Select Kf Version
	kfNames, kfReleases, err := fetchReleases(ctx)
	if err != nil {
		return err
	}
	idx, _, err := SelectPrompt(
		ctx,
		"Select Kf Version",
		kfNames...,
	)
	if err != nil {
		return err
	}

	kfRelease := kfReleases[idx]

	// kubectl apply various yaml files{
	for _, yaml := range []struct {
		name string
		yaml string
	}{
		{name: "Knative Build", yaml: KnativeBuildYAML},
		{name: "kf", yaml: kfRelease},
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

// fetchReleases returns a list of releases YAML files from github and the
// nightly. The nightly will be the first entry. The two slices are related.
func fetchReleases(ctx context.Context) (names, addrs []string, err error) {
	Logf(ctx, "fetching Kf releases...")
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(ctx, "google", "kf", nil)
	if err != nil {
		return nil, nil, err
	}

	var vs versions

	for _, r := range releases {
		if r.Name == nil || r.TagName == nil {
			continue
		}

		var assetURL string
		for _, a := range r.Assets {
			if a.Name != nil && a.BrowserDownloadURL != nil && *a.Name == "release.yaml" {
				assetURL = *a.BrowserDownloadURL
				break
			}
		}

		if assetURL == "" {
			continue
		}

		v, err := semver.ParseTolerant(*r.TagName)
		if err != nil {
			Logf(ctx, "invalid semver tag %q: %s", *r.TagName, err)
			continue
		}

		vs = append(vs, version{
			name:   *r.Name,
			semver: v,
			addr:   assetURL,
		})
	}

	sort.Sort(vs)

	names = []string{"nightly"}
	addrs = []string{KfNightlyBuildYAML}

	for _, v := range vs {
		names = append(names, v.name)
		addrs = append(addrs, v.addr)
	}

	return names, addrs, nil
}

type version struct {
	semver semver.Version
	name   string
	addr   string
}

type versions []version

// Len is the number of elements in the collection.
func (v versions) Len() int {
	return len(v)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (v versions) Less(i int, j int) bool {
	return v[i].semver.LT(v[j].semver)
}

// Swap swaps the elements with indexes i and j.
func (v versions) Swap(i int, j int) {
	tmp := v[i]
	v[i] = v[j]
	v[j] = tmp
}
