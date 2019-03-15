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
