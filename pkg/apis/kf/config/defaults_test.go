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

package config

import (
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	cmtesting "knative.dev/pkg/configmap/testing"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultsConfigTestName is the name of the defaults configmap used for testing
	// It is expected to update this testfile when ever new fields are added to the default config map
	DefaultsConfigTestName = "config-defaults"
)

func TestPatchConfigMap(t *testing.T) {
	allowedPredefinedKey := []string{
		spaceBuildpacksV2Key,
		spaceStacksV2Key,
		spaceStacksV3Key,
		spaceDefaultToV3StackKey,
		spaceClusterDomainsKey,
		routeServiceProxyImageKey,
		buildKanikoExecutorImageKey,
		buildInfoImageKey,
		buildTokenDownloadImageKey,
		buildHelpersImageKey,
		buildpacksV2LifecycleImageKey,
		buildDisableIstioSidecarKey,
		buildPodResourcesKey,
		featureFlagsKey,
		nopImageKey,
		appCPUMinKey,
		appCPUPerGBOfRAMKey,
		appDisableStartCommandLookupKey,
		progressDeadlineSecondsKey,
		terminationGracePeriodSecondsKey,
		routeTrackVirtualServiceKey,
		taskDefaultTimeoutMinutesKey,
	}
	_, configMap := cmtesting.ConfigMapsFromTestFile(t, DefaultsConfigTestName, allowedPredefinedKey...)
	// sanity check the configmap, add more assertions below when new fields
	// are added. Add 2 for the underscore templates
	testutil.AssertEqual(t, "configMap entries", len(allowedPredefinedKey)+2, len(configMap.Data))

	origConfigMap := configMap.DeepCopy()
	defaultsConfig, err := NewDefaultsConfigFromConfigMap(configMap)
	testutil.AssertNil(t, "err", err)

	// Patch values from DefaultsConfig onto the original ConfigMap.
	// Check that existing values outside of DefaultsConfig are not erased.
	// Since DefaultsConfig has not been modified, the resulting ConfigMap should be the same.
	err = defaultsConfig.PatchConfigMap(configMap)
	testutil.AssertNil(t, "err", err)
	testutil.AssertEqual(t, "ConfigMap", origConfigMap, configMap)

	// Delete the slices to ensure the patch doesn't use the empty values.
	defaultsConfig.SpaceStacksV2 = nil
	defaultsConfig.SpaceStacksV3 = nil
	defaultsConfig.SpaceBuildpacksV2 = nil
	defaultsConfig.SpaceClusterDomains = nil

	// Delete a string and ensure the patch doesn't use the empty value.
	defaultsConfig.BuildKanikoExecutorImage = ""

	// Change a feature flag value in DefaultsConfig, then patch the original ConfigMap with this change.
	ffToChange := "enable_some_feature"
	defaultsConfig.FeatureFlags[ffToChange] = false
	err = defaultsConfig.PatchConfigMap(configMap)
	testutil.AssertNil(t, "err", err)
	for k := range origConfigMap.Data {
		if k != featureFlagsKey {
			testutil.AssertEqual(t, "ConfigMap values", origConfigMap.Data[k], configMap.Data[k])
		} else {
			// Check that only one feature flag value was changed
			origFeatureFlags := FeatureFlagToggles{}
			newFeatureFlags := FeatureFlagToggles{}
			err = yaml.Unmarshal([]byte(configMap.Data[k]), &origFeatureFlags)
			testutil.AssertNil(t, "err", err)

			err = yaml.Unmarshal([]byte(configMap.Data[k]), &newFeatureFlags)
			testutil.AssertNil(t, "err", err)
			for ffName := range origFeatureFlags {
				if ffName != ffToChange {
					testutil.AssertEqual(t, "Feature flag value", origFeatureFlags[ffName], newFeatureFlags[ffName])
				} else {
					testutil.AssertEqual(t, "Updated FF value", false, newFeatureFlags[ffName])
				}
			}
		}
	}
}

func ExampleIsYAMLEqual() {
	spaceStacksOrigYAML := `- name: aaa
  image: aaa/image
  description: aaa stack
- name: bbb
  image: bbb/image
  description: bbb stack
`
	spaceStacksAlphabetizedYAML := `- description: aaa stack
  image: aaa/image
  name: aaa
- description: bbb stack
  image: bbb/image
  name: bbb
`

	fmt.Println("Orig SpaceStacksV2 YAML:")
	fmt.Println(spaceStacksOrigYAML)
	fmt.Println("Reformatted (alphabetized) SpaceStacksV2 YAML:")
	fmt.Println(spaceStacksAlphabetizedYAML)
	equalYAML, err := IsYAMLEqual(spaceStacksOrigYAML, spaceStacksAlphabetizedYAML)
	fmt.Println("IsYAMLEqual:", equalYAML)
	fmt.Println("Error:", err)

	// Output: Orig SpaceStacksV2 YAML:
	// - name: aaa
	//   image: aaa/image
	//   description: aaa stack
	// - name: bbb
	//   image: bbb/image
	//   description: bbb stack
	//
	// Reformatted (alphabetized) SpaceStacksV2 YAML:
	// - description: aaa stack
	//   image: aaa/image
	//   name: aaa
	// - description: bbb stack
	//   image: bbb/image
	//   name: bbb
	//
	// IsYAMLEqual: true
	// Error: <nil>
}
