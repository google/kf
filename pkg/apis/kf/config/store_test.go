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

package config

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/kf/v2/pkg/kf/testutil"
	logtesting "knative.dev/pkg/logging/testing"

	cmtesting "knative.dev/pkg/configmap/testing"
)

const sampleFlag = "enable_some_feature"

func TestStoreLoadDefault(t *testing.T) {
	store := NewDefaultConfigStore(logtesting.TestLogger(t))
	config, err := FromContext(store.ToContext(context.Background())).Defaults()
	testutil.AssertNil(t, "err", err)
	testutil.AssertEqual(t, "config", BuiltinDefaultsConfig(), config)
}

func TestStoreLoadWithContext(t *testing.T) {
	store := NewDefaultConfigStore(logtesting.TestLogger(t))
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
	_, configMap := cmtesting.ConfigMapsFromTestFile(t, DefaultsConfigName, allowedPredefinedKey...)
	// sanity check the configmap, add more assertions below when new fields
	// are added. Add 2 for the underscore templates
	testutil.AssertEqual(t, "configMap entries", len(allowedPredefinedKey)+2, len(configMap.Data))

	store.OnConfigChanged(configMap)
	config := FromContext(store.ToContext(context.Background()))

	configDefaults, err := config.Defaults()
	testutil.AssertNil(t, "err", err)
	expected, _ := NewDefaultsConfigFromConfigMap(configMap)
	if diff := cmp.Diff(expected, configDefaults); diff != "" {
		t.Errorf("Unexpected defaults config (-want, +got): %v", diff)
	}

	testutil.AssertEqual(t, "SpaceClusterDomains", []DomainTemplate{
		{Domain: "$(SPACE_NAME).prod.example.com"},
		{Domain: "$(SPACE_NAME).kf.us-east1.prod.example.com"},
		{Domain: "$(SPACE_NAME).$(CLUSTER_INGRESS_IP).nip.io"},
		{Domain: "$(SPACE_NAME)-apps.internal", GatewayName: "kf/internal-gateway"},
	}, configDefaults.SpaceClusterDomains)
	testutil.AssertEqual(t, "SpaceContainerRegistry", "gcr.io/my-project", configDefaults.SpaceContainerRegistry)
	testutil.AssertEqual(t, "SpaceBuildpacksV2", BuildpackV2List{
		{
			Name: "name-of-buildpack",
			URL:  "https://github.com/cloudfoundry/name-of-buildpack",
		},
	}, configDefaults.SpaceBuildpacksV2)

	testutil.AssertEqual(t, "SpaceStacksV2", StackV2List{
		{
			Name:  "cflinuxfs3",
			Image: "cloudfoundry/cflinuxfs3",
		},
	}, configDefaults.SpaceStacksV2)

	testutil.AssertEqual(t, "SpaceStacksV3", StackV3List{
		{
			Name:         "heroku-18",
			Description:  "The official Heroku stack based on Ubuntu 18.04",
			BuildImage:   "heroku/pack:18-build",
			RunImage:     "heroku/pack:18",
			NodeSelector: map[string]string{"kubernetes.io/os": "windows"},
		},
	}, configDefaults.SpaceStacksV3)

	testutil.AssertEqual(t, "SpaceStackDefaultToV3", true, configDefaults.SpaceDefaultToV3Stack)

	testutil.AssertEqual(t, "RouteServiceProxyImage", "ko://github.com/google/kf/v2/route-service-proxy-src", configDefaults.RouteServiceProxyImage)

	expectFeatureFlags := []struct {
		name         string       // name of the flag
		isSet        bool         // whether we expect it to exist in the configmap
		wantSetValue bool         // the value we expect in the configmap
		wantEnabled  bool         // whether we expect IsEnabled to return true
		flag         *FeatureFlag // the actual flag
	}{
		{
			name:         disableCustomBuildsFlag,
			isSet:        true,
			wantSetValue: false,
			wantEnabled:  false,
			flag:         configDefaults.FeatureFlags.DisableCustomBuildsFlag(),
		},
		{
			name:         dockerfileBuildsFlag,
			isSet:        true,
			wantSetValue: true,
			wantEnabled:  true,
			flag:         configDefaults.FeatureFlags.DockerfileBuilds(),
		},
		{
			name:         customBuildpacksFlag,
			isSet:        true,
			wantSetValue: true,
			wantEnabled:  true,
			flag:         configDefaults.FeatureFlags.CustomBuildpacks(),
		},
		{
			name:         customStacksFlag,
			isSet:        true,
			wantSetValue: true,
			wantEnabled:  true,
			flag:         configDefaults.FeatureFlags.CustomStacks(),
		},
		{
			name:         sampleFlag,
			isSet:        true,
			wantSetValue: true,
			wantEnabled:  true,
			flag:         nil,
		},
	}

	expectedFlags := FeatureFlagToggles{}
	for _, flag := range expectFeatureFlags {
		if flag.isSet {
			expectedFlags[flag.name] = flag.wantSetValue
		}
	}

	testutil.AssertEqual(t, "FeatureFlags", expectedFlags, configDefaults.FeatureFlags)
	testutil.AssertEqual(t, "Number of feature flags set", len(expectedFlags), len(configDefaults.FeatureFlags))

	for _, flag := range expectFeatureFlags {
		t.Run("flag:"+flag.name, func(t *testing.T) {
			if flag.flag == nil {
				return
			}

			testutil.AssertEqual(t, "IsEnabled", flag.wantEnabled, flag.flag.IsEnabled())
			testutil.AssertEqual(t, "IsDisabled", !flag.wantEnabled, flag.flag.IsDisabled())
		})
	}
}
