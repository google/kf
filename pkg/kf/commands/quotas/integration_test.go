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

package quotas

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/google/kf/pkg/kf/testutil"
)

// Skipping quota integration tests for now.
// TODO(#442): Move tests into space e2e tests, or create and delete space
// in these tests

// TestIntegration_Create creates a resourcequota, then tries to retrieve the
// quota info.  It finally deletes the quota.
func TestIntegration_Create(t *testing.T) {
	t.Skip("#442")
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		spaceName := fmt.Sprintf("integration-quota-space-%d", time.Now().UnixNano())
		defer kf.DeleteQuota(ctx, spaceName)

		memQuantity := "55Gi"
		cpuQuantity := "11"
		routesQuantity := "22"

		createQuotaOutput, err := kf.CreateQuota(ctx, spaceName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)

		AssertNil(t, "create quota error", err)
		AssertContainsAll(t, strings.Join(createQuotaOutput, " "), []string{spaceName, "successfully created"})
		Logf(t, "done ensuring quota was created.")

		// Get the quota
		getQuotaOutput, err := kf.GetQuota(ctx, spaceName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{spaceName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring correct quota info returned.")
	})
}

// TestIntegration_Delete creates a quota and then deletes it. It then makes
// sure the quota no longer exists.
func TestIntegration_Delete(t *testing.T) {
	t.Skip("#442")
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		spaceName := fmt.Sprintf("integration-quota-space-%d", time.Now().UnixNano())

		memQuantity := "55Gi"
		cpuQuantity := "11"
		routesQuantity := "22"

		kf.CreateQuota(ctx, spaceName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)

		getQuotaOutput, err := kf.GetQuota(ctx, spaceName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{spaceName, memQuantity, cpuQuantity, routesQuantity})

		deleteQuotaOutput, err := kf.DeleteQuota(ctx, spaceName)
		AssertNil(t, "delete quota error", err)
		AssertContainsAll(t, strings.Join(deleteQuotaOutput, " "), []string{spaceName, "successfully deleted"})

		// Getting the quota should output an error saying the quota is not found
		getQuotaOutput, err = kf.GetQuota(ctx, spaceName)
		AssertNotNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{spaceName, "Error", "not found"})
	})
}

// TestIntegration_Update creates a quota then updates it.
// It checks that the quota is saved with the new values.
// Afterwards, the quota is deleted.
func TestIntegration_Update(t *testing.T) {
	t.Skip("#442")
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		spaceName := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())
		defer kf.DeleteQuota(ctx, spaceName)

		memQuantity := "55Gi"
		cpuQuantity := "11"

		kf.CreateQuota(ctx, spaceName,
			"-m", memQuantity,
			"-c", cpuQuantity,
		)

		getQuotaOutput, err := kf.GetQuota(ctx, spaceName)
		AssertNil(t, "get quota error", err)

		// Check that routes quota appears by default as "0"
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{spaceName, memQuantity, cpuQuantity, "0"})
		Logf(t, "done ensuring correct quota info returned.")

		memQuantity = "66Gi"
		cpuQuantity = "22"
		routesQuantity := "8"

		_, err = kf.UpdateQuota(ctx, spaceName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)
		AssertNil(t, "update quota error", err)

		getQuotaOutput, err = kf.GetQuota(ctx, spaceName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{spaceName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring quota info updated.")

		// Reset memory quota
		memQuantity = "0"
		_, err = kf.UpdateQuota(ctx, spaceName,
			"-m", memQuantity,
		)
		AssertNil(t, "update quota error", err)

		getQuotaOutput, err = kf.GetQuota(ctx, spaceName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{spaceName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring memory quota was reset.")
	})
}

var checkOnce sync.Once

func checkClusterStatus(t *testing.T) {
	checkOnce.Do(func() {
		testIntegration_WaitForCluster(t)
	})
}

// testIntegration_WaitForCluster runs the doctor command. It ensures the
// cluster the tests are running against is in good shape.
func testIntegration_WaitForCluster(t *testing.T) {
	t.Skip("#442")
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		kf.WaitForCluster(ctx)
	})
}
