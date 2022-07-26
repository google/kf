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

package troubleshooter

import (
	"context"

	"github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ExampleComponent_DiagnoseSingle() {
	c := Component{
		Type: &genericcli.KubernetesType{
			KfName: "test",
		},
		Problems: []Problem{
			{
				Description: "Some problem.",
				Filter: func(_ *unstructured.Unstructured) bool {
					return true
				},
				Causes: []Cause{
					{
						Description: "Not the cause.",
						Filter: func(_ *unstructured.Unstructured) bool {
							return false
						},
					},
					{
						Description:    "Possible cause.",
						Filter:         nil,
						Recommendation: "Check your inputs.",
					},
					{
						Description: "Definite cause.",
						Filter: func(_ *unstructured.Unstructured) bool {
							return true
						},
						Recommendation: "Do XYZ.",
					},
				},
			},
		},
	}

	obj := &unstructured.Unstructured{}
	obj.SetName("test-obj")
	obj.SetNamespace("test-ns")

	c.DiagnoseSingle(context.Background(), doctor.NewDefaultDiagnostic(), obj)

	// Output: === RUN	doctor/namespace="test-ns",name="test-obj"
	// === RUN	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem.
	// === RUN	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Not the cause.
	// === RUN	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Possible cause.
	// === LOG	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Possible cause.
	// Cause: Possible cause.
	// Recommendation: Check your inputs.
	//
	// === RUN	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Definite cause.
	// === LOG	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Definite cause.
	// Cause: Definite cause.
	// Recommendation: Do XYZ.
	//
	// === LOG	doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem.
	// === LOG	doctor/namespace="test-ns",name="test-obj"
	// --- FAIL: doctor/namespace="test-ns",name="test-obj"
	//     --- FAIL: doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem.
	//         --- PASS: doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Not the cause.
	//         --- WARN: doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Possible cause.
	//         --- FAIL: doctor/namespace="test-ns",name="test-obj"/Symptom: Some problem./Cause: Definite cause.
}
