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

package algorithms

import (
	"sort"
)

// Interface is the interface used by the algorithms.
type Interface interface {
	sort.Interface

	// Set stores the value located at b[j] to a[i].
	Set(i int, a Interface, j int, b Interface)

	// Slice returns a slice (e.g., s[i:j]) of the object.
	Slice(i, j int) Interface

	// Append returns i and a appended to each other (e.g., append(i, a...)).
	Append(a Interface) Interface

	// Clone returns a clone of the object.
	Clone() Interface
}

// Strings implements Interface for a slice of strings.
type Strings []string

// Len implements Interface.
func (s Strings) Len() int {
	return len(s)
}

// Swap implements Interface.
func (s Strings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements Interface.
func (s Strings) Less(i, j int) bool {
	return s[i] < s[j]
}

// Set implements Interface.
func (s Strings) Set(i int, a Interface, j int, b Interface) {
	a.(Strings)[i] = b.(Strings)[j]
}

// Clone implements Interface.
func (s Strings) Clone() Interface {
	return append(Strings{}, s...)
}

// Append implements Interface.
func (s Strings) Append(a Interface) Interface {
	return append(s, a.(Strings)...)
}

// Slice implements Interface.
func (s Strings) Slice(i, j int) Interface {
	return s[i:j]
}

// Ints implements Interface for a slice of ints.
type Ints []int

// Len implements Interface.
func (s Ints) Len() int {
	return len(s)
}

// Swap implements Interface.
func (s Ints) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements Interface.
func (s Ints) Less(i, j int) bool {
	return s[i] < s[j]
}

// Set implements Interface.
func (s Ints) Set(i int, a Interface, j int, b Interface) {
	a.(Ints)[i] = b.(Ints)[j]
}

// Clone implements Interface.
func (s Ints) Clone() Interface {
	return append(Ints{}, s...)
}

// Append implements Interface.
func (s Ints) Append(a Interface) Interface {
	return append(s, a.(Ints)...)
}

// Slice implements Interface.
func (s Ints) Slice(i, j int) Interface {
	return s[i:j]
}

// Dedupe removes any duplicates from the given collection. This does not
// alter the input. First element is always chosen.
func Dedupe(s Interface) Interface {
	s = s.Clone()
	sort.Stable(s)

	var idx int

	for i := 0; i < s.Len(); i++ {
		// We know that the interface is ascending order. Therefore, each
		// value should be greater than the one previous. If they are equal,
		// then they are not greater.
		if i != 0 && !greater(i, i-1, s) {
			continue
		}
		s.Set(idx, s, i, s)
		idx++
	}
	return s.Slice(0, idx)
}

// Delete removes any items in 'b' from the 'a' collection. This does not
// alter the input. It outputs the lengh of the new collection. Therefore, if
// the interface is wrapping a slice, then the slice should be truncated via
// the result (e.g., slice[:returnValue]).
func Delete(a, b Interface) Interface {
	a = a.Clone()
	b = b.Clone()
	sort.Sort(b)

	var currentIdx int
	for i := 0; i < a.Len(); i++ {
		if Search(i, a, b) {
			continue
		}
		a.Set(currentIdx, a, i, a)
		currentIdx++
	}

	return a.Slice(0, currentIdx)
}

func equal(i, j int, s sort.Interface) bool {
	// If something is neither less, nor greater than the given value, then it
	// must be equal.
	return !s.Less(i, j) && !s.Less(j, i)
}

func greater(i, j int, s sort.Interface) bool {
	// sort.Interface only gives us less. Therefore, if we reverse that
	// operation (via sort.Reverse), we get greater.
	return sort.Reverse(s).Less(i, j)
}

// Search will look for the item at a[i] in b. If found, it will return true.
// Otherwise it will return false.
func Search(i int, a, b Interface) bool {
	// b must be sorted, but we don't want to destroy the input.
	b = b.Clone()
	sort.Sort(b)

	idx := index(i, a, b)

	if idx >= b.Len() {
		return false
	}

	n := a.Clone().Slice(i, i+1).Append(b.Slice(idx, idx+1))
	return equal(0, 1, n)
}

func index(i int, a, b Interface) int {
	return sort.Search(b.Len(), func(j int) bool {
		n := b.Clone().Slice(j, j+1).Append(a.Slice(i, i+1))
		return greater(0, 1, n) || equal(0, 1, n)
	})
}

// Merge will combine the two collections. It replaces values from a with b if
// there is a collision. It assumes both a and b have been Deduped. It uses
// Append to create new memory and therefore does not destroy the input.
func Merge(a, b Interface) Interface {
	b = b.Clone()
	sort.Sort(b)

	return Dedupe(b.Append(a))
}
