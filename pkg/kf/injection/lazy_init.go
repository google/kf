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

package injection

import (
	"sync"
)

// LazyInit is used as a uniformed way of lazily setting up clients. This is
// necessary so that fake packages can use the same type.
type LazyInit struct {
	once   sync.Once
	client interface{}
	init   func() interface{}
}

// NewLazyInit returns a LazyInit that invokes the given function once.
// Its return value should then be used in the injection package's Get
// function.
func NewLazyInit(f func() interface{}) *LazyInit {
	return &LazyInit{init: f}
}

// Create invokes the given Init function.
func (li *LazyInit) Create() interface{} {
	li.once.Do(func() {
		li.client = li.init()
	})
	return li.client
}
