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

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

// defaultVolumeCapacity is a no-op k8s volume capacity value used to construct volume claim.
const defaultVolumeCapacity = "1Gi"

// GetPeristentVolumeName returns the desired name for the PersistentVolume of the service instance.
func GetPersistentVolumeName(serviceInstanceName string, namespace string) string {
	return v1alpha1.GenerateName(serviceInstanceName, namespace, "pv")
}

// GetPeristentVolumeClaimName returns the desired name for the PersistentVolumeClaim of the service instance.
func GetPersistentVolumeClaimName(serviceInstanceName string) string {
	return v1alpha1.GenerateName(serviceInstanceName, "pvc")
}

// MakePersistentVolume constructs a k8s volume for the volume service.
func MakePersistentVolume(
	serviceInstance *v1alpha1.ServiceInstance,
	params *v1alpha1.VolumeInstanceParams,
	persistentvolumeclaim *corev1.PersistentVolumeClaim,
) (*corev1.PersistentVolume, error) {

	capacity, err := resource.ParseQuantity(defaultVolumeCapacity)
	if err != nil {
		return nil, fmt.Errorf("invalid capacity: %v", defaultVolumeCapacity)
	}

	items := strings.SplitN(params.Share, "/", 2)
	if len(items) != 2 {
		return nil, fmt.Errorf("invalid share: %v", params.Share)
	}

	server, path := items[0], "/"+items[1]

	spec := corev1.PersistentVolumeSpec{
		Capacity: corev1.ResourceList{
			corev1.ResourceStorage: capacity,
		},
		AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany, corev1.ReadWriteMany, corev1.ReadWriteOnce},
		PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
		StorageClassName:              "",
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			NFS: &corev1.NFSVolumeSource{
				Server: server,
				Path:   path,
			},
		},
	}

	if persistentvolumeclaim != nil {
		spec.ClaimRef = &corev1.ObjectReference{
			Name:      persistentvolumeclaim.Name,
			Namespace: persistentvolumeclaim.Namespace,
		}
	}

	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   GetPersistentVolumeName(serviceInstance.Name, serviceInstance.Namespace),
			Labels: v1alpha1.UnionMaps(serviceInstance.GetLabels()),
		},
		Spec: spec,
	}, nil
}

// MakePersistentVolumeClaim constructs a k8s volume claim for the volume service.
func MakePersistentVolumeClaim(
	serviceInstance *v1alpha1.ServiceInstance,
) (*corev1.PersistentVolumeClaim, error) {

	capacity, err := resource.ParseQuantity(defaultVolumeCapacity)
	if err != nil {
		return nil, fmt.Errorf("invalid capacity: %v", defaultVolumeCapacity)
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetPersistentVolumeClaimName(serviceInstance.Name),
			Namespace: serviceInstance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(serviceInstance),
			},
			Labels: v1alpha1.UnionMaps(serviceInstance.GetLabels()),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: capacity,
				},
			},
			StorageClassName: ptr.String(""),
			VolumeName:       GetPersistentVolumeName(serviceInstance.Name, serviceInstance.Namespace),
		},
	}, nil
}
