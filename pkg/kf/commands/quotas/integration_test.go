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

// TestIntegration_Create creates a resourcequota, then tries to retrieve the quota info.
// It finally deletes the quota.
func TestIntegration_Create(t *testing.T) {
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		quotaName := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())
		defer kf.DeleteQuota(ctx, quotaName)

		memQuantity := "55Gi"
		cpuQuantity := "11"
		routesQuantity := "22"

		createQuotaOutput, err := kf.CreateQuota(ctx, quotaName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)

		AssertNil(t, "create quota error", err)
		AssertContainsAll(t, strings.Join(createQuotaOutput, " "), []string{quotaName, "successfully created"})
		Logf(t, "done ensuring quota was created.")

		// Get the quota
		getQuotaOutput, err := kf.GetQuota(ctx, quotaName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{quotaName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring correct quota info returned.")
	})
}

// TestIntegration_Delete creates a quota and then deletes it. It then makes
// sure the quota no longer exists.
func TestIntegration_Delete(t *testing.T) {
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		quotaName := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())

		memQuantity := "55Gi"
		cpuQuantity := "11"
		routesQuantity := "22"

		kf.CreateQuota(ctx, quotaName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)

		getQuotaOutput, err := kf.GetQuota(ctx, quotaName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{quotaName, memQuantity, cpuQuantity, routesQuantity})

		deleteQuotaOutput, err := kf.DeleteQuota(ctx, quotaName)
		AssertNil(t, "delete quota error", err)
		AssertContainsAll(t, strings.Join(deleteQuotaOutput, " "), []string{quotaName, "successfully deleted"})

		// Getting the quota should output an error saying the quota is not found
		getQuotaOutput, err = kf.GetQuota(ctx, quotaName)
		AssertNotNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, " "), []string{quotaName, "Error", "not found"})
	})
}

// TestIntegration_Update creates a quota then updates it.
// It checks that the quota is saved with the new values.
// Afterwards, the quota is deleted.
func TestIntegration_Update(t *testing.T) {
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		quotaName := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())
		defer kf.DeleteQuota(ctx, quotaName)

		memQuantity := "55Gi"
		cpuQuantity := "11"

		kf.CreateQuota(ctx, quotaName,
			"-m", memQuantity,
			"-c", cpuQuantity,
		)

		getQuotaOutput, err := kf.GetQuota(ctx, quotaName)
		AssertNil(t, "get quota error", err)

		// Check that routes quota appears by default as "0"
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{quotaName, memQuantity, cpuQuantity, "0"})
		Logf(t, "done ensuring correct quota info returned.")

		memQuantity = "66Gi"
		cpuQuantity = "22"
		routesQuantity := "8"

		_, err = kf.UpdateQuota(ctx, quotaName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)
		AssertNil(t, "update quota error", err)

		getQuotaOutput, err = kf.GetQuota(ctx, quotaName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{quotaName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring quota info updated.")

		// Reset memory quota
		memQuantity = "0"
		_, err = kf.UpdateQuota(ctx, quotaName,
			"-m", memQuantity,
		)
		AssertNil(t, "update quota error", err)

		getQuotaOutput, err = kf.GetQuota(ctx, quotaName)
		AssertNil(t, "get quota error", err)
		AssertContainsAll(t, strings.Join(getQuotaOutput, "\n"), []string{quotaName, memQuantity, cpuQuantity, routesQuantity})
		Logf(t, "done ensuring memory quota was reset.")
	})
}

// TestIntegration_List creates multiple quotas, then checks that all of them are listed.
// Afterwards, all of the quotas are deleted.
func TestIntegration_List(t *testing.T) {
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		quotaName := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())
		quota2Name := fmt.Sprintf("integration-quota-%d", time.Now().UnixNano())
		defer kf.DeleteQuota(ctx, quotaName)
		defer kf.DeleteQuota(ctx, quota2Name)

		memQuantity := "55Gi"
		cpuQuantity := "11"
		routesQuantity := "33"

		memQuantity2 := "44Gi"
		cpuQuantity2 := "6666m"
		routesQuantity2 := "22"

		kf.CreateQuota(ctx, quotaName,
			"-m", memQuantity,
			"-c", cpuQuantity,
			"-r", routesQuantity,
		)

		kf.CreateQuota(ctx, quota2Name,
			"-m", memQuantity2,
			"-c", cpuQuantity2,
			"-r", routesQuantity2,
		)

		listQuotasOutput, err := kf.Quotas(ctx)
		AssertNil(t, "list quotas error", err)
		AssertContainsAll(t, strings.Join(listQuotasOutput, "\n"), []string{
			quotaName, memQuantity, cpuQuantity, routesQuantity,
			quota2Name, memQuantity2, cpuQuantity2, routesQuantity2})
		Logf(t, "done ensuring info listed for all quotas.")
	})
}

var checkOnce sync.Once

func checkClusterStatus(t *testing.T) {
	checkOnce.Do(func() {
		testIntegration_Doctor(t)
	})
}

// testIntegration_Doctor runs the doctor command. It ensures the cluster the
// tests are running against is in good shape.
func testIntegration_Doctor(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		kf.Doctor(ctx)
	})
}
