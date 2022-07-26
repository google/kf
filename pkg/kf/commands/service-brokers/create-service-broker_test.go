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
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/osbutil"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	secrets "github.com/google/kf/v2/pkg/kf/secrets/fake"
	cluster "github.com/google/kf/v2/pkg/kf/service-brokers/cluster/fake"
	namespaced "github.com/google/kf/v2/pkg/kf/service-brokers/namespaced/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestNewCreateServiceBrokerCommand(t *testing.T) {
	type mocks struct {
		p                *config.KfParams
		clusterClient    *cluster.FakeClient
		namespacedClient *namespaced.FakeClient
		secretsClient    *secrets.FakeClient
	}

	someBrokerSecretName := v1alpha1.GenerateName("some-broker", "auth")
	nsBrokerSecretName := v1alpha1.GenerateName("ns-broker", "auth")
	clusterBrokerSecretName := v1alpha1.GenerateName("cluster-broker", "auth")
	cases := map[string]struct {
		args    []string
		setup   func(t *testing.T, mocks mocks)
		wantErr error
		wantOut string
	}{
		"no params": {
			args:    []string{},
			wantErr: errors.New("accepts 4 arg(s), received 0"),
		},
		"no namespace space scoped": {
			args: []string{"some-broker", "user", "pw", "https://broker-url", "--space-scoped"},
			setup: func(t *testing.T, mocks mocks) {
				// unset namespace
				mocks.p.Space = ""
			},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"no namespace global scoped": {
			args: []string{"some-broker", "user", "pw", "https://broker-url"},
			setup: func(t *testing.T, mocks mocks) {
				// unset namespace
				mocks.p.Space = ""

				fakeClusterBroker := populateV1alpha1ClusterBrokerTemplate("some-broker", someBrokerSecretName)
				// expect ok
				mocks.clusterClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fakeClusterBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), v1alpha1.KfNamespace, osbutil.NewBasicAuthSecret(someBrokerSecretName, "user", "pw", "https://broker-url", fakeClusterBroker))
				mocks.clusterClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantOut: "Creating cluster service broker \"some-broker\"...\nSuccess\n",
		},
		"global scoped arguments get passed correctly": {
			args: []string{"cluster-broker", "user", "pw", "https://broker-url"},
			setup: func(t *testing.T, mocks mocks) {
				fakeClusterBroker := populateV1alpha1ClusterBrokerTemplate("cluster-broker", clusterBrokerSecretName)
				mocks.clusterClient.EXPECT().Create(gomock.Any(), fakeClusterBroker).Return(fakeClusterBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), v1alpha1.KfNamespace, osbutil.NewBasicAuthSecret(clusterBrokerSecretName, "user", "pw", "https://broker-url", fakeClusterBroker))
				mocks.clusterClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "cluster-broker", gomock.Any())
			},
			wantOut: "Creating cluster service broker \"cluster-broker\"...\nSuccess\n",
		},
		"namespaced arguments get passed correctly": {
			args: []string{"ns-broker", "user", "pw", "https://broker-url", "--space-scoped"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.p.Space = "custom-ns"
				fakeNsBroker := populateV1alpha1SpaceBrokerTemplate("custom-ns", "ns-broker", nsBrokerSecretName)
				mocks.namespacedClient.EXPECT().Create(gomock.Any(), "custom-ns", fakeNsBroker).Return(fakeNsBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), "custom-ns", osbutil.NewBasicAuthSecret(nsBrokerSecretName, "user", "pw", "https://broker-url", fakeNsBroker))
				mocks.namespacedClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns", "ns-broker", gomock.Any())
			},
			wantOut: "Creating service broker \"ns-broker\" in Space \"custom-ns\"...\nSuccess\n",
		},
		"global broker async creation": {
			args: []string{"cluster-broker", "user", "pw", "https://broker-url", "--async"},
			setup: func(t *testing.T, mocks mocks) {
				fakeClusterBroker := populateV1alpha1ClusterBrokerTemplate("cluster-broker", clusterBrokerSecretName)
				mocks.clusterClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fakeClusterBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantOut: "Creating cluster service broker \"cluster-broker\" asynchronously\n",
		},
		"namespaced broker async creation": {
			args: []string{"ns-broker", "user", "pw", "https://broker-url", "--space-scoped", "--async"},
			setup: func(t *testing.T, mocks mocks) {
				fakeNsBroker := populateV1alpha1SpaceBrokerTemplate("custom-ns", "ns-broker", nsBrokerSecretName)
				mocks.namespacedClient.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeNsBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantOut: "Creating service broker \"ns-broker\" in Space \"default\" asynchronously\n",
		},
		"global provision failure": {
			args: []string{"cluster-broker", "user", "pw", "https://broker-url"},
			setup: func(t *testing.T, mocks mocks) {
				fakeClusterBroker := populateV1alpha1ClusterBrokerTemplate("cluster-broker", clusterBrokerSecretName)
				mocks.clusterClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fakeClusterBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any())
				mocks.clusterClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "cluster-broker", gomock.Any()).Return(nil, errors.New("timeout"))
			},
			wantErr: errors.New("timeout"),
		},
		"namespaced provision failure": {
			args: []string{"ns-broker", "user", "pw", "https://broker-url", "--space-scoped"},
			setup: func(t *testing.T, mocks mocks) {
				mocks.p.Space = "custom-ns"
				fakeNsBroker := populateV1alpha1SpaceBrokerTemplate("custom-ns", "ns-broker", nsBrokerSecretName)
				mocks.namespacedClient.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any()).Return(fakeNsBroker, nil)
				mocks.secretsClient.EXPECT().Create(gomock.Any(), "custom-ns", gomock.Any())
				mocks.namespacedClient.EXPECT().WaitForConditionReadyTrue(gomock.Any(), "custom-ns", "ns-broker", gomock.Any()).Return(nil, errors.New("timeout"))
			},
			wantErr: errors.New("timeout"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			args := mocks{
				p: &config.KfParams{
					Space: "default",
				},
				clusterClient:    cluster.NewFakeClient(ctrl),
				namespacedClient: namespaced.NewFakeClient(ctrl),
				secretsClient:    secrets.NewFakeClient(ctrl),
			}

			if tc.setup != nil {
				tc.setup(t, args)
			}

			buf := new(bytes.Buffer)
			cmd := NewCreateServiceBrokerCommand(
				args.p,
				args.clusterClient,
				args.namespacedClient,
				args.secretsClient,
			)
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
