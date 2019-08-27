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

package system

import (
	"log"
	"os"
)

const (
	// NamespaceEnvKey is the environment variable that holds kf's namespace.
	NamespaceEnvKey = "SYSTEM_NAMESPACE"

	// KnativeServingNamespaceEnvKey is the environment variable that holds
	// the knative serving namespace.
	KnativeServingNamespaceEnvKey = "KNATIVE_SERVING_NAMESPACE"
)

// Namespace holds the K8s namespace where our serving system
// components run.
func Namespace() string {
	if ns := os.Getenv(NamespaceEnvKey); ns != "" {
		return ns
	}

	log.Fatalf(`The environment variable %q is not set
If this is a process running on Kubernetes, then it should be using the downward
API to initialize this variable via:
  env:
  - name: %s
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
If this is a Go unit test consuming serverside.Namespace() then it should add the
following import:
import (
	_ "github.com/google/kf/pkg/serverside/testing"
)`, NamespaceEnvKey, NamespaceEnvKey)

	// log.Fatalf should terminate the process but go doesn't recognize that
	panic("log.Fatalf should have crashed the process")
}

// KnativeServingNamespace holds the K8s namespace where our serving system
// components run.
func KnativeServingNamespace() string {
	if ns := os.Getenv(KnativeServingNamespaceEnvKey); ns != "" {
		return ns
	}

	log.Fatalf(`The environment variable %q is not set
If this is a process running on Kubernetes, then it should be using the downward
API to initialize this variable via:
  env:
  - name: %s
    value: knative-serving
If this is a Go unit test consuming serverside.KnativeServingNamespace() then
it should add the following import:
import (
	_ "github.com/google/kf/pkg/serverside/testing"
)`, KnativeServingNamespaceEnvKey, KnativeServingNamespaceEnvKey)

	// log.Fatalf should terminate the process but go doesn't recognize that
	panic("log.Fatalf should have crashed the process")
}
