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

package dynamichelper

import (
	context "context"
	"flag"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

var (
	directoryRefreshRate time.Duration
)

func init() {
	flag.DurationVar(&directoryRefreshRate, "directory_refresh_rate", 30*time.Second, "Period with which to refresh failed directory lookups.")
}

// See injection lib for interface def.
type dynamicHelper struct {
	dynamic.Interface
	*restmapper.DeferredDiscoveryRESTMapper
}

// CreateDynamicHelper creates an dynamichelper by extracting the dynamicclient
// from the context given, and using the provided k8s interface (which cannot come from
// the context due to type constraints in knative injection).
func CreateDynamicHelper(ctx context.Context, kc kubernetes.Interface, dc dynamic.Interface) Interface {
	memClient := memory.NewMemCacheClient(kc.Discovery())
	ticker := time.NewTicker(directoryRefreshRate)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				memClient.Invalidate()
			}
		}
	}()
	return &dynamicHelper{
		dc,
		restmapper.NewDeferredDiscoveryRESTMapper(memClient),
	}
}

func (d dynamicHelper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return d.DeferredDiscoveryRESTMapper.KindFor(resource)
}

func (d dynamicHelper) Lookup(kind schema.GroupKind, namespace string, versions ...string) (dynamic.ResourceInterface, error) {
	m, err := d.RESTMapping(kind, versions...)
	if err != nil {
		return nil, err
	}
	c := d.Resource(m.Resource)
	if namespace != "" {
		return c.Namespace(namespace), nil
	}
	return c, nil
}
