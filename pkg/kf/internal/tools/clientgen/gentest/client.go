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

// Package gentest contains tests for the client generator.
package gentest

import v1 "k8s.io/client-go/kubernetes/typed/core/v1"
import corev1 "k8s.io/api/core/v1"

//go:generate go run ../../option-builder/option-builder.go --pkg gentest ../common-options.yml zz_generated.clientoptions.go
//go:generate go run ../genclient.go client.yml

type ClientExtension interface {
}

// NewExampleClient creates an example client with a mutator and membership
// validator that filter based on a label.
func NewExampleClient(mockK8s v1.PodsGetter) Client {
	return &coreClient{
		kclient:      mockK8s,
		upsertMutate: LabelSetMutator(map[string]string{"is-a": "OperatorConfig"}),
	}
}

func LabelSetMutator(labels map[string]string) Mutator {
	return func(obj *corev1.Pod) error {
		if obj.Labels == nil {
			obj.Labels = make(map[string]string)
		}

		for key, value := range labels {
			obj.Labels[key] = value
		}

		return nil
	}
}
