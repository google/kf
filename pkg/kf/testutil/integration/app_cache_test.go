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
	context "context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestAppCache(t *testing.T) {
	t.Parallel()

	ac := newAppCache()
	t.Cleanup(func() {
		testutil.AssertErrorsEqual(t, nil, ac.Close())
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Initial storage.
	v, ok, err := ac.Load(ctx, "a")
	testutil.AssertFalse(t, "loaded", ok)
	testutil.AssertErrorsEqual(t, nil, ac.Store(ctx, "a", "1"))
	testutil.AssertErrorsEqual(t, nil, err)

	// Store it again and load.
	testutil.AssertErrorsEqual(t, nil, ac.Store(ctx, "a", "2"))
	v, ok, err = ac.Load(ctx, "a")
	testutil.AssertTrue(t, "loaded", ok)
	testutil.AssertErrorsEqual(t, nil, err)
	testutil.AssertEqual(t, "value", v, "1")

	// Store new.
	testutil.AssertErrorsEqual(t, nil, ac.Store(ctx, "b", "3"))
	v, ok, err = ac.Load(ctx, "b")
	testutil.AssertTrue(t, "loaded", ok)
	testutil.AssertErrorsEqual(t, nil, err)
	testutil.AssertEqual(t, "value", v, "3")

	// Cancelled context returns error.
	cancel()
	_, _, err = ac.Load(ctx, "b")
	testutil.AssertErrorsEqual(t, errors.New("context canceled"), err)
}

func TestAppCache_race(t *testing.T) {
	t.Parallel()

	ac := newAppCache()
	t.Cleanup(func() {
		testutil.AssertErrorsEqual(t, nil, ac.Close())
	})

	go func() {
		for i := 0; i < 100; i++ {
			ac.Load(context.Background(), fmt.Sprintf("%d", i))
		}
	}()

	for i := 0; i < 100; i++ {
		ac.Store(context.Background(), fmt.Sprintf("%d", i), "1")
	}
}
