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

package resources

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	systemenvinjectorfake "github.com/google/kf/pkg/kf/systemenvinjector/fake"
	"github.com/google/kf/pkg/kf/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleKfInjectedEnvSecretName() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	space := &v1alpha1.Space{}
	space.Name = "my-space"

	fmt.Println(KfInjectedEnvSecretName(app))

	// Output: kf-injected-envs-my-app
}

func TestMakeKfInjectedEnvSecret_happyPath(t *testing.T) {
	vcapApplication := v1.EnvVar{
		Name:  "VCAP_APPLICATION",
		Value: `{"application_name":"some-app"}`,
	}

	vcapServices := v1.EnvVar{
		Name:  "VCAP_SERVICES",
		Value: "{}",
	}

	envVars := []v1.EnvVar{vcapApplication, vcapServices}
	ctrl := gomock.NewController(t)
	fakeInjector := systemenvinjectorfake.NewFakeSystemEnvInjector(ctrl)
	fakeInjector.EXPECT().ComputeSystemEnv(gomock.Any()).Return(envVars, nil)

	app := v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
			Name:      "some-app-name",
			Labels:    map[string]string{"a": "1", "b": "2"},
		},
	}

	space := v1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
			Name:      "some-namespace",
		},
	}

	secret, err := MakeKfInjectedEnvSecret(&app, &space, fakeInjector)

	testutil.AssertNil(t, "err", err)
	testutil.AssertNotNil(t, "secret", secret)
	testutil.AssertEqual(
		t,
		"secret.Name",
		KfInjectedEnvSecretName(&app),
		secret.Name,
	)

	testutil.AssertEqual(t, "secret.Labels", map[string]string{
		"a":                     "1",
		"b":                     "2",
		v1alpha1.NameLabel:      "some-app-name",
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "secret",
	}, secret.Labels)

	testutil.AssertEqual(t,
		"secret.OwnerReferences",
		"some-app-name",
		secret.OwnerReferences[0].Name,
	)

	testutil.AssertEqual(t,
		"Num of env vars",
		2,
		len(secret.Data),
	)

	testutil.AssertEqual(t,
		"VCAP application",
		vcapApplication.Value,
		string(secret.Data[vcapApplication.Name]),
	)

	testutil.AssertEqual(t,
		"VCAP services",
		vcapServices.Value,
		string(secret.Data[vcapServices.Name]),
	)
}
