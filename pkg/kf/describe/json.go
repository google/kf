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

package describe

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"unicode"
)

// JSONData prints out a JSON struct.
func JSONData(w io.Writer, data []byte) error {
	deserialized := make(map[string]interface{})
	if err := json.Unmarshal(data, &deserialized); err != nil {
		return err
	}

	printJSONObject(w, deserialized)
	return nil
}

func printJSONObject(w io.Writer, obj map[string]interface{}) {
	TabbedWriter(w, func(w io.Writer) {
		var keys []string
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, rawKey := range keys {
			rawVal := obj[rawKey]
			key := JSONKeyToTitleCase(rawKey)

			switch val := rawVal.(type) {
			case map[string]interface{}:
				SectionWriter(w, key, func(w io.Writer) {
					printJSONObject(w, val)
				})
			case []interface{}:
				SectionWriter(w, key, func(w io.Writer) {
					printJSONArray(w, val)
				})
			default:
				fmt.Fprintf(w, "%s:\t%v\n", key, val)
			}
		}
	})
}

func printJSONArray(w io.Writer, anArray []interface{}) {
	for _, rawVal := range anArray {
		switch val := rawVal.(type) {
		case map[string]interface{}:
			printJSONObject(w, val)
		case []interface{}:
			printJSONArray(w, val)
		default:
			fmt.Fprintf(w, "%v\n", val)
		}
	}
}

func JSONKeyToTitleCase(name string) string {
	out := ""

	var last rune
	for _, curr := range name {
		insertBreak := false
		switch {
		case unicode.IsLower(curr):
			// lower only indicates a new word if from a digit
			// e.g. CSS3andJS
			insertBreak = unicode.IsNumber(last)
		case unicode.IsUpper(curr):
			// upper case only gets broken if last was not upper
			insertBreak = !unicode.IsUpper(last)
		case unicode.IsDigit(curr):
			// digits get broken from lower but not from other digits or upper
			// e.g. HTML5
			insertBreak = unicode.IsLower(last)
		default:
			return name
		}

		// First character is always capitalized and doesn't
		// get a break.
		if last == rune(0) {
			insertBreak = false
			curr = unicode.ToUpper(curr)
		}

		if insertBreak {
			out += " "
			curr = unicode.ToUpper(curr)
		}
		out += string(curr)
		last = curr
	}

	return out
}
