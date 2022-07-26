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

package config_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	kntesting "knative.dev/pkg/configmap/testing"
	logtesting "knative.dev/pkg/logging/testing"
)

func TestStoreLoadWithContext(t *testing.T) {
	store := config.NewSecretsConfigStore(logtesting.TestLogger(t))
	_, secretsConfig := kntesting.ConfigMapsFromTestFile(t, config.SecretsConfigName)
	store.OnConfigChanged(secretsConfig)
	cfg := config.FromContext(store.ToContext(context.Background()))
	actualSecrets, err := cfg.Secrets()
	testutil.AssertNil(t, "err", err)

	expected, _ := config.NewSecretsConfigFromConfigMap(secretsConfig)
	if diff := cmp.Diff(expected, actualSecrets); diff != "" {
		t.Errorf("Unexpected secrets config (-want, +got): %v", diff)
	}

	testutil.AssertEqual(t, "BuildImagePushSecrets", "test-build-image-push-secret", expected.BuildImagePushSecrets[0].Name)
}
