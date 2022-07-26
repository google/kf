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

package runbookgenerator

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestGenTroubleshooterRunbook(t *testing.T) {

	nsType := &genericcli.KubernetesType{
		NsScoped: true,
		KfName:   "Foo",
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "foos",
	}

	clusterType := &genericcli.KubernetesType{
		NsScoped: false,
		KfName:   "ClusterFoo",
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "clusterfoos",
	}

	problems := []troubleshooter.Problem{
		{
			Description: "First problem.",
			Causes: []troubleshooter.Cause{
				{Description: "First Cause.", Recommendation: "**Some** recommendation [link](https://example.com)."},
				{Description: "Second Cause.", Recommendation: "**Some** recommendation [link](https://example.com)."},
			},
		},
		{
			Description: "Second problem.",
			Causes: []troubleshooter.Cause{
				{Description: "Cause 2.1.", Recommendation: "**Some** recommendation [link](https://example.com)."},
				{Description: "Cause 2.2.", Recommendation: "**Some** recommendation [link](https://example.com)."},
			},
		},
	}

	cases := map[string]struct {
		component  troubleshooter.Component
		docVersion string
	}{
		"namespaced empty": {
			component: troubleshooter.Component{
				Type: nsType,
			},
			docVersion: "v1.0",
		},
		"cluster empty": {
			component: troubleshooter.Component{
				Type: clusterType,
			},
			docVersion: "v1.0",
		},
		"namespaced empty v2": {
			component: troubleshooter.Component{
				Type: nsType,
			},
			docVersion: "v2.0",
		},
		"namespaced with problems": {
			component: troubleshooter.Component{
				Type:     clusterType,
				Problems: problems,
			},
			docVersion: "v1.0",
		},
		"cluster with problems": {
			component: troubleshooter.Component{
				Type:     clusterType,
				Problems: problems,
			},
			docVersion: "v1.0",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := &bytes.Buffer{}
			GenTroubleshooterRunbook(context.Background(), buf, tc.component, tc.docVersion)
			testutil.AssertGoldenContext(t, "runbook", buf.Bytes(), map[string]interface{}{
				"namespaced?":  tc.component.Type.Namespaced(),
				"friendlyname": tc.component.Type.FriendlyName(),
				"gvr":          tc.component.Type.GroupVersionResource(context.Background()),
				"version":      tc.docVersion,
			})
		})
	}
}
