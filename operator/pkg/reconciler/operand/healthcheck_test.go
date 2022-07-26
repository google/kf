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

package operand_test

import (
	"context"
	"kf-operator/pkg/reconciler/operand"
	. "kf-operator/pkg/testing/k8s"
	"testing"

	mftest "kf-operator/pkg/testing/manifestival"

	poperand "kf-operator/pkg/operand"
)

var (
	controllerDeployment = Deployment("controller")
	webhookService       = Service("webhook")
)

func TestHealthCheckApply(t *testing.T) {
	tests := []struct {
		name            string
		manifest        []mftest.Object
		existingObjs    []mftest.Object
		expectedCreates []mftest.Object
		expectedFailure bool
	}{{
		name:            "Test ordering.",
		manifest:        []mftest.Object{role(), roleBinding(), webhookService, job(), controllerDeployment},
		existingObjs:    nothing,
		expectedCreates: addReconcile([]mftest.Object{role(), roleBinding(), webhookService, job(), controllerDeployment}),
		expectedFailure: false,
	},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actionList, dh, kc := injectFakeClientsWithResources(t, test.existingObjs)
			manifestR := operand.NewManifestReconciler(dh)
			r := operand.NewHealthcheckReconciler(manifestR, kc)

			err := r.Apply(context.Background(), toUnstructured(test.manifest))

			if err == nil && test.expectedFailure {
				t.Fatalf("Expected Apply to fail.")
			}
			if err != nil && !test.expectedFailure {
				t.Fatalf("Expected Apply to succeed %+v", err)
			}
			actions, err := actionList.ActionsByVerb()
			if err != nil {
				t.Fatalf("Error sorting actions %+v.", err)
			}
			if len(test.expectedCreates) > 0 {
				CheckCreates(t, actions.Creates, test.expectedCreates...)
			} else {
				CheckNoMutates(t, actionList)
			}
		})
	}
}

func TestHealthCheckGetState(t *testing.T) {
	tests := []struct {
		name          string
		manifest      []mftest.Object
		existingObjs  []mftest.Object
		expectedState string
	}{{
		name:          "Not ready due to one deployment()",
		manifest:      []mftest.Object{Deployment("one"), Deployment("two")},
		existingObjs:  []mftest.Object{Deployment("one"), Deployment("two", DeploymentReady)},
		expectedState: operand.HealthCheckDeploymentsNotReady,
	}, {
		name:          "Not ready due to other deployment()",
		manifest:      []mftest.Object{Deployment("one"), Deployment("two")},
		existingObjs:  []mftest.Object{Deployment("one", DeploymentReady), Deployment("two")},
		expectedState: operand.HealthCheckDeploymentsNotReady,
	}, {
		name:          "Installed",
		manifest:      []mftest.Object{Deployment("one"), Deployment("two"), AddAnnotation(Job("job", WithBackoffLimit(12)), "operator.knative.dev/mode", "EnsureExists")},
		existingObjs:  []mftest.Object{Deployment("one", DeploymentReady), Deployment("two", DeploymentReady), Job("job", WithBackoffLimit(11))},
		expectedState: poperand.Installed,
	}}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			actionList, dh, kc := injectFakeClientsWithResources(t, test.existingObjs)
			manifestR := operand.NewManifestReconciler(dh)
			r := operand.NewHealthcheckReconciler(manifestR, kc)

			got, err := r.GetState(context.Background(), toUnstructured(test.manifest))
			if err != nil {
				t.Fatalf("Wanted no error got %+v", err)
			}
			if got != test.expectedState {
				t.Fatalf("Wanted: [%+v] Got: [%+v]", test.expectedState, got)
			}
			CheckNoMutates(t, actionList)
		})
	}
}
