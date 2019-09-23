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
	"context"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	"knative.dev/pkg/apis"
)

func TestApplication_Validation(t *testing.T) {
	cases := map[string]struct {
		spec Application
		want *apis.FieldError
	}{
		"valid": {
			spec: Application{},
		},
		"entrypoint and args": {
			spec: Application{
				Entrypoint: "python",
				Args:       []string{"-m", "SimpleHTTPServer"},
			},
		},
		"command and args": {
			spec: Application{
				Command: "python",
				Args:    []string{"-m", "SimpleHTTPServer"},
			},
			want: apis.ErrMultipleOneOf("args", "command"),
		},
		"entrypoint and command": {
			spec: Application{
				Entrypoint: "/lifecycle/launcher",
				Command:    "python",
			},
			want: apis.ErrMultipleOneOf("entrypoint", "command"),
		},
		"buildpack and buildpacks": {
			spec: Application{
				LegacyBuildpack: "default",
				Buildpacks:      []string{"java", "node"},
			},
			want: apis.ErrMultipleOneOf("buildpack", "buildpacks"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
