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

package resources

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/knative/serving/pkg/resources"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// ResourceQuotaName gets the name of the resource quota given the space.
func ResourceQuotaName(space *v1alpha1.Space) string {
	return "space-quota"
}

// MakeResourceQuota creates a ResourceQuota from a Space object.
func MakeResourceQuota(space *v1alpha1.Space) (*v1.ResourceQuota, error) {
	quota := &v1.ResourceQuota{}
	quota.ObjectMeta = metav1.ObjectMeta{
		Name:      ResourceQuotaName(space),
		Namespace: NamespaceName(space),
		OwnerReferences: []metav1.OwnerReference{
			*kmeta.NewControllerRef(space),
		},
		Labels: resources.UnionMaps(space.GetLabels(), map[string]string{
			"app.kubernetes.io/managed-by": "kf",
		}),
	}
	quota.Spec.Hard = space.Spec.ResourceLimits.SpaceQuota
	return quota, nil
}
