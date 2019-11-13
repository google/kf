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

package generator

import (
	"fmt"
	"strings"
	"time"
)

func ExampleGenImports() {
	imports := GenImports(map[string]string{
		"os":                 "",
		"some/custom/import": "alias",
	})

	fmt.Println(imports)

	// Output: import (
	// 	"os"
	// 	alias "some/custom/import"
	// )
}

func ExampleGenImports_empty() {
	imports := GenImports(map[string]string{})

	fmt.Println(imports)

	// Output:
}

func ExampleGenNotice() {
	fmt.Println(GenNotice("some-file.go"))

	// Output: // This file was generated with some-file.go, DO NOT EDIT IT.
}

func ExampleGenLicense() {
	lic, err := GenLicense()
	if err != nil {
		panic(err)
	}

	currentYear := fmt.Sprintf("%d", time.Now().Year())

	fmt.Println("license contains year?", strings.Contains(lic, currentYear))
	fmt.Println("license contains apache text?", strings.Contains(lic, "Licensed under the Apache License, Version 2.0"))

	// Output: license contains year? true
	// license contains apache text? true
}
