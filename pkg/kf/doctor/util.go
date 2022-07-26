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

package doctor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

// DiscoverControllerNamespaces attempts to discover controller namespaces by
// reading webhooks. The results are best-effort, ignoring errors.
func DiscoverControllerNamespaces(ctx context.Context, kubernetes kubernetes.Interface) []string {
	namespaces := sets.NewString()

	if configs, err := kubernetes.AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		List(ctx, metav1.ListOptions{}); err == nil {
		for _, config := range configs.Items {
			for _, webhook := range config.Webhooks {
				if svc := webhook.ClientConfig.Service; svc != nil {
					namespaces.Insert(svc.Namespace)
				}
			}
		}
	}

	if configs, err := kubernetes.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		List(ctx, metav1.ListOptions{}); err == nil {
		for _, config := range configs.Items {
			for _, webhook := range config.Webhooks {
				if svc := webhook.ClientConfig.Service; svc != nil {
					namespaces.Insert(svc.Namespace)
				}
			}
		}
	}

	return namespaces.List()
}
