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
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/kf/v2/pkg/kf/internal/tools/includefile"
)

func main() {
	pkg := flag.String("pkg", "", "set the package")
	variable := flag.String("variable", "", "set the name of the include")
	filePath := flag.String("file", "", "file to include")
	flag.Parse()

	contents, err := ioutil.ReadFile(*filePath)
	if err != nil {
		log.Fatalln("could not read contents", err)
	}

	params := includefile.Params{
		Package:  *pkg,
		Variable: *variable,
		File:     *filePath,
		Contents: contents,
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

	f, err := os.Create(fmt.Sprintf("zz_generated.include.%s.go", strings.ToLower(*variable)))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(formatted)
}
