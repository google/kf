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
	"testing"

	"github.com/golang/mock/gomock"
)

type controllerKey struct{}

func withController(ctx context.Context, t *testing.T) context.Context {
	// TODO(poy): Ideally we could use t.Cleanup() to invoke ctrl.Finish(). I
	// suspect this change will be large so do it later.
	ctrl := gomock.NewController(t)
	return context.WithValue(ctx, controllerKey{}, ctrl)
}

// GetController returns the gomock.Controller associated with this context.
func GetController(ctx context.Context) *gomock.Controller {
	return ctx.Value(controllerKey{}).(*gomock.Controller)
}
