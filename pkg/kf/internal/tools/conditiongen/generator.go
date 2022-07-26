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

	"github.com/google/kf/v2/pkg/kf/internal/tools/conditiongen"
)

func main() {
	pkg := flag.String("pkg", "", "set the package")
	prefix := flag.String("prefix", "", "set the condition prefix")
	statusType := flag.String("status-type", "", "set the status type")
	batch := flag.Bool("batch", false, "set the roll-up status type to Succeeded rather than Ready")
	flag.Parse()
	conds := flag.Args()

	params := conditiongen.Params{
		Package:            *pkg,
		ConditionPrefix:    *prefix,
		StatusType:         *statusType,
		Conditions:         conds,
		LivingConditionSet: !(*batch),
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

	f, err := os.Create(fmt.Sprintf("zz_generated.conditions.%s.go", strings.ToLower(*statusType)))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(formatted)
}
