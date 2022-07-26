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

package networking

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

func makeConfigMap(name string, namespace string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = name
	cm.Namespace = namespace

	return cm
}

func makeDeployment(rev, revSemver string) *v1.Deployment {
	deployment := v1.Deployment{}
	deployment.Labels = map[string]string{
		"istio.io/rev":              rev,
		"operator.istio.io/version": revSemver,
	}
	return &deployment
}

func TestLookupASMRev(t *testing.T) {
	t.Run("managed ASM", func(t *testing.T) {
		asmRev, err := LookupASMRev(
			func(namespace string, selector labels.Selector) ([]*v1.Deployment, error) {
				return []*v1.Deployment{
					makeDeployment("asm-195-2", "1.9.5-asm.2"),
				}, nil
			}, func(namespace, configMapName string) (*corev1.ConfigMap, error) {
				return makeConfigMap("istio-asm-managed", "istio-system"), nil
			})
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "asmRev", "asm-managed", asmRev)
	})

	t.Run("single control plane", func(t *testing.T) {
		asmRev, err := LookupASMRev(
			func(namespace string, selector labels.Selector) ([]*v1.Deployment, error) {
				return []*v1.Deployment{
					makeDeployment("asm-195-2", "1.9.5-asm.2"),
				}, nil
			}, func(namespace, configMapName string) (*corev1.ConfigMap, error) {
				return nil, nil
			})
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "asmRev", "asm-195-2", asmRev)
	})

	t.Run("no control plane", func(t *testing.T) {
		_, err := LookupASMRev(
			func(namespace string, selector labels.Selector) ([]*v1.Deployment, error) {
				return []*v1.Deployment{}, nil
			}, func(namespace, configMapName string) (*corev1.ConfigMap, error) {
				return nil, nil
			})
		testutil.AssertErrorContainsAll(t, err, []string{"failed to find istio deployment"})
	})

	t.Run("multiple control plane", func(t *testing.T) {
		asmRev, err := LookupASMRev(
			func(namespace string, selector labels.Selector) ([]*v1.Deployment, error) {
				return []*v1.Deployment{
					makeDeployment("asm-1101-1", "1.10.1-asm.1"),
					makeDeployment("asm-183-2", "1.8.3-asm.2"),
					makeDeployment("asm-195-2", "1.9.5-asm.2"),
				}, nil
			}, func(namespace, configMapName string) (*corev1.ConfigMap, error) {
				return nil, nil
			})
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "asmRev", "asm-1101-1", asmRev)
	})
}

func TestIsManagedASM(t *testing.T) {
	t.Run("managed ASM", func(t *testing.T) {
		configMap := makeConfigMap("istio-asm-managed", "istio-system")

		isManaged, err := IsManagedASM(func(namespace, configMapName string) (*corev1.ConfigMap, error) {
			return configMap, nil
		})

		testutil.AssertNil(t, "err", err)
		testutil.AssertTrue(t, "isManaged", isManaged)
	})

	t.Run("in-cluster ASM", func(t *testing.T) {
		isManaged, err := IsManagedASM(func(namespace, configMapName string) (*corev1.ConfigMap, error) {
			return nil, apierrors.NewNotFound(v1.Resource("configmap"), "istio-asm-managed")
		})

		testutil.AssertNil(t, "err", err)
		testutil.AssertFalse(t, "isManaged", isManaged)
	})
}
