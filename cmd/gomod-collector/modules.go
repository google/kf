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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Module struct {
	Module         string
	ModuleVersion  string
	Replace        string
	ReplaceVersion string

	VendorDir string
}

func (m *Module) RealModule() string {
	if m.Replace != "" {
		return m.Replace
	}

	return m.Module
}

func (m *Module) RealVersion() string {
	if m.Replace != "" {
		return m.ReplaceVersion
	}

	return m.ModuleVersion
}

func (m *Module) Description() string {
	orig := fmt.Sprintf("%s %s", m.Module, m.ModuleVersion)

	if m.Replace != "" {
		return fmt.Sprintf("%s %s (imported as %s)", m.Replace, m.ReplaceVersion, orig)
	}

	return orig
}

// Source contains the source of the package.
func (m *Module) Source() string {
	return filepath.Join(m.VendorDir, filepath.FromSlash(m.Module))
}

func CollectVendoredModules(projectDirectories []string) ([]Module, error) {
	var vendored []Module

	for _, projectDir := range projectDirectories {
		vendorDir := filepath.Join(projectDir, "vendor")

		text, err := ioutil.ReadFile(filepath.Join(vendorDir, "modules.txt"))
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(text), "\n")

		for _, line := range lines {
			// Module lines take one of three forms.
			// Direct import:
			// # github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c
			// Replacement import:
			// # github.com/google/go-cmp v0.2.0 => github.com/google/go-cmp v0.3.0
			// Replacement local:
			// # github.com/google/kf/pkg/kf/commands/install v0.0.0 => ./pkg/kf/commands/install
			if !isModImport(line) {
				continue
			}

			module, moduleVersion, replace, replaceVersion := parseModImport(line)

			mod := Module{
				Module:         module,
				ModuleVersion:  moduleVersion,
				Replace:        replace,
				ReplaceVersion: replaceVersion,

				VendorDir: vendorDir,
			}

			// Skip if not in vendor
			if entries, err := ioutil.ReadDir(mod.Source()); os.IsNotExist(err) || len(entries) == 0 {
				continue
			}

			vendored = append(vendored, mod)
		}
	}

	return vendored, nil
}

func isModImport(line string) bool {
	return strings.HasPrefix(line, "#")
}

func parseModImport(line string) (module, moduleVersion, replace, replaceVersion string) {
	trimmed := strings.TrimPrefix(line, "# ")

	out := strings.SplitN(trimmed, "=>", 2)

	module, moduleVersion = parseModVersion(strings.TrimSpace(out[0]))

	if len(out) == 2 {
		replace, replaceVersion = parseModVersion(strings.TrimSpace(out[1]))
	}

	return
}

func parseModVersion(line string) (module, version string) {
	out := strings.SplitN(line, " ", 2)
	module = out[0]
	if len(out) == 2 {
		version = out[1]
	}

	return
}
