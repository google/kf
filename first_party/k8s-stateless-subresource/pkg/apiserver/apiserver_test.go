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

package apiserver_test

import (
	"context"
	"reflect"
	"testing"
)

func TestMainWithConfig_endpointReturnsHelloWorld(t *testing.T) {
	end2endTest(t, func(ctx context.Context) {
		t := testingFromContext(ctx)
		get(ctx, "/apis/helloworld.k8s.io/v1alpha1/namespaces/some-namespace/some-resource",
			func(m map[string]interface{}) bool {
				if actual, expected := m["hello"], "world"; !reflect.DeepEqual(actual, expected) {
					t.Fatalf("expected %v, got %v", expected, actual)
				}

				return true
			})
	})
}

func TestMainWithConfig_discovery(t *testing.T) {
	end2endTest(t, func(ctx context.Context) {
		t := testingFromContext(ctx)
		get(ctx, "/apis/helloworld.k8s.io/v1alpha1",
			func(m map[string]interface{}) bool {
				if actual, expected := m["name"], "Yyy.helloworld.k8s.io/ZZZ"; !reflect.DeepEqual(actual, expected) {
					t.Logf("expected %v, got %v", expected, actual)
					return false
				}
				return true
			})

	})
}
