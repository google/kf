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

package activeoperand

import (
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/configmap"
	ktesting "knative.dev/pkg/reconciler/testing"

	"kf-operator/pkg/transformer"

	// Register the fakes for any informers and clients our controller accesses.
	_ "kf-operator/pkg/client/injection/client/fake"
	_ "kf-operator/pkg/client/injection/informers/operand/v1alpha1/activeoperand/fake"
	_ "kf-operator/pkg/client/injection/informers/operand/v1alpha1/clusteractiveoperand/fake"
	_ "kf-operator/pkg/client/injection/informers/operator/v1alpha1/cloudrun/fake"
	_ "kf-operator/pkg/operand/injection/dynamichelper/fake"

	_ "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment/fake"
	_ "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func TestNewControllerWithManifestClient(t *testing.T) {
	originalPath := os.Getenv("KO_DATA_PATH")
	defer os.Setenv("KO_DATA_PATH", originalPath)

	os.Setenv("KO_DATA_PATH", "../../../cmd/manager/kodata/")
	defer transformer.TestOnlyChangeRequiredAnnotationsFromEnvVars(nil)()
	ctx, _ := ktesting.SetupFakeContext(t)
	c := NewController(ctx, configmap.NewStaticWatcher(cm("config-appdevexperience", nil)))
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}

func cm(name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}
}
