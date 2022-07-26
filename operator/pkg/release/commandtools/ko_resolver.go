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

package commandtools

import (
	"kf-operator/pkg/release/filetools"
	"kf-operator/pkg/release/stringtools"

	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// Substitution - Entity used to define literal text substitutions in string instances
type Substitution struct {
	Origin string
	Target string
}

type executeTerminalCommandFunc func(command string, arguments ...string) string

type globFunc func(pattern string) (matches []string, err error)

type countLinesFunc func(filePath string) int

type writeLinesFunc func(filePath, content string)

// KoResolver runs ko resolve command.
type KoResolver struct {
	executeTerminalCommand executeTerminalCommandFunc
	glob                   globFunc
	countLines             countLinesFunc
	writeFile              writeLinesFunc
}

// Resolve method runs ko resolve command with given arguments
func (r *KoResolver) Resolve(configPath, tag, resultFilePath, version string, substitutions ...Substitution) string {
	log.Printf("Processing path: %q", configPath)
	// Count lines in existing yaml files
	yamlFiles, err := r.glob(filepath.Join(configPath, "*.yaml"))
	if err != nil {
		panic(fmt.Sprintf("Yaml files could not be fetched in %q: %v", configPath, err))
	}

	log.Printf("Files: %v", yamlFiles)
	maxExpectedLineCount := 0
	for _, file := range yamlFiles {
		maxExpectedLineCount += r.countLines(file) + 2 // number of separators needed
	}
	if maxExpectedLineCount == 0 {
		panic(fmt.Sprintf("Was given a configPath %q with no yaml files", configPath))
	}
	minExpectedLineCount := maxExpectedLineCount - maxExpectedLineCount/10

	// Invoke ko resolve command
	args := []string{"resolve", "--base-import-paths", "--tags", version + ",latest", "-l",
		"operator.knative.dev/release!=test", "-f", configPath + "/"}
	resultOutput := r.executeTerminalCommand("ko", args...)

	resultOutput = strings.ReplaceAll(resultOutput,
		"operator.knative.dev/release: devel",
		fmt.Sprintf("operator.knative.dev/release: \"%s\"", tag))

	producedLines := strings.Split(resultOutput, "\n")
	filteredLines := make([]string, 0, len(producedLines))

	for _, line := range producedLines {
		if line == "" {
			continue
		}
		if line == "--- null" {
			continue
		}
		// Make all dynamic substitutions
		for _, substitution := range substitutions {
			line = strings.ReplaceAll(line, substitution.Origin, substitution.Target)
		}
		filteredLines = append(filteredLines, line)
	}

	// Check that produced file has lines in expected range
	if len(filteredLines) == 0 {
		panic("File produced by ko is empty!")
	}
	if len(filteredLines) < minExpectedLineCount {
		panic(fmt.Sprintf("File produced by ko has %d which is less than expected min of %d",
			len(filteredLines), minExpectedLineCount))
	}
	if len(filteredLines) > maxExpectedLineCount {
		panic(fmt.Sprintf("File produced by ko has %d which is greater than expected max of %d",
			len(filteredLines), maxExpectedLineCount))
	}

	// For every image key value line found in the produced yaml, check that it is a valid gcr image
	imageRe := regexp.MustCompilePOSIX(`^\\s*[^\\s]*[Ii]mage: `)
	// If it's a gke.gcr.io url, then we can't verify it exist
	// but it matches something like gke.gcr.io/some/path:valid_gke_release_tag
	// Then we assume it's a previously released image not produced by ko that is still valid
	// and thus skip checking if it exists
	gkeGrcRe := regexp.MustCompilePOSIX(`gke.gcr.io/[^:]+:[0-9]+\.[0-9]+\.[0-9]+-gke\.[0-9]+`)
	for _, line := range filteredLines {
		if imageRe.MatchString(line) && !gkeGrcRe.MatchString(line) {
			// All image keys should have values that are valid GCR urls to images we can find
			// gcloud will return an error return code if it is not able to find the image
			// and this will cause the program to exit
			r.executeTerminalCommand("gcloud", "container", "images", "describe", line)
		}
	}

	r.writeFile(resultFilePath, strings.Join(filteredLines, "\n")+"\n")

	return resultOutput
}

// NewKoResolver creates a KoResolver.
func NewKoResolver() *KoResolver {
	return &KoResolver{
		executeTerminalCommand: ExecuteTerminalCommand,
		glob:                   filepath.Glob,
		countLines:             stringtools.CountLines,
		writeFile:              filetools.WriteFile,
	}
}
