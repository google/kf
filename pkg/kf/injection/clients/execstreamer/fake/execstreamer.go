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

package fake

import (
	"context"

	"github.com/google/kf/v2/pkg/kf/injection/clients/execstreamer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"knative.dev/pkg/injection"
)

func init() {
	injection.Fake.RegisterClient(withExecStreamer)
}

// Get returns the Fake that is injected into the context.
func Get(ctx context.Context) *Fake {
	return ctx.Value(execstreamer.Key{}).(*Fake)
}

func withExecStreamer(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, execstreamer.Key{}, &Fake{
		StreamF: func(
			ctx context.Context,
			podSelector metav1.ListOptions,
			execOpts corev1.PodExecOptions,
			streamOpts remotecommand.StreamOptions,
		) error {
			// NOP
			return nil
		},
	})
}

// Fake is a fake ExecStreamer. The StreamF method should be replaced.
type Fake struct {
	StreamF func(
		ctx context.Context,
		podSelector metav1.ListOptions,
		execOpts corev1.PodExecOptions,
		streamOpts remotecommand.StreamOptions,
	) error
}

func (f *Fake) Stream(
	ctx context.Context,
	podSelector metav1.ListOptions,
	execOpts corev1.PodExecOptions,
	streamOpts remotecommand.StreamOptions,
) error {
	return f.StreamF(ctx, podSelector, execOpts, streamOpts)
}
