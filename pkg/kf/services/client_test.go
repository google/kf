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

package services

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func TestProvisionSuccess(t *testing.T) {
	// These tests are adapted from SC
	// https://github.com/kubernetes-sigs/service-catalog/blob/73229d43efccdff38e3a0265229c5814af24e0c3/pkg/svcat/service-catalog/instance_test.go#L598

	type args struct {
		obj *v1beta1.ServiceInstance
		err error
	}
	tests := map[string]struct {
		args     args
		wantDone bool
		wantErr  error
	}{
		"error returned as final": {
			args: args{
				err: errors.New("bad request"),
			},
			wantDone: true,
			wantErr:  errors.New("bad request"),
		},

		"not ready": {
			args: args{
				obj: &v1beta1.ServiceInstance{
					Status: v1beta1.ServiceInstanceStatus{
						Conditions: []v1beta1.ServiceInstanceCondition{
							{Type: v1beta1.ServiceInstanceConditionReady, Status: v1beta1.ConditionFalse},
						},
					},
				},
			},
			wantDone: false,
			wantErr:  nil,
		},
		"failure": {
			args: args{
				obj: &v1beta1.ServiceInstance{
					Status: v1beta1.ServiceInstanceStatus{
						Conditions: []v1beta1.ServiceInstanceCondition{
							{
								Type:    v1beta1.ServiceInstanceConditionFailed,
								Status:  v1beta1.ConditionTrue,
								Message: "SomeMessage",
								Reason:  "SomeReason",
							},
						},
					},
				},
			},
			wantDone: true,
			wantErr:  errors.New("provision failed, message: SomeMessage reason: SomeReason"),
		},
		"ready set but AsyncOpInProgress": {
			// This is copied from a real provision right after the object is created
			// but before it's synchronized with k8s. Ready goes to false, but
			// AsyncOpInProgress is true.
			args: args{
				obj: &v1beta1.ServiceInstance{
					Status: v1beta1.ServiceInstanceStatus{
						AsyncOpInProgress: true,
						Conditions: []v1beta1.ServiceInstanceCondition{
							{
								Type:    v1beta1.ServiceInstanceConditionReady,
								Status:  v1beta1.ConditionFalse,
								Message: "The instance is being provisioned asynchronously",
								Reason:  "Provisioning",
							},
						},
					},
				},
			},
			wantDone: false,
			wantErr:  nil,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			actualDone, actualErr := ProvisionSuccess(tc.args.obj, tc.args.err)
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			testutil.AssertEqual(t, "done", tc.wantDone, actualDone)
		})
	}
}
