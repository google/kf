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

package routeutil_test

import (
	"math/rand"
	"regexp"
	"testing"

	"github.com/google/kf/pkg/kf/internal/routeutil"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestEncodeRouteName_Deterministic(t *testing.T) {
	t.Parallel()

	r1 := routeutil.EncodeRouteName("host-1", "example1.com", "somePath1")
	r2 := routeutil.EncodeRouteName("host-1", "example1.com", "somePath1")
	r3 := routeutil.EncodeRouteName("host-2", "example1.com", "somePath1")
	r4 := routeutil.EncodeRouteName("host-1", "example2.com", "somePath1")
	r5 := routeutil.EncodeRouteName("host-1", "example1.com", "somePath2")

	testutil.AssertEqual(t, "r1 and r2", r1, r2)
	testutil.AssertEqual(t, "r1 and r2", r1, r2)

	for _, r := range []string{r3, r4, r5} {
		if r1 == r {
			t.Fatalf("expected %s to not equal %s", r, r1)
		}
	}
}

func TestEncodeRouteName_ValidDNS(t *testing.T) {
	t.Parallel()

	// We'll use an instantiation of rand so we can seed it with 0 for
	// repeatable tests.
	rand := rand.New(rand.NewSource(0))
	randStr := func() string {
		buf := make([]byte, rand.Intn(19)+1)
		for i := range buf {
			buf[i] = byte(rand.Intn('z'-'a') + 'a')
		}
		return string(buf)
	}

	pattern := regexp.MustCompile(`[^a-z0-9-_]`)
	history := map[string]bool{}

	// Basically we're going to try a mess of different things to ensure that
	// certain rules are followed:
	// [a-z0-9_-]
	for i := 0; i < 10000; i++ {
		r := routeutil.EncodeRouteName(randStr(), randStr(), randStr())
		testutil.AssertEqual(t, "invalid rune: "+r, false, pattern.MatchString(r))

		testutil.AssertEqual(t, "collison", false, history[r])
		history[r] = true
	}
}
