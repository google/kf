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
	"context"
	"encoding/json"
	"kf-operator/pkg/apis/kfsystem/kf"

	mf "github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// FeatureFlagsAnnotation holds a map of each feature flag to a bool indicating whether the feature is enabled or not.
	FeatureFlagsAnnotation = "kf.dev/feature-flags"
)

// AddFeatureFlags add feature flags annotation to kf namespace
func AddFeatureFlags(context context.Context, featureFlags kf.FeatureFlagToggles) mf.Transformer {
	// Consolidate all feature flags into one annotation in Kf namespace
	return func(u *unstructured.Unstructured) error {
		if len(featureFlags) == 0 {
			return nil
		}

		if u.GetKind() == "Namespace" && u.GetName() == "kf" {
			marshaledMap, err := json.Marshal(featureFlags)
			if err != nil {
				return err
			}
			u.SetAnnotations(map[string]string{
				FeatureFlagsAnnotation: string(marshaledMap),
			})
		}

		return nil
	}
}
