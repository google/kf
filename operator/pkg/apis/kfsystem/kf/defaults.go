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

package kf

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultsConfigName is the name of the defaults configmap.
	DefaultsConfigName               = "config-defaults"
	spaceContainerRegistryKey        = "spaceContainerRegistry"
	spaceClusterDomainsKey           = "spaceClusterDomains"
	spaceBuildpacksV2Key             = "spaceBuildpacksV2"
	spaceStacksV2Key                 = "spaceStacksV2"
	spaceStacksV3Key                 = "spaceStacksV3"
	spaceDefaultToV3StackKey         = "spaceDefaultToV3Stack"
	routeServiceProxyImageKey        = "routeServiceProxyImage"
	featureFlagsKey                  = "featureFlags"
	buildDisableIstioSidecarKey      = "buildDisableIstioSidecar"
	buildPodResourcesKey             = "buildPodResources"
	buildRetentionCountKey           = "buildRetentionCount"
	taskRetentionCountKey            = "taskRetentionCount"
	buildTimeoutKey                  = "buildTimeout"
	buildNodeSelectorsKey            = "buildNodeSelectors"
	appCPUPerGBOfRAMKey              = "appCPUPerGBOfRAM"
	appCPUMinKey                     = "appCPUMin"
	appDisableStartCommandLookupKey  = "appDisableStartCommandLookup"
	progressDeadlineSecondsKey       = "progressDeadlineSeconds"
	terminationGracePeriodSecondsKey = "terminationGracePeriodSeconds"
	routeTrackVirtualServiceKey      = "routeTrackVirtualService"
	taskDefaultTimeoutMinutesKey     = "taskDefaultTimeoutMinutes"

	// Images used for build purposes

	buildKanikoExecutorImageKey   = "buildKanikoExecutorImage"
	buildInfoImageKey             = "buildInfoImage"
	buildTokenDownloadImageKey    = "buildTokenDownloadImage"
	buildHelpersImageKey          = "buildHelpersImage"
	buildpacksV2LifecycleImageKey = "buildpacksV2LifecycleImage"
	nopImageKey                   = "nopImage"
)

// DefaultsConfig contains the configuration for defaults.
type DefaultsConfig struct {
	// NOTE: The JSON tags are included here for correctly formatting the config
	// for .golden files.

	// SpaceContainerRegistry is the default container registry to assign
	// to spaces.
	SpaceContainerRegistry string `json:"spaceContainerRegistry,omitempty"`

	// SpaceClusterDomains is the default domain for spaces. Apps in the
	// space will get the name <app>.<space>.<defaultClusterDomain> by default.
	SpaceClusterDomains []DomainTemplate `json:"spaceClusterDomains,omitempty"`

	// SpaceBuildpacksV2 contains a sorted list of buildpacks to be used for CF
	// compatible builds.
	SpaceBuildpacksV2 BuildpackV2List `json:"spaceBuildpacksV2,omitempty"`

	// SpaceStacksV2 contains a list of stacks that can be used for CF compatible
	// builds.
	SpaceStacksV2 StackV2List `json:"spaceStacksV2,omitempty"`

	// SpaceStacksV3 contains a list of stacks that can be used for CF compatible
	// builds.
	SpaceStacksV3 StackV3List `json:"spaceStacksV3,omitempty"`

	// SpaceDefaultToV3Stack determines whether v3 stacks should be used by
	// default over v2 stacks.
	SpaceDefaultToV3Stack bool `json:"spaceDefaultToV3Stack,omitempty"`

	// RouteServiceProxyImage is the image URL for the Kf route service proxy deployment.
	RouteServiceProxyImage string `json:"routeServiceProxyImage,omitempty"`

	BuildKanikoExecutorImage string `json:"buildKanikoExecutorImage,omitempty"`
	BuildInfoImage           string `json:"buildInfoImage,omitempty"`
	BuildTokenDownloadImage  string `json:"buildTokenDownloadImage,omitempty"`
	BuildHelpersImage        string `json:"buildHelpersImage,omitempty"`
	NopImage                 string `json:"nopImage,omitempty"`

	// BuildpacksV2LifecycleImage is the image URL for the V2 buildpack
	// lifecycle binaries. It is expected to contain the `launcher` and
	// `builder` binaries AND to self extract those binaries into /workspace.
	BuildpacksV2LifecycleImage string `json:"buildpacksV2LifecycleImage,omitempty"`

	// BuildDisableIstioSidecar when set to true, will prevent the Istio
	// sidecar from being attached to the Build pods.
	BuildDisableIstioSidecar bool `json:"buildDisableIstioSidecar,omitempty"`

	// BuildPodResources sets the Build pod resources field.
	// NOTE: This is only applicable for built-in Tasks. For V2 builds, this
	// will be set on two steps and one for V3 and Dockerfiles. This implies
	// for a V2 build, the required Pod size will be the limit doubled.  For
	// example, if the memory limit is 1Gi, then the pod will require 2Gi.
	BuildPodResources *corev1.ResourceRequirements `json:"buildPodResources,omitempty"`

	// BuildRetentionCount is the number of completed Builds each App will
	// keep before garbage collecting.
	BuildRetentionCount *uint `json:"buildRetentionCount,omitempty"`

	// TaskRetentionCount is the number of completed Tasks each App will
	// keep before garbage collecting.
	TaskRetentionCount *uint `json:"taskRetentionCount,omitempty"`

	// BuildTimeout is the timeout value to set at the underlying TaskRun of Build, it is the time duration
	// to fail a build that takes too long to complete, valid values are specified in time unit, e.g. "1h", "30m", "60s"
	BuildTimeout string `json:"buildTimeout,omitempty"`

	// BuildNodeSelectors is the node selectors for the build node pool.
	// When it's specified, the Kf build pods are only assigned to the nodes that have labels matching the selectors.
	BuildNodeSelectors BuildNodeSelectors `json:"buildNodeSelectors,omitempty"`

	// FeatureFlags is a map of feature names to their status (enabled/disabled) represented as a bool.
	FeatureFlags FeatureFlagToggles `json:"featureFlags,omitempty"`

	// Default amount of CPU to assign an app per GB of RAM.
	AppCPUPerGBOfRAM *resource.Quantity `json:"appCPUPerGBOfRAM,omitempty"`

	// Minimum amount of CPU to assign an app.
	AppCPUMin *resource.Quantity `json:"appCPUMin,omitempty"`

	// AppDisableStartCommandLookup disables the App reconciler from looking
	// up the start command for Apps which requires fetching the container
	// configuration for every App.
	AppDisableStartCommandLookup bool `json:"appDisableStartCommandLookup,omitempty"`

	// ProgressDeadlineSeconds contains the maximum time in seconds for a deployment to make progress before it
	// is considered to be failed.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty"`

	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// RouteTrackVirtualService when set to true, will update Route status with VirtualService conditions.
	RouteTrackVirtualService bool `json:"routeTrackVirtualService,omitempty"`

	// TaskDefaultTimeoutMinutes sets the cluster-wide timeout for tasks.
	// If the value is null, the timeout is inherited from Tekton.
	// If the value is <= 0, then an infinite timeout is set.
	TaskDefaultTimeoutMinutes *int32 `json:"taskDefaultTimeoutMinutes,omitempty"`
}

