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

package apps

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// NewHealthCheck creates a corev1.Probe that maps the health checks CloudFoundry
// does.
func NewHealthCheck(healthCheckType string, endpoint string, timeoutSeconds int) (*corev1.Probe, error) {
	if timeoutSeconds < 0 {
		return nil, errors.New("health check timeouts can't be negative")
	}

	probe := &corev1.Probe{TimeoutSeconds: int32(timeoutSeconds)}

	switch healthCheckType {
	case "http":
		probe.Handler.HTTPGet = &corev1.HTTPGetAction{Path: endpoint}
		return probe, nil

	case "port", "": // By default, cf uses a port based health check.
		if endpoint != "" {
			return nil, errors.New("health check endpoints can only be used with http checks")
		}

		probe.Handler.TCPSocket = &corev1.TCPSocketAction{}
		return probe, nil

	case "process", "none":
		return nil, errors.New("kf doesn't support the process health check type")

	default:
		return nil, fmt.Errorf("unknown health check type %s, supported types are http and port", healthCheckType)
	}
}
