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

package servicebindings_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/pkg/kf/commands/config"
	servicebindings "github.com/google/kf/pkg/kf/commands/service-bindings"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewVcapServicesCommand(t *testing.T) {
	type serviceTest struct {
		Args      []string
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}

	secretSerialized := `{
    "apiVersion": "v1",
    "data": {
        "VCAP_SERVICES": "eyJzb21lIjoic2VydmljZXMifQ=="
    },
    "kind": "Secret",
    "metadata": {
        "labels": {
            "app.kubernetes.io/component": "secret",
            "app.kubernetes.io/managed-by": "kf",
            "app.kubernetes.io/name": "APP_NAME"
        },
        "name": "kf-injected-envs-APP_NAME",
        "namespace": "custom-ns"
    },
    "type": "Opaque"
}
	`

	var secret *corev1.Secret
	err := json.Unmarshal([]byte(secretSerialized), &secret)
	fmt.Println(secret.Name)
	testutil.AssertNil(t, "err", err)
	k8sclient := k8sfake.NewSimpleClientset(secret)

	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"APP_NAME"},
			Namespace: "custom-ns",
			ExpectedStrings: []string{
				`{"some":"services"}`,
			},
		},
		"empty namespace": {
			Args:        []string{"APP_NAME"},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.Namespace,
			}

			cmd := servicebindings.NewVcapServicesCommand(p, k8sclient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
				return
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
		})
	}
}
