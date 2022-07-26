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
)

func TestServiceBroker_SetDefaults(t *testing.T) {
	cases := testutil.ApisDefaultableTestSuite{
		"empty": {
			Context: defaultContext(),
			Input:   &ServiceBroker{},
			Want: (func() *ServiceBroker {
				out := &ServiceBroker{}
				out.Spec.UpdateRequests = 1
				return out
			}()),
		},
	}

	cases.Run(t)
}

func TestClusterServiceBroker_SetDefaults(t *testing.T) {
	cases := testutil.ApisDefaultableTestSuite{
		"empty": {
			Context: defaultContext(),
			Input:   &ClusterServiceBroker{},
			Want: (func() *ClusterServiceBroker {
				out := &ClusterServiceBroker{}
				out.Spec.UpdateRequests = 1
				return out
			}()),
		},
	}

	cases.Run(t)
}
