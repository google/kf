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

//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"

	fakelister "github.com/google/kf/v2/pkg/kf/internal/tools/fakelister"
)

func main() {
	pkg := flag.String("pkg", "", "set the package")
	objectType := flag.String("object-type", "", "set the object type (e.g., Topic")
	objectPkg := flag.String("object-pkg", "", "set the object package (e.g., github.com/google/kueue/pkg/apis/kueue/v1alpha1)")
	listerPkg := flag.String("lister-pkg", "", "set the lister package (e.g., github.com/google/kueue/pkg/client/kueue/listers/kueue/v1alpha1)")
	namespaced := flag.Bool("namespaced", true, "namespaced object")
	flag.Parse()

	params := fakelister.Params{
		Package:       *pkg,
		ObjectType:    *objectType,
		ObjectPackage: *objectPkg,
		ListerPackage: *listerPkg,
		Namespaced:    *namespaced,
	}

	buf, err := params.Render()
	if err != nil {
		log.Fatal(err)
		return
	}

	formatted, err := format.Source(buf)
	if err != nil {
		log.Fatal(err)
		return
	}

	f, err := os.Create(fmt.Sprintf("zz_generated.fakelister.%s_test.go", strings.ToLower(*objectType)))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(formatted)
}
