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

package stringtools

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// Returns temp file name
func setupInFile(filenamePattern string, contents string) (string, error) {

	tempFile, err := ioutil.TempFile("", filenamePattern)
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary file for %q pattern", filenamePattern)
	}
	defer tempFile.Close()

	_, err = tempFile.WriteString(contents)
	if err != nil {
		return "", fmt.Errorf("Could not write data to file %q", tempFile.Name())
	}
	return tempFile.Name(), nil
}

func deleteFile(filename string) error {
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("Failed to remove temporary file")
	}
	return nil
}

func TestReplaceRegExpInFile(t *testing.T) {
	regexTestContents := "xyz\nabc\ndef"

	replaceRegexTests := []struct {
		name           string
		regex          string
		replacement    string
		expectContents []string
	}{
		{
			name:           "start of line",
			regex:          "x",
			replacement:    "q",
			expectContents: []string{"qyz", "abc", "def"},
		}, {
			name:           "end of line",
			regex:          "f",
			replacement:    "q",
			expectContents: []string{"xyz", "abc", "deq"},
		}, {
			name:           "middle of line",
			regex:          "b",
			replacement:    "q",
			expectContents: []string{"xyz", "aqc", "def"},
		},
	}
	for _, tt := range replaceRegexTests {
		pattern := "string_manip_test_in"
		infilename, err := setupInFile(pattern, regexTestContents)
		if err != nil {
			t.Fatalf("Couldn't create temporary file.  Test: %v", tt.name)
		}

		pattern = "string_manip_test_out"
		outfile, err := ioutil.TempFile("", pattern)
		if err != nil {
			log.Fatalf("Failed to create temporary file for %q pattern.  Test: %v", pattern, tt.name)
		}
		defer outfile.Close()
		defer deleteFile(outfile.Name())

		actualErr := ReplaceRegExpInFile(tt.regex, tt.replacement, infilename, outfile.Name())
		if actualErr != nil {
			t.Errorf("ReplaceRegExpInFile threw an error.  Test: %v", tt.name)
		}

		lines := ReadLines(outfile.Name())
		if err != nil {
			t.Fatalf("Couldn't read output file.  Test: %v", tt.name)
		}
		if len(lines) != len(tt.expectContents) {
			t.Errorf(
				"Actual output doesn't match expected output.  Expected: %v Actual: %v.  Test: %v",
				tt.expectContents, lines, tt.name)
		}
		for i := range lines {
			if lines[i] != tt.expectContents[i] {
				t.Errorf(
					"Actual output doesn't match expected output on line %v.  Expected: %v Actual: %v.  Test: %v",
					i, tt.expectContents, lines, tt.name)
			}
		}

	}
}
