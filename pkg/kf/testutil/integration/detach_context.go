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

package integration

import (
	"context"
	"time"
)

// DetachContext returns a context from the given context that does not have
// the underlying deadline, error, or done channel. It JUST takes the
// underlying values.
func DetachContext(ctx context.Context) context.Context {
	return &detachedContext{ctx: ctx}
}

type detachedContext struct {
	ctx context.Context
}

// Deadline implements context.Context.
func (d *detachedContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

// Done implements context.Context.
func (d *detachedContext) Done() <-chan struct{} {
	return nil
}

// Done implements context.Context.
func (d *detachedContext) Err() error {
	return nil
}

// Value implements context.Context.
func (d *detachedContext) Value(key interface{}) interface{} {
	return d.ctx.Value(key)
}
