/*
Copyright 2020 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"kf-operator/pkg/release/commandtools"
	"kf-operator/pkg/release/filetools"
	"kf-operator/pkg/release/stringtools"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// gcsHTTP - HTTP endpoint for Google Cloud Storage
	gcsHTTP = "https://storage.googleapis.com"
	// tarTmpFile - Temporary tarball file name
	tarTmpFile = "vendor.tar.gz"
	// gcrPathRegExp - GCR Path Extended Regular Expression
	gcrPathRegExp = `((us|eu|asia).)?gcr.io/[^/]*`
)

// INTERNAL_BIN - Internal images with vendored source tarball
var internalBinList = [...]string{"manager"}

func getEnvironmentVariable(variable string, mandatory bool) string {
	value := os.Getenv(variable)
	if value == "" && mandatory {
		log.Fatalf("%q environment variable is empty", variable)
	}
	return value
}

func generateImageLinks(images []string, linkPathTemplate string, originPath string) {
	for _, image := range images {
		filetools.CreateFileLinks(
			fmt.Sprintf(originPath, tarTmpFile),
			fmt.Sprintf(linkPathTemplate, image))
	}
}

func exportEnvironmentVariables(gcrPath, localAddonYaml, addonYaml string) {
	// Resolve the config.
	os.Setenv("KO_DOCKER_REPO", strings.TrimSuffix(gcrPath, "/"))

	// Export the LOCAL_ADDON_YAML so that other script, such as upgrade-test's build script, can use it to locate the local yaml.
	// https://source.corp.google.com/cloud-run-on-gke/test/prow/upgrade-testing/build-infra/build.sh;l=54-56;rcl=c7f16fd03f9666422724cbbbc0522986ce70fa37
	os.Setenv("LOCAL_ADDON_YAML", localAddonYaml)

	os.Setenv("ADDON_YAML", addonYaml)
}

func verifyYamlFilesContent(localCrpYaml string, crpTemplatedRepoString string, localCrpNonTemplatedYaml string) {
	log.Printf("============= Verifying Yaml content =============")
	// Check to confirm the produced templated yaml has at least 1 image key line (and cache result in variable)
	localCrpYamlLines := stringtools.ReadLines(localCrpYaml)
	localCrpYamlValidLines := make([]string, 0)
	imageRe := regexp.MustCompilePOSIX("[Ii]mage: ")
	for index, line := range localCrpYamlLines {
		if imageRe.MatchString(line) {
			localCrpYamlValidLines = append(localCrpYamlValidLines, fmt.Sprint(index)+":"+line)
		}
	}

	if len(localCrpYamlValidLines) == 0 {
		log.Fatalf("Found zero image keys in %q. Failing...", localCrpYaml)
	}

	for _, line := range localCrpYamlValidLines {
		if !strings.Contains(line, crpTemplatedRepoString) {
			log.Fatalf("Expected all images to be templated in %q, but some still exist:\n%s",
				localCrpYaml, line)
		}
	}

	localCrpNonTemplatedYamlLines := stringtools.ReadLines(localCrpNonTemplatedYaml)
	gcrPathRegExpRe := regexp.MustCompilePOSIX(gcrPathRegExp)
	for _, line := range localCrpNonTemplatedYamlLines {
		if gcrPathRegExpRe.MatchString(line) || imageRe.MatchString(line) {
			log.Fatalf("Expected no images in non templated yaml, but some exist:\n%v",
				localCrpNonTemplatedYamlLines)
		}
	}
}

func generateReleases(development bool, gcrPath, gcsPath, version string) {
	// Location of the config folder where the yaml files are located in the repository
	configPath := "config"
	tag := fmt.Sprintf("v%s", version)
	// Location of relevant files. The names are purely for humans to simplify development.
	localProdAddonYaml := filetools.CreateTempFile("serverless-*.yaml")

	// Upload the add-on manifest.
	remoteName := func(name string) string {
		return fmt.Sprintf("%s/cloudrun/%s-%s.yaml", gcsPath, name, version)
	}
	prodAddonYaml := remoteName("serverless-operator")

	// Export environment variables for outer scope scripts
	// TODO: This should be removed in favor of arguments and returned values
	exportEnvironmentVariables(gcrPath, localProdAddonYaml, prodAddonYaml)

	koResolver := commandtools.NewKoResolver()

	if development {
		// Meant for development/test installation.
		// Use GCR staging path for Istio images, this is to ensure when the Istio
		// images are under embargo, we can still apply the yaml.
		log.Printf("============= Processing Development/Test Yaml =============")
		koResolver.Resolve(
			configPath,
			tag,
			localProdAddonYaml,
			version)

		log.Printf("============= Uploading Yaml files =============")
		commandtools.RunGsCpUtil(
			localProdAddonYaml,
			"gs://"+prodAddonYaml)

		log.Printf("All yaml files generated and uploaded successfully")
	} else {
		//// PROD YAML
		// Meant for production. Uses production GCR path [gke.gcr.io], with
		// addonmanager label
		log.Printf("============= Processing Prod Yaml =============")

		operatorYamlDirectory := filetools.CreateTempDirectory("")
		filetools.Copy(filepath.Join(configPath, "*.yaml"), operatorYamlDirectory)
		// Remove CRDs as they are installed separately from the hub controller.
		filetools.RemoveAll(filepath.Join(operatorYamlDirectory, "200-cloudrun-crd.yaml"))
		filetools.RemoveAll(filepath.Join(operatorYamlDirectory, "200-kfsystem-crd.yaml"))
		// Remove CR as it is installed by the user to the cluster.
		filetools.RemoveAll(filepath.Join(operatorYamlDirectory, "999-cloud-run-cr.yaml"))

		koResolver.Resolve(
			operatorYamlDirectory,
			tag,
			localProdAddonYaml,
			version)

		log.Printf("Uploading to %s", prodAddonYaml)
		commandtools.RunGsCpUtil(
			localProdAddonYaml,
			"gs://"+prodAddonYaml)
	}
}

func cleanUpResources() {
	// Clean up internal images symlinks
	for _, bin := range internalBinList {
		filetools.RemoveAll(fmt.Sprintf("cmd/%s/kodata/vendor.tar.gz", bin))
	}

	// Clean up tarball, to avoid checking in the tarball
	filetools.RemoveAll(tarTmpFile)
}

func main() {
	o := options{}
	o.addFlags()
	flag.Parse()
	if err := o.mustProcessFlags(); err != nil {
		log.Fatal(err)
	}

	// Make sure we're not in the GOPATH as that causes ko and go to have some serious caching issues.
	filetools.MakeTempModuleOutsideGoPath("./")

	// Create tar file for vendored source, needed for license compliance
	filetools.CreateTarballFile(tarTmpFile, "vendor/")

	// Add symlink for internal images with vendored source tarball
	generateImageLinks(
		internalBinList[:],
		"cmd/%s/kodata/vendor.tar.gz",
		"../../../%s",
	)

	// Generate releases
	generateReleases(o.development, o.gcrPath, o.gcsPath, o.version)

	// Clean up
	cleanUpResources()
}