// BuiltinDefaultsConfig creates a defaults configuration with default values.
func BuiltinDefaultsConfig() *DefaultsConfig {
	// CPU isn't defaulted in Cloud Foundry so we assume apps are I/O bound
	// and roughly 10 should run on a single core machine.
	defaultCPUPerGBOfRAM := resource.MustParse("100m")
	defaultAppCPUMin := resource.MustParse("100m")

	return &DefaultsConfig{
		// Default to v2 stacks like CF users expect.
		SpaceDefaultToV3Stack: false,

		// Default to what Kf did prior to scaling CPU by GB/RAM.
		AppCPUPerGBOfRAM: &defaultCPUPerGBOfRAM,

		// Prevent an app from being starved even if it doesn't need much RAM.
		AppCPUMin: &defaultAppCPUMin,
	}
}

// NewDefaultsConfigFromConfigMap creates a DefaultsConfig from the supplied
// ConfigMap with defaults provided by DefaultSpaceConfig for any blank fields.
func NewDefaultsConfigFromConfigMap(configMap *corev1.ConfigMap) (*DefaultsConfig, error) {
	defaultsConfig := BuiltinDefaultsConfig()
	// Populate string values
	for k, dest := range defaultsConfig.getStringValues() {
		if setValue, ok := configMap.Data[k]; ok {
			*dest = setValue
		}
	}

	// Populate JSON/YAML values
	for k, dest := range defaultsConfig.getInterfaceValues(false) {
		if setValue, ok := configMap.Data[k]; ok {
			if err := yaml.Unmarshal([]byte(setValue), dest); err != nil {
				return nil, err
			}
		}
	}

	return defaultsConfig, nil
}

