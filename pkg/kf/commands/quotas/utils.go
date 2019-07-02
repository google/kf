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
	"fmt"

	"github.com/google/kf/pkg/kf/quotas"
	"k8s.io/apimachinery/pkg/api/resource"
)

// setQuotaValues updates a KfQuota to have the inputted resource quota values.
func setQuotaValues(memory string, cpu string, routes string, kfquota *quotas.KfQuota) error {
	var quotaInputs = []struct {
		Value    string
		Setter   func(r resource.Quantity)
		Resetter func()
	}{
		{memory, kfquota.SetMemory, kfquota.ResetMemory},
		{cpu, kfquota.SetCPU, kfquota.ResetCPU},
		{routes, kfquota.SetServices, kfquota.ResetServices},
	}

	// Only update resource quotas for inputted flags
	for _, quota := range quotaInputs {
		if quota.Value != defaultQuota {
			quantity, err := resource.ParseQuantity(quota.Value)
			if err != nil {
				return fmt.Errorf("couldn't parse resource quantity %s: %v", quota.Value, err)
			}
			// Passing in 0 for a resource resets its quota to unlimited
			if quantity.IsZero() {
				quota.Resetter()
			} else {
				quota.Setter(quantity)
			}
		}
	}
	return nil
}
