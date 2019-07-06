/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/helper"
)

func main() {
	path := os.Args[1]

	code, err := p(path)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}

	os.Exit(code)
}

func p(path string) (int, error) {
	credentials, ok, err := helper.FindServiceCredentials("google-stackdriver-debugger", "PrivateKeyData")
	if err != nil {
		return 1, err
	}

	if !ok {
		credentials, ok, err = helper.FindServiceCredentials("google-stackdriver-profiler", "PrivateKeyData")
		if err != nil {
			return 1, err
		}
	}

	if !ok {
		return 0, nil
	}

	data, ok := credentials["PrivateKeyData"].(string)
	if !ok {
		return 1, fmt.Errorf("PrivateKeyData is not a Base64 encoded string")
	}

	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 1, err
	}

	if err = ioutil.WriteFile(path, b, 0755); err != nil {
		return 1, err
	}

	return 0, nil
}
