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

package sourceimage

import "testing"

func TestIsSubPath(t *testing.T) {
	type args struct {
		path   string
		prefix string
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"identity": {
			args: args{
				path:   "/var/run/prefix",
				prefix: "/var/run/prefix",
			},
			want: true,
		},
		"sub-path": {
			args: args{
				path:   "/var/run/prefix/README.md",
				prefix: "/var/run/prefix",
			},
			want: true,
		},
		"super-path": {
			args: args{
				path:   "/var/run",
				prefix: "/var/run/prefix",
			},
			want: false,
		},
		"similar name": {
			args: args{
				path:   "/var/run/prefixy",
				prefix: "/var/run/prefix",
			},
			want: false,
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if got := IsSubPath(tc.args.path, tc.args.prefix); got != tc.want {
				t.Errorf("IsSubPath() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTrimPathPrefix(t *testing.T) {
	type args struct {
		path   string
		prefix string
	}
	cases := map[string]struct {
		args args
		want string
	}{
		"identity": {
			args: args{
				path:   "/var/run/prefix",
				prefix: "/var/run/prefix",
			},
			want: "",
		},
		"sub-path": {
			args: args{
				path:   "/var/run/prefix/README.md",
				prefix: "/var/run/prefix",
			},
			want: "README.md",
		},
		"super-path": {
			args: args{
				path:   "/var/run",
				prefix: "/var/run/prefix",
			},
			want: "/var/run",
		},
		"similar name": {
			args: args{
				path:   "/var/run/prefixy",
				prefix: "/var/run/prefix",
			},
			want: "/var/run/prefixy",
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if got := TrimPathPrefix(tc.args.path, tc.args.prefix); got != tc.want {
				t.Errorf("TrimPathPrefix() = %v, want %v", got, tc.want)
			}
		})
	}
}
