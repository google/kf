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

// +build ignore

package main

import (
	"go/format"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/kf/pkg/kf/internal/tools/clientgen"
	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("use: genclient.go /path/to/defn.yml")
	}

	optionsPath := os.Args[1]

	contents, err := ioutil.ReadFile(optionsPath)
	if err != nil {
		log.Fatal(err)
	}

	params := &clientgen.ClientParams{}
	if err := yaml.Unmarshal(contents, params); err != nil {
		log.Fatal(err)
		return
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

	f, err := os.Create("zz_generated.client.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(formatted)
}
