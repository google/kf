// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/ptr"
)

func TestApplication_ToResourceRequests(t *testing.T) {
	cases := map[string]struct {
		source       Application
		expectedList corev1.ResourceList
		expectedErr  error
	}{
		"full": {
			source: Application{
				Memory:    "30MB", // CF uses X and XB as SI units, these get changed to SI
				DiskQuota: "1G",
				KfApplicationExtension: KfApplicationExtension{
					CPU: "200m",
				},
			},
			expectedList: corev1.ResourceList{
				corev1.ResourceMemory:           resource.MustParse("30Mi"),
				corev1.ResourceCPU:              resource.MustParse("200m"),
				corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
			},
		},
		"normal cf subset": {
			source: Application{
				Memory:    "30M",
				DiskQuota: "1Gi",
			},
			expectedList: corev1.ResourceList{
				corev1.ResourceMemory:           resource.MustParse("30Mi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
			},
		},
		"bad quantity": {
			source: Application{
				Memory: "30Y",
			},
			expectedErr: errors.New("couldn't parse resource quantity 30Y: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"),
		},
		"no quotas": {
			source:       Application{},
			expectedList: nil, // explicitly want nil rather than the empty map
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualList, actualErr := tc.source.ToResourceRequests()

			testutil.AssertErrorsEqual(t, tc.expectedErr, actualErr)
			testutil.AssertEqual(t, "resource lists", tc.expectedList, actualList)
		})
	}
}

func TestApplication_ToAppSpecInstances(t *testing.T) {
	cases := map[string]struct {
		source   Application
		expected v1alpha1.AppSpecInstances
	}{
		"blank app": {
			source:   Application{},
			expected: v1alpha1.AppSpecInstances{},
		},
		"stopped autoscaled app": {
			source: Application{
				KfApplicationExtension: KfApplicationExtension{
					NoStart:  ptr.Bool(true),
					MinScale: intPtr(2),
					MaxScale: intPtr(300),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Stopped: true,
				Min:     intPtr(2),
				Max:     intPtr(300),
			},
		},
		"started app with instances": {
			source: Application{
				Instances: intPtr(3),
				KfApplicationExtension: KfApplicationExtension{
					NoStart: ptr.Bool(false),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Exactly: intPtr(3),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tc.source.ToAppSpecInstances()

			testutil.AssertEqual(t, "instances", tc.expected, actual)
		})
	}
}
