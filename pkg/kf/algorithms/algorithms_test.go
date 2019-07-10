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

package algorithms_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/google/kf/pkg/kf/algorithms"
	"github.com/google/kf/pkg/kf/testutil"
)

func TestDedupe(t *testing.T) {
	t.Parallel()

	// We use 0 so we get the same tests every time.
	rand := rand.New(rand.NewSource(0))
	for i := 0; i < 5000; i++ {
		var slice algorithms.Ints
		for j := 0; j < rand.Intn(1000)+1000; j++ {
			slice = append(slice, rand.Intn(10))
		}

		slice = algorithms.Dedupe(algorithms.Ints(slice)).(algorithms.Ints)

		// While it might already be sorted, the requirements  of the
		// algorithm don't require it. We'll do it here to make testing it
		// easier (even if it's redundant).
		sort.Ints(slice)

		testutil.AssertEqual(t, "len", 10, len(slice))
		testutil.AssertEqual(t, "values", algorithms.Ints{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, slice)
	}
}

func ExampleDedupe() {
	s := []string{"a", "b", "a", "d", "b", "d", "c"}
	s = []string(algorithms.Dedupe(algorithms.Strings(s)).(algorithms.Strings))

	fmt.Println(strings.Join(s, ", "))

	// Outputs: a, b, c, d
}

func TestSearch(t *testing.T) {
	t.Parallel()

	// We use 0 so we get the same tests every time.
	rand := rand.New(rand.NewSource(0))
	for i := 0; i < 5000; i++ {
		var slice []int
		for j := 0; j < rand.Intn(1000); j++ {
			slice = append(slice, rand.Intn(1000))
		}

		sort.Sort(algorithms.Ints(slice))

		stdlibSearch := func(a []int, x int) bool {
			idx := sort.SearchInts(a, x)
			return idx < len(a) && a[idx] == x
		}

		value := rand.Intn(1000)
		testutil.AssertEqual(
			t,
			"found",
			stdlibSearch(slice, value),
			algorithms.Search(0, algorithms.Ints{value}, algorithms.Ints(slice)),
		)
	}

	a := algorithms.Strings([]string{"a", "x"})
	b := algorithms.Strings([]string{"b", "a", "c", "d"})
	sort.Sort(b)

	testutil.AssertEqual(t, "found", true, algorithms.Search(0, a, b))
	testutil.AssertEqual(t, "not found", false, algorithms.Search(1, a, b))
}

func ExampleSearch() {
	haystack := algorithms.Strings{"a", "c", "d", "b", "x"}
	sort.Sort(haystack)
	needles := algorithms.Strings{"x", "y", "z", "c"}

	for needleIdx := range needles {
		fmt.Println(algorithms.Search(needleIdx, needles, haystack))
	}

	// Output: true
	// false
	// false
	// true
}

func TestDelete(t *testing.T) {
	t.Parallel()

	a := algorithms.Strings{"c", "b", "a", "d"}
	b := algorithms.Strings{"d", "c"}

	a = algorithms.Delete(a, b).(algorithms.Strings)

	// While it might already be sorted, the requirements  of the
	// algorithm don't require it. We'll do it here to make testing it
	// easier (even if it's redundant).
	sort.Strings(a)

	testutil.AssertEqual(t, "len", 2, len(a))
	testutil.AssertEqual(t, "values", algorithms.Strings{"a", "b"}, a)
}

func ExampleDelete() {
	a := []string{"c", "b", "a", "d"}
	b := []string{"d", "c"}

	a = []string(algorithms.Delete(algorithms.Strings(a), algorithms.Strings(b)).(algorithms.Strings))

	// Sort for readability
	sort.Sort(algorithms.Strings(a))

	fmt.Println(strings.Join(a, ", "))

	// Outputs: a, b
}

func TestMerge(t *testing.T) {
	t.Parallel()

	a := algorithms.Strings{"c", "b", "a", "d"}
	b := algorithms.Strings{"d", "c", "e"}

	r := algorithms.Merge(a, b)

	sort.Sort(r)
	testutil.AssertEqual(t, "values", algorithms.Strings{"a", "b", "c", "d", "e"}, r.(algorithms.Strings))
}

func ExampleMerge() {
	a := algorithms.Strings{"c", "b", "a", "d"}
	b := algorithms.Strings{"d", "c", "e"}

	r := algorithms.Merge(a, b)

	// Sort for display purposes.
	sort.Sort(r)

	for _, x := range r.(algorithms.Strings) {
		fmt.Println(x)
	}

	// Output: a
	// b
	// c
	// d
	// e
}