// PatchConfigMap merges values from a DefaultsConfig onto a v1.ConfigMap. This updates the keys that are part of DefaultsConfig
// while leaving the rest of the original v1.ConfigMap (such as _example) untouched.
func (defaultsConfig *DefaultsConfig) PatchConfigMap(cm *corev1.ConfigMap) error {
	// Patch string values
	for k, setValue := range defaultsConfig.getStringValues() {
		if len(*setValue) == 0 {
			continue
		}
		cm.Data[k] = *setValue
	}

	// Patch JSON/YAML values
	for k, setValue := range defaultsConfig.getInterfaceValues(true) {
		if setValue == nil {
			continue
		}

		marshaled, err := yaml.Marshal(setValue)
		if err != nil {
			return err
		}
		// Don't rewrite values that are already equivalent. This also avoids adding
		// formatting diffs caused by the YAML marshaller.
		equalYAML, err := IsYAMLEqual(cm.Data[k], string(marshaled))
		if err != nil {
			return err
		}
		if !equalYAML {
			cm.Data[k] = string(marshaled)
		}
	}

	return nil
}

// getStringValues returns a map of the key/value pairs on a DefaultsConfig that are string values.
func (defaultsConfig *DefaultsConfig) getStringValues() map[string]*string {
	return map[string]*string{
		spaceContainerRegistryKey:     &defaultsConfig.SpaceContainerRegistry,
		routeServiceProxyImageKey:     &defaultsConfig.RouteServiceProxyImage,
		buildKanikoExecutorImageKey:   &defaultsConfig.BuildKanikoExecutorImage,
		buildInfoImageKey:             &defaultsConfig.BuildInfoImage,
		buildTokenDownloadImageKey:    &defaultsConfig.BuildTokenDownloadImage,
		buildHelpersImageKey:          &defaultsConfig.BuildHelpersImage,
		buildpacksV2LifecycleImageKey: &defaultsConfig.BuildpacksV2LifecycleImage,
		buildTimeoutKey:               &defaultsConfig.BuildTimeout,
		nopImageKey:                   &defaultsConfig.NopImage,
	}
}

// getInterfaceValues returns a map of the key/value pairs on a DefaultsConfig that are JSON/YAML encoded values.
func (defaultsConfig *DefaultsConfig) getInterfaceValues(leaveEmpty bool) map[string]interface{} {
	m := map[string]interface{}{
		spaceDefaultToV3StackKey:         &defaultsConfig.SpaceDefaultToV3Stack,
		buildDisableIstioSidecarKey:      &defaultsConfig.BuildDisableIstioSidecar,
		buildPodResourcesKey:             &defaultsConfig.BuildPodResources,
		buildRetentionCountKey:           &defaultsConfig.BuildRetentionCount,
		taskRetentionCountKey:            &defaultsConfig.TaskRetentionCount,
		buildNodeSelectorsKey:            &defaultsConfig.BuildNodeSelectors,
		appCPUPerGBOfRAMKey:              &defaultsConfig.AppCPUPerGBOfRAM,
		appCPUMinKey:                     &defaultsConfig.AppCPUMin,
		appDisableStartCommandLookupKey:  &defaultsConfig.AppDisableStartCommandLookup,
		progressDeadlineSecondsKey:       &defaultsConfig.ProgressDeadlineSeconds,
		terminationGracePeriodSecondsKey: &defaultsConfig.TerminationGracePeriodSeconds,
		routeTrackVirtualServiceKey:      &defaultsConfig.RouteTrackVirtualService,
		taskDefaultTimeoutMinutesKey:     &defaultsConfig.TaskDefaultTimeoutMinutes,
	}

	if !leaveEmpty {
		m[spaceBuildpacksV2Key] = &defaultsConfig.SpaceBuildpacksV2
		m[spaceClusterDomainsKey] = &defaultsConfig.SpaceClusterDomains
		m[spaceStacksV2Key] = &defaultsConfig.SpaceStacksV2
		m[spaceStacksV3Key] = &defaultsConfig.SpaceStacksV3
		m[featureFlagsKey] = &defaultsConfig.FeatureFlags
	}

	if len(defaultsConfig.SpaceBuildpacksV2) > 0 {
		m[spaceBuildpacksV2Key] = &defaultsConfig.SpaceBuildpacksV2
	}
	if len(defaultsConfig.SpaceClusterDomains) > 0 {
		m[spaceClusterDomainsKey] = &defaultsConfig.SpaceClusterDomains
	}
	if len(defaultsConfig.SpaceStacksV2) > 0 {
		m[spaceStacksV2Key] = &defaultsConfig.SpaceStacksV2
	}
	if len(defaultsConfig.SpaceStacksV3) > 0 {
		m[spaceStacksV3Key] = &defaultsConfig.SpaceStacksV3
	}
	if len(defaultsConfig.FeatureFlags) > 0 {
		m[featureFlagsKey] = &defaultsConfig.FeatureFlags
	}

	return m
}

// IsYAMLEqual returns whether the expexted YAML matches the actual YAML.
func IsYAMLEqual(expected, actual string) (bool, error) {
	var em, am interface{}
	err := yaml.Unmarshal([]byte(expected), &em)
	if err != nil {
		return false, err
	}
	err = yaml.Unmarshal([]byte(actual), &am)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(em, am), nil
}
