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

package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ParseJSONOrFile parses the value as JSON if it's valid or else it tries to
// read the value as a file on the filesystem.
func ParseJSONOrFile(jsonOrFile string) (map[string]interface{}, error) {
	if json.Valid([]byte(jsonOrFile)) {
		return ParseJSONString(jsonOrFile)
	}

	contents, err := ioutil.ReadFile(jsonOrFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file: %v", err)
	}

	result, err := ParseJSONString(string(contents))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s as JSON: %v", jsonOrFile, err)
	}

	return result, nil
}

// ParseJSONString converts a string of JSON to a Go map.
func ParseJSONString(jsonString string) (map[string]interface{}, error) {
	p := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonString), &p); err != nil {
		return nil, fmt.Errorf("invalid JSON provided: %q", jsonString)
	}
	return p, nil
}
