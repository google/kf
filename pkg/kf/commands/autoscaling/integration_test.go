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

package autoscaling

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
)

func TestIntegration_Autoscaling(t *testing.T) {
	appName := fmt.Sprintf("integration-autoscaling-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/echo"
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		integration.WithApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
			// Create autoscaling rule.
			kf.RunCommand(ctx, "create-autoscaling-rule", appName, "cpu", "10", "80")

			// Update autoscaling limits to scale up App instances.
			kf.RunCommand(ctx, "update-autoscaling-limits", appName, "3", "5")
			app, ok := kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			testutil.AssertEqual(t, "app instances", "1", app.Instances)

			// Enable autoscaling for a deployment/App.
			kf.RunCommand(ctx, "enable-autoscaling", appName)
			integration.Logf(t, "ensuring App is scaled up")
			app, ok = kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			// Only checking autoscaling status, instances may not be scaled up yet
			testutil.AssertContainsAll(t, app.Instances, []string{"autoscaled 3 to 5"})
			integration.Logf(t, "done ensuring App is scaled up")

			// Update autoscaling limits to scale down App instances.
			kf.RunCommand(ctx, "update-autoscaling-limits", appName, "1", "1")
			integration.Logf(t, "ensuring App is scaled down")
			app, ok = kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			testutil.AssertContainsAll(t, app.Instances, []string{"autoscaled 1 to 1"})
			integration.Logf(t, "done ensuring App is scaled down")

			// Disable autoscaling.
			kf.RunCommand(ctx, "disable-autoscaling", appName)
			// Make sure autoscaling is turned off
			integration.Logf(t, "ensuring autoscaling is turned off")
			app, ok = kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			testutil.AssertEqual(t, "app instances", "1", app.Instances)
			integration.Logf(t, "done ensuring App with autoscaling is turned off")

			// Updating autoscaling limits will not scale App after autoscaling is disabled
			kf.RunCommand(ctx, "update-autoscaling-limits", appName, "3", "5")
			integration.Logf(t, "ensuring App is not scaled up")
			app, ok = kf.Apps(ctx)[appName]
			testutil.AssertEqual(t, "app presence", true, ok)
			testutil.AssertEqual(t, "app instances", "1", app.Instances)
			integration.Logf(t, "done ensuring App is not scaled up")
		})
	})
}
