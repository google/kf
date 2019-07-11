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

package routes_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/google/kf/pkg/kf/testutil"
)

// TestIntegration_Routes creates a route via `create-route`, verifies it with
// `routes`, deletes it via `delete-route` and then verifies again.
func TestIntegration_Routes(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		hostname := fmt.Sprintf("some-host-%d", time.Now().UnixNano())

		findRoute := func(shouldFind bool) {
			RetryOnPanic(ctx, t, func() {
				var found bool
				for _, line := range kf.Routes(ctx) {
					expected := hostname + " example.com /some-path"
					actual := strings.Join(strings.Fields(line), " ")
					if expected == actual {
						found = true
						break
					}
				}

				if shouldFind != found {
					// We'll panic so we can use our retry logic
					panic(fmt.Errorf("Wanted %v, got %v", shouldFind, found))
				}
			})
		}

		kf.Target(ctx, "default")

		// TODO: use the domain from the cluster.
		kf.CreateRoute(ctx, "example.com", "--hostname="+hostname, "--path=some-path")
		findRoute(true)
		kf.DeleteRoute(ctx, "example.com", "--hostname="+hostname, "--path=some-path")
		findRoute(false)
	})
}
