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

package servicebrokers

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	cluster "github.com/google/kf/pkg/kf/service-brokers/cluster/fake"
	namespaced "github.com/google/kf/pkg/kf/service-brokers/namespaced/fake"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestNewCreateServiceBrokerCommand(t *testing.T) {
	type mocks struct {
		p                *config.KfParams
		clusterClient    *cluster.FakeClient
		namespacedClient *namespaced.FakeClient
	}

	cases := map[string]struct {
		args    []string
		setup   func(t *testing.T, mocks mocks)
		wantErr error
		wantOut string
	}{
		"no params": {
			args:    []string{},
			wantErr: errors.New("accepts 2 arg(s), received 0"),
		},
		"no namespace space scoped": {
			args: []string{"some-broker", "https://broker-url", "--space-scoped"},
			setup: func(t *testing.T, mocks mocks) {
				// unset namespace
				mocks.p.Namespace = ""
			},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"no namespace global scoped": {
			args: []string{"some-broker", "https://broker-url"},
			setup: func(t *testing.T, mocks mocks) {
				// unset namespace
				mocks.p.Namespace = ""

				// expect ok
				mocks.clusterClient.EXPECT().Create(gomock.Any())
				mocks.clusterClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantOut: "Creating cluster service broker \"some-broker\"...\nSuccess\n",
		},
		"global scoped arguments get passed correctly": {
			args: []string{"cluster-broker", "https://broker-url"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.clusterClient.EXPECT().Create(PopulateClusterBrokerTemplate("cluster-broker", "https://broker-url"))
				mocks.clusterClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "cluster-broker", gomock.Any())
			},
			wantOut: "Creating cluster service broker \"cluster-broker\"...\nSuccess\n",
		},
		"namespaced arguments get passed correctly": {
			args: []string{"ns-broker", "https://broker-url", "--space-scoped"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.p.Namespace = "custom-ns"
				mocks.namespacedClient.EXPECT().Create("custom-ns", PopulateSpaceBrokerTemplate("custom-ns", "ns-broker", "https://broker-url"))
				mocks.namespacedClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns", "ns-broker", gomock.Any())
			},
			wantOut: "Creating service broker \"ns-broker\" in space \"custom-ns\"...\nSuccess\n",
		},
		"global broker async creation": {
			args: []string{"cluster-broker", "https://broker-url", "--async"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.clusterClient.EXPECT().Create(gomock.Any())
			},
			wantOut: "Creating cluster service broker \"cluster-broker\" asynchronously\n",
		},
		"namespaced broker async creation": {
			args: []string{"ns-broker", "https://broker-url", "--space-scoped", "--async"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.namespacedClient.EXPECT().Create(gomock.Any(), gomock.Any())
			},
			wantOut: "Creating service broker \"ns-broker\" in space \"default\" asynchronously\n",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			args := mocks{
				p: &config.KfParams{
					Namespace: "default",
				},
				clusterClient:    cluster.NewFakeClient(ctrl),
				namespacedClient: namespaced.NewFakeClient(ctrl),
			}

			if tc.setup != nil {
				tc.setup(t, args)
			}

			buf := new(bytes.Buffer)
			cmd := NewCreateServiceBrokerCommand(args.p, args.clusterClient, args.namespacedClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			_, actualErr := cmd.ExecuteC()
			if tc.wantErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "output", buf.String(), tc.wantOut)
		})
	}
}

func ExamplePopulateSpaceBrokerTemplate() {
	template := PopulateSpaceBrokerTemplate("some-ns", "some-name", "some-url")
	fmt.Println("Name:", template.Name)
	fmt.Println("Namespace:", template.Namespace)
	fmt.Println("URL:", template.Spec.URL)

	// Output: Name: some-name
	// Namespace: some-ns
	// URL: some-url
}

func ExamplePopulateClusterBrokerTemplate() {
	template := PopulateClusterBrokerTemplate("some-name", "some-url")
	fmt.Println("Name:", template.Name)
	fmt.Println("URL:", template.Spec.URL)

	// Output: Name: some-name
	// URL: some-url
}
