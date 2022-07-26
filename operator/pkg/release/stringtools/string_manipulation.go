/*
Copyright 2020 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package stringtools

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

// ReplaceRegExpInFile - Replace regular expression matches to a literal string from inputFile to outputFile
func ReplaceRegExpInFile(template string, repl string, inputFile string, outputFile string) error {
	b, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("Could not read file %q", inputFile)
	}
	s := string(b)

	re, err := regexp.CompilePOSIX(template)
	if err != nil {
		return err
	}
	result := re.ReplaceAllString(s, repl)

	if err := ioutil.WriteFile(outputFile, []byte(result), 0644); err != nil {
		return err
	}
	return nil
}

// ReadLines - Read file and return list of lines contained
func ReadLines(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Could not open file %q", filePath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

// CountLines returns number of lines of file in filePath.
func CountLines(filePath string) int {
	return len(ReadLines(filePath))
}
