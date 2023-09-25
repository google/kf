// Copyright 2019 Google LLC
// Copyright 2018 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/licenseclassifier"
)

var LicenseNames = []string{
	"LICENCE",
	"LICENSE",
	"License",
	"LICENSE.code",
	"LICENSE.md",
	"license.md",
	"LICENSE.txt",
	"LICENSE.MIT",
	"LICENSE-APACHE-2.0.txt",
	"COPYING",
	"copyright",
}

const MatchThreshold = 0.9

type LicenseFile struct {
	Mod         Module
	LicensePath string
}

func (lf *LicenseFile) Body() (string, error) {
	body, err := ioutil.ReadFile(lf.LicensePath)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (lt *LicenseFile) Classify(classifier *licenseclassifier.License) (string, error) {
	body, err := lt.Body()
	if err != nil {
		return "", err
	}
	m := classifier.NearestMatch(body)
	if m == nil {
		return "", fmt.Errorf("unable to classify license: %v", lt.Mod)
	}
	return m.Name, nil
}

func (lt *LicenseFile) Check(classifier *licenseclassifier.License) error {
	body, err := lt.Body()
	if err != nil {
		return err
	}
	ms := classifier.MultipleMatch(body, false)
	for _, m := range ms {
		return fmt.Errorf("Found matching forbidden license in %v: %v", lt.Mod, m.Name)
	}
	return nil
}

func (lt *LicenseFile) CSVRow(site string, classifier *licenseclassifier.License) (string, error) {
	classification, err := lt.Classify(classifier)
	if err != nil {
		return "", err
	}

	return strings.Join([]string{
		lt.Mod.Module,
		"Static",
		"", // TODO(mattmoor): Modifications?
		"https://" + site + "/blob/master/vendor/" + lt.Mod.Module + "/" + filepath.Base(lt.LicensePath),
		classification,
	}, ","), nil
}

func findLicense(mod Module) (*LicenseFile, error) {
	dir := mod.Source()
	for {
		// When we reach the root of our workspace, stop searching.
		if dir == WorkingDir {
			return nil, fmt.Errorf("unable to find license for %q", mod.Module)
		}

		for _, name := range LicenseNames {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err != nil {
				continue
			}

			return &LicenseFile{
				Mod:         mod,
				LicensePath: p,
			}, nil
		}
	}
}

type LicenseCollection []*LicenseFile

func (lc LicenseCollection) CSV(site string, classifier *licenseclassifier.License) (string, error) {
	sections := make([]string, 0, len(lc))
	for _, entry := range lc {
		row, err := entry.CSVRow(site, classifier)
		if err != nil {
			return "", err
		}
		sections = append(sections, row)
	}
	return strings.Join(sections, "\n"), nil
}

func (lc LicenseCollection) Check(classifier *licenseclassifier.License) error {
	errors := []string{}
	for _, entry := range lc {
		if err := entry.Check(classifier); err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf("Errors validating licenses:\n%v", strings.Join(errors, "\n"))
}

func (lc LicenseCollection) GroupedEntries() (string, error) {
	w := &bytes.Buffer{}

	grouped := make(map[string]LicenseCollection)

	for _, lic := range lc {
		body, err := lic.Body()
		if err != nil {
			return "", err
		}

		grouped[body] = append(grouped[body], lic)
	}

	// sort by the biggest chunk of text so diffs are nicer for git
	var texts []string
	for text := range grouped {
		texts = append(texts, text)
	}
	sort.Stable((sort.StringSlice(texts)))

	// Many licenses are exact duplicates of one another. Grouping by these
	// cuts down the size by ~2/3.
	for _, text := range texts {
		licenses := grouped[text]
		fmt.Fprintln(w)
		fmt.Fprintln(w, "===========================================================")

		for _, license := range licenses {
			modText := ""
			switch {
			case license.Mod.Module == "":
				modText = license.Mod.ModuleVersion

			case license.Mod.Replace != "":
				modText = fmt.Sprintf("%s (imported as %s)", license.Mod.Replace, license.Mod.Module)

			default:
				modText = license.Mod.Module
			}

			fmt.Fprintln(w, "Module:", modText)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w)

		fmt.Fprintln(w, text)
	}

	return w.String(), nil
}

func CollectLicenses(packages []Module) (LicenseCollection, error) {
	// for each of the import paths, search for a license file.
	var licenseFiles []*LicenseFile
	for _, pkg := range packages {
		log.Printf("  finding license for: %s\n", pkg.Description())

		lf, err := findLicense(pkg)
		if err != nil {
			return nil, err
		}

		licenseFiles = append(licenseFiles, lf)
	}

	sort.SliceStable(licenseFiles, func(i, j int) bool {
		return licenseFiles[i].Mod.IsLessThan(licenseFiles[j].Mod)
	})

	return LicenseCollection(licenseFiles), nil
}

func ScanForMetadataLicenses(path string) (LicenseCollection, error) {
	log.Printf("  scanning path %s for METADATA files\n", WorkingDir)

	var collection LicenseCollection
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		_, err = os.Stat(filepath.Join(path, "METADATA"))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}

		module := Module{
			ModuleVersion: path,
		}

		l, err := findLicense(module)
		if err != nil {
			return err
		}

		if l != nil {
			collection = append(collection, l)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return collection, nil
}
