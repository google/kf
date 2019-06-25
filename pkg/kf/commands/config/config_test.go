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

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

func ExampleKfParams_ApplyDefaults() {
	original := &KfParams{
		Config:    "my-custom-config",
		Namespace: "",
	}

	defaults := KfParams{
		Config:    "default-config-path",
		Namespace: "default",
	}

	original.ApplyDefaults(defaults)

	fmt.Println("Config:", original.Config)
	fmt.Println("Namespace:", original.Namespace)

	// Output: Config: my-custom-config
	// Namespace: default
}

func ExampleKfParams_ConfigPath() {
	params := &KfParams{}
	userHome, _ := homedir.Dir()
	defaultPath := strings.ReplaceAll(params.ConfigPath(), userHome, "$HOME")

	fmt.Println("Default:", defaultPath)

	params.Config = "some-custom-path.yaml"
	fmt.Println("Custom:", params.ConfigPath())

	// Output: Default: $HOME/.kf
	// Custom: some-custom-path.yaml
}

func ExampleKfParams_ReadConfig() {
	dir, err := ioutil.TempDir("", "kfcfg")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	configFile := path.Join(dir, "kf.yaml")

	{
		toWrite := &KfParams{
			Config:    configFile,
			Namespace: "my-namespace",
		}

		if err := toWrite.WriteConfig(); err != nil {
			panic(err)
		}
	}

	{
		toRead := &KfParams{
			Config: configFile,
		}

		if err := toRead.ReadConfig(); err != nil {
			panic(err)
		}

		fmt.Println("Read namespace:", toRead.Namespace)
	}

	// Output: Read namespace: my-namespace
}
