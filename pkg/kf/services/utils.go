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
	"sort"

	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

// ParseJSONOrFile parses the value as JSON if it's valid or else it tries to
// read the value as a file on the filesystem.
func ParseJSONOrFile(jsonOrFile string) (json.RawMessage, error) {
	if json.Valid([]byte(jsonOrFile)) {
		return AssertJSONMap([]byte(jsonOrFile))
	}

	contents, err := ioutil.ReadFile(jsonOrFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file: %v", err)
	}

	result, err := AssertJSONMap(contents)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s as JSON: %v", jsonOrFile, err)
	}

	return result, nil
}

// AssertJSONMap asserts that the string is a JSON map.
func AssertJSONMap(jsonString []byte) (json.RawMessage, error) {
	p := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonString), &p); err != nil {
		return nil, fmt.Errorf("value must be a JSON map, got: %q", jsonString)
	}
	return json.RawMessage(jsonString), nil
}

// LastStatusCondition returns the last condition based on time.
func LastStatusCondition(si v1beta1.ServiceInstance) v1beta1.ServiceInstanceCondition {
	if len(si.Status.Conditions) == 0 {
		return v1beta1.ServiceInstanceCondition{
			Reason: "Unknown",
		}
	}

	sort.SliceStable(si.Status.Conditions, func(i, j int) bool {
		return si.Status.Conditions[j].LastTransitionTime.After(
			si.Status.Conditions[i].LastTransitionTime.Time,
		)
	})

	return si.Status.Conditions[len(si.Status.Conditions)-1]
}
