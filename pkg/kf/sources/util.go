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

package sources

import (
	"context"
	"fmt"
	"io"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// SourceStatus gets the status of the given source.
// Complete will be set to true if the build has completed (or doesn't exist).
// Error will be set if the build completed with an error (or doesn't exist).
// A successful result is one that completed and error is nil.
func SourceStatus(source v1alpha1.Source) (finished bool, err error) {
	condition := source.Status.GetCondition(v1alpha1.SourceConditionSucceeded)
	if condition == nil {
		// no success condition means the build hasn't propigated yet
		return false, nil
	}

	switch condition.Status {
	case corev1.ConditionTrue: // the build was a success
		return true, nil

	case corev1.ConditionFalse: // the build was a failure
		return true, fmt.Errorf("build failed for reason: %s with message: %s", condition.Reason, condition.Message)

	default: // the build is in a transition state
		return false, nil
	}
}

// BuildTailerFunc converts a func into a BuildTailer.
type BuildTailerFunc func(ctx context.Context, out io.Writer, buildName, namespace string) error

// Tail implements BuildTailer.
func (f BuildTailerFunc) Tail(ctx context.Context, out io.Writer, buildName, namespace string) error {
	return f(ctx, out, buildName, namespace)
}
