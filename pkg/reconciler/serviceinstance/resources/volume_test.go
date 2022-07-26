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

package resources

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeVolume(t *testing.T) {
	t.Parallel()
	wantCapacity := "1Gi"

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		params          *v1alpha1.VolumeInstanceParams
		pvc             *corev1.PersistentVolumeClaim
		wantErr         error
	}{
		"invalid quantity ok": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share:    "host/mount",
				Capacity: "9z",
			},
		},
		"happy case": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share:    "host/mount",
				Capacity: "9Gi",
			},
		},
		"happy partial case": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share: "host/mount",
			},
		},
		"happy case with claimref": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share:    "host/mount",
				Capacity: "9Gi",
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc",
					Namespace: "test-space",
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			pv, err := MakePersistentVolume(
				tc.serviceInstance,
				tc.params,
				tc.pvc,
			)

			if tc.wantErr != nil {
				testutil.AssertEqual(t, "error", tc.wantErr, err)
			} else {
				testutil.AssertNil(t, "error", err)
				testutil.AssertNotNil(t, "pv", pv)
				testutil.AssertNotNil(t, "nfs", pv.Spec.NFS)
				testutil.AssertEqual(t, "", pv.Spec.StorageClassName, "")
				testutil.AssertEqual(t, "", pv.Spec.Capacity[corev1.ResourceStorage], resource.MustParse(wantCapacity))
				items := strings.SplitN(tc.params.Share, "/", 2)
				server, path := items[0], "/"+items[1]

				testutil.AssertEqual(t, "", pv.Spec.NFS.Server, server)
				testutil.AssertEqual(t, "", pv.Spec.NFS.Path, path)
				if tc.pvc != nil {
					testutil.AssertEqual(t, "claimRef name", pv.Spec.ClaimRef.Name, tc.pvc.Name)
					testutil.AssertEqual(t, "claimRef namespace", pv.Spec.ClaimRef.Namespace, tc.pvc.Namespace)
				}
			}
		})
	}
}

func TestMakeVolumeClaim(t *testing.T) {
	t.Parallel()
	wantCapacity := "1Gi"

	cases := map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		params          *v1alpha1.VolumeInstanceParams
		wantErr         error
	}{
		"invalid quantity ok": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share:    "host/mount",
				Capacity: "9z",
			},
		},
		"happy case": {
			serviceInstance: fakeVolumeInstance(),
			params: &v1alpha1.VolumeInstanceParams{
				Share: "host/mount",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			pvc, err := MakePersistentVolumeClaim(
				tc.serviceInstance,
			)

			if tc.wantErr != nil {
				testutil.AssertEqual(t, "error", tc.wantErr, err)
			} else {
				testutil.AssertNil(t, "error", err)
				testutil.AssertNotNil(t, "pv", pvc)
				testutil.AssertEqual(t, "storageclass", *pvc.Spec.StorageClassName, "")

				testutil.AssertEqual(t, "resource", pvc.Spec.Resources.Requests[corev1.ResourceStorage], resource.MustParse(wantCapacity))
				testutil.AssertEqual(t, "volumename", pvc.Spec.VolumeName, fmt.Sprintf("%s-%s-pv", tc.serviceInstance.Name, tc.serviceInstance.Namespace))
			}
		})
	}
}

func fakeVolumeInstance() *v1alpha1.ServiceInstance {
	instance := &v1alpha1.ServiceInstance{}
	instance.Name = "mydb"
	instance.Namespace = "test-ns"
	instance.UID = "00000000-0000-0000-0000-000008675309"
	instance.Spec.Volume = &v1alpha1.OSBInstance{
		ClassUID: "class-uid",
		PlanUID:  "plan-uid",
	}

	return instance
}
