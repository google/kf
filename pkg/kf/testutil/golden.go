// Copyright 2020 Google LLC
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

package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/segmentio/textio"
	"sigs.k8s.io/yaml"
)

func updatingGolden() bool {
	return os.Getenv("UPDATE_GOLDEN") == "true"
}

func goldenPath(name, subtest string) string {
	filename := fmt.Sprintf("%s_%s.golden", cleanName(name), cleanName(subtest))

	return filepath.Join("testdata", "golden", filename)
}

func cleanName(name string) string {
	lower := strings.ToLower(name)

	// Strip out anything that could be used by the filesystem and replace with
	// underscores.
	split := strings.FieldsFunc(lower, func(r rune) bool {
		okay := unicode.IsDigit(r) || unicode.IsLower(r)
		return !okay
	})

	return strings.Join(split, "_")
}

// AssertGoldenJSONContext is like AssertGoldenJSON but prepends contextual
// test information to the golden file.
func AssertGoldenJSONContext(t Failable, fieldName string, object interface{}, context map[string]interface{}) {
	results, err := json.MarshalIndent(object, "", "    ")
	AssertNil(t, "json.Serialize error", err)

	AssertGoldenContext(t, fieldName, results, context)
}

// AssertGoldenJSON asserts the given test's field matches the .golden file in
// testdata after the field has been serialized to JSON.
func AssertGoldenJSON(t Failable, fieldName string, object interface{}) {
	results, err := json.MarshalIndent(object, "", "    ")
	AssertNil(t, "json.Serialize error", err)

	AssertGolden(t, fieldName, results)
}

// AssertGoldenContext adds contextual info to a header of hash marks at the
// start of the .golden file.
func AssertGoldenContext(t Failable, fieldName string, actualValue []byte, context map[string]interface{}) {
	t.Helper()

	tmp := &bytes.Buffer{}
	// Write a header with context info.
	{
		header := textio.NewPrefixWriter(tmp, "# ")
		fmt.Fprintf(header, "Test:\t%s\n", t.Name())
		results, err := yaml.Marshal(context)
		AssertNil(t, "context yaml.Marshal error", err)
		header.Write(results)
		header.Flush()
	}

	fmt.Fprintln(tmp)
	tmp.Write(actualValue)

	AssertGolden(t, fieldName, tmp.Bytes())
}

// AssertGolden asserts the given test's field matches the .golden file in
// testdata.
func AssertGolden(t Failable, fieldName string, bytes []byte) {
	t.Helper()

	path := goldenPath(t.Name(), fieldName)

	if updatingGolden() {
		parent := filepath.Dir(path)
		if err := os.MkdirAll(parent, 0755); err != nil && !os.IsExist(err) {
			AssertNil(t, "creating golden path file error", err)
		}

		err := ioutil.WriteFile(path, bytes, 0644)
		AssertNil(t, "golden path update error", err)

		t.Fatalf("Updated golden file, re-run without UPDATE_GOLDEN to test")
	}

	contents, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		t.Fatalf("Missing golden file, run with UPDATE_GOLDEN=true to generate")
		return
	}
	AssertNil(t, "error reading golden file", err)

	AssertEqual(t, fmt.Sprintf("golden results in %s for %s", path, fieldName), string(contents), string(bytes))
}
