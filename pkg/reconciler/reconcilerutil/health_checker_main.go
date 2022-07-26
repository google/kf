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

package reconcilerutil

import (
	"context"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
)

// HealthChecker has a Healthy method that gets invoked periodically to
// determine the health of the controller.
type HealthChecker interface {
	Healthy(ctx context.Context) error
}

// HealthCheckerMain wraps sharedmain.Main and therefore blocks until the
// controller is shutdown. The main difference is it starts a health endpoint
// on addr. Any constructor that implements HealthChecker will be used to
// determine if the endpoint returns a 204 or a 503.
func HealthCheckerMain(ctx context.Context, addr, component string, ctors ...injection.ControllerConstructor) {
	var wrappers []injection.ControllerConstructor
	hcs := sync.Map{}

	// We need to wait for each controller to be constructed and checked
	// before the health endpoint will be ready.
	ctorsWg := sync.WaitGroup{}
	ctorsWg.Add(len(ctors))

	// Find all the HealthCheckers
	for i := range ctors {
		ctor := ctors[i]

		// Wrap the given ControllerConstructor so that the resulting
		// controller can be check to see if it implements HealthChecker. If
		// it does, then keep track of it so that its health can be polled.
		wrappers = append(wrappers, func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
			ctorsWg.Done()
			ctrl := ctor(ctx, cmw)
			if hc, ok := ctrl.Reconciler.(HealthChecker); ok {
				hcs.Store(hc, ctx)
			}
			return ctrl
		})
	}

	// Poll the HealthCheckers. If there is an empty list of HealthCheckers,
	// then default to healthy. Otherwise default to unhealthy so that the
	// status starts out as not ready.
	healthyInt := int32(0)
	go func() {
		ctorsWg.Wait()

		// Poll HealthCheckers
		for range time.Tick(5 * time.Second) {
			// Check health
			healthy := int32(1)
			hcs.Range(func(key, value interface{}) bool {
				// Don't type check because we want this to panic if this gets
				// messed up from earlier in the function.
				hc := key.(HealthChecker)
				ctx, cancel := context.WithTimeout(value.(context.Context), 30*time.Second)
				defer cancel()

				if err := hc.Healthy(ctx); err != nil {
					logging.FromContext(ctx).Warnf("health check failed: %v", err)
					healthy = 0
				}

				// Only continue if we're still healthy
				return healthy == 1
			})

			// Update status
			atomic.StoreInt32(&healthyInt, healthy)
		}
	}()

	// Start the health check endpoint on the given address.
	go func() {
		ctorsWg.Wait()

		log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if atomic.LoadInt32(&healthyInt) != 1 {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			},
		)))
	}()

	sharedmain.MainWithContext(ctx, component, wrappers...)
}
