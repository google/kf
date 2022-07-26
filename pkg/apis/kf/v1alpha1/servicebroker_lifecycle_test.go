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

package v1alpha1

import (
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	apitesting "knative.dev/pkg/apis/testing"
)

func initTestCommonServiceBrokerStatus(t *testing.T) *CommonServiceBrokerStatus {
	t.Helper()
	status := &CommonServiceBrokerStatus{}
	status.InitializeConditions()

	// sanity check conditions get initiailized as unknown
	for _, cond := range status.Conditions {
		apitesting.CheckConditionOngoing(status.duck(), cond.Type, t)
	}

	// sanity check total conditions (add 1 for "Ready")
	testutil.AssertEqual(t, "conditions count", 4, len(status.Conditions))

	return status
}

func TestCommonServiceBrokerStatus_lifecycle(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		Init func(status *CommonServiceBrokerStatus)

		ExpectSucceeded []apis.ConditionType
		ExpectFailed    []apis.ConditionType
		ExpectOngoing   []apis.ConditionType
	}{
		"happy path": {
			Init: func(status *CommonServiceBrokerStatus) {
				status.CredsSecretCondition().MarkSuccess()
				status.CredsSecretPopulatedCondition().MarkSuccess()
				status.CatalogCondition().MarkSuccess()
			},
			ExpectSucceeded: []apis.ConditionType{
				CommonServiceBrokerConditionReady,
				CommonServiceBrokerConditionCatalogReady,
				CommonServiceBrokerConditionCredsSecretReady,
				CommonServiceBrokerConditionCredsSecretPopulatedReady,
			},
		},
		"deletion blocked": {
			Init: func(status *CommonServiceBrokerStatus) {
				status.PropagateDeletionBlockedStatus()
			},
			ExpectSucceeded: []apis.ConditionType{},
			ExpectFailed: []apis.ConditionType{
				CommonServiceBrokerConditionReady,
			},
		},
		"populated secret": {
			Init: func(status *CommonServiceBrokerStatus) {
				secret := &corev1.Secret{}
				secret.Data = map[string][]byte{
					"key": {},
				}
				status.PropagateSecretStatus(secret)
			},
			ExpectSucceeded: []apis.ConditionType{
				CommonServiceBrokerConditionCredsSecretReady,
				CommonServiceBrokerConditionCredsSecretPopulatedReady,
			},
		},
		"blank secret": {
			Init: func(status *CommonServiceBrokerStatus) {
				status.PropagateSecretStatus(&corev1.Secret{})
			},
			ExpectSucceeded: []apis.ConditionType{
				CommonServiceBrokerConditionCredsSecretReady,
			},
			ExpectOngoing: []apis.ConditionType{
				CommonServiceBrokerConditionCredsSecretPopulatedReady,
			},
		},
		"nil secret": {
			Init: func(status *CommonServiceBrokerStatus) {
				status.PropagateSecretStatus(nil)
			},
			ExpectOngoing: []apis.ConditionType{
				CommonServiceBrokerConditionCredsSecretReady,
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			status := initTestCommonServiceBrokerStatus(t)

			tc.Init(status)

			for _, exp := range tc.ExpectFailed {
				apitesting.CheckConditionFailed(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectOngoing {
				apitesting.CheckConditionOngoing(status.duck(), exp, t)
			}

			for _, exp := range tc.ExpectSucceeded {
				apitesting.CheckConditionSucceeded(status.duck(), exp, t)
			}
		})
	}
}
