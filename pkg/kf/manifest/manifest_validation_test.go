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
	"math"
	"testing"

	kfapis "github.com/google/kf/v2/pkg/apis/kf"
	"github.com/google/kf/v2/pkg/kf/testutil"
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
				KfApplicationExtension: KfApplicationExtension{
					Args:       []string{"-m", "SimpleHTTPServer"},
					Entrypoint: "python",
				},
			},
		},
		"command and args": {
			spec: Application{
				Command: "python",
				KfApplicationExtension: KfApplicationExtension{
					Args: []string{"-m", "SimpleHTTPServer"},
				},
			},
			want: apis.ErrMultipleOneOf("args", "command"),
		},
		"entrypoint and command": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Entrypoint: "/lifecycle/launcher",
				},
				Command: "python",
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
		"good ports and routes": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: protocolHTTP},
						{Port: 8081, Protocol: protocolHTTP2},
						{Port: 8082, Protocol: protocolTCP},
					},
				},
				Routes: []Route{
					{Route: "default"},
					{Route: "explicit", AppPort: 8080},
				},
			},
			want: nil,
		},
		"duplicate port": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: protocolHTTP},
						{Port: 8080, Protocol: protocolHTTP2},
					},
				},
			},
			want: kfapis.ErrDuplicateValue(8080, "ports[1].port"),
		},
		"bad protocol": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: "foo"},
					},
				},
			},
			want: apis.ErrInvalidValue("must be one of: [http http2 tcp]", "ports[0].protocol"),
		},
		"bad port": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 80808080, Protocol: "tcp"},
					},
				},
			},
			want: apis.ErrOutOfBoundsValue(80808080, 1, math.MaxUint16, "ports[0].port"),
		},
		"route port missing": {
			spec: Application{
				Routes: []Route{
					{Route: "missing-port", AppPort: 8080},
				},
			},
			want: apis.ErrInvalidValue("must match a declared port", "routes[0].appPort"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
