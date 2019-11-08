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
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/google/kf/pkg/reconciler/source/config"
	corev1 "k8s.io/api/core/v1"
)

func TestNewSecretsConfigFromConfigMap(t *testing.T) {
	t.Parallel()

	sc, err := config.NewSecretsConfigFromConfigMap(&corev1.ConfigMap{
		Data: map[string]string{
			config.BuildImagePushSecretKey: "some-value-0, some-value-1",
		},
	})
	testutil.AssertNil(t, "err", err)
	testutil.AssertEqual(t, "BuildImagePushSecret Count", 2, len(sc.BuildImagePushSecrets))
	testutil.AssertEqual(t, "BuildImagePushSecrets[0]", "some-value-0", sc.BuildImagePushSecrets[0].Name)
	testutil.AssertEqual(t, "BuildImagePushSecrets[1]", "some-value-1", sc.BuildImagePushSecrets[1].Name)
}
