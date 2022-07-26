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

package networkpolicies

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

// TestIntegration_NetworkPolicies tests that network policies can be applied.
func TestIntegration_NetworkPolicies(t *testing.T) {
	// This test tends to take a little while longer, so we're going to give
	// it more time because of all the retries.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName(
			"integration-np",
			fmt.Sprint(time.Now().UnixNano()),
		)

		path := filepath.Join(integration.RootDir(ctx, t), "samples", "apps", "helloworld")
		kf.CachePush(ctx, appName, path)

		kf.WithProxy(ctx, appName, func(addr string) {
			// Make sure the value can be fetched normally, retry in case the
			// proxy isn't ready right away.
			{
				resp, respCancel := integration.RetryGet(ctx, t, addr, 90*time.Second, http.StatusOK)
				defer resp.Body.Close()
				defer respCancel()
				integration.Logf(t, "testing for 200")
			}

			// Disallow networking.
			kf.ConfigureSpace(ctx, "set-app-ingress-policy", v1alpha1.DenyAllNetworkPolicy)

			// Sleep for a moment for the Pods to be updated. There is no way
			// to query the NetworkPolicy directly to determine if Calico has
			// updated the iptables rules.
			time.Sleep(5 * time.Second)

			testutil.AssertRetrySucceeds(ctx, t, 5, 5*time.Second, func() error {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				var err error

				// Start a new proxy because reusing the previous one will
				// allow the connection to continue and the updated rules
				// won't terminate an existing connection.
				kf.WithProxy(ctx, appName, func(addr string) {
					// Wait for the proxy to start
					time.Sleep(2 * time.Second)

					// Manually create this request since RetryGet assumes we
					// don't want a failure.
					req, err := http.NewRequest(http.MethodGet, addr, nil)
					if err != nil {
						err = fmt.Errorf("request creation error: %v", err)
						return
					}

					// Expect timeout rather than a success
					rctx, rctxcancel := context.WithTimeout(ctx, 3*time.Second)
					defer rctxcancel()
					req = req.WithContext(rctx)
					_, err = http.DefaultClient.Do(req)
					if !errors.Is(err, context.DeadlineExceeded) {
						err = fmt.Errorf("Expected DeadlineExceeded error, got: %v", err)
						return
					}
				})

				return err
			})
		})
	})
}
