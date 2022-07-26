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

package main

// TODO(b/155773558): Share library with knative/cloudrun

import (
	"flag"
	"fmt"
	"log"
	"regexp"
)

type options struct {
	// development is for execution mode (development/test vs production)
	development bool
	// gcrPath is the path to GCR where release image will be stored
	gcrPath string
	// gcsPath is the path to GCS where YAML files will be stored
	gcsPath string
	// version is the target version that will be released
	version string
}

func (o *options) mustProcessFlags() error {
	if o.gcrPath == "" {
		return fmt.Errorf("gcrPath is empty")
	}
	if o.gcsPath == "" {
		return fmt.Errorf("gcsPath is empty")
	}

	gcrPathRe := regexp.MustCompilePOSIX(gcrPathRegExp)
	if !gcrPathRe.MatchString(o.gcrPath) {
		return fmt.Errorf("Failed to parse GCR_PATH_ROOT from GCR path=%s\n"+
			"Make sure GCR path is a proper gcr.io url",
			o.gcrPath)
	}

	const (
		versionRegExp  = "^[0-9]+.[0-9]+.[0-9]+-gke.[0-9]+$"
		commitIDRegExp = "^[a-f0-9]{7,}$"
	)
	versionRe := regexp.MustCompilePOSIX(versionRegExp)
	commitIDRe := regexp.MustCompilePOSIX(commitIDRegExp)
	// The release version can always be a valid semver.
	// If develoipment mode is requested, a  7+ digit short commit hash is also valid
	if !versionRe.MatchString(o.version) && (!commitIDRe.MatchString(o.version) || !o.development) {
		return fmt.Errorf("Release version %q must either be in the form of X.Y.Z-gke.A or if development argument is set, a 7+ digit short commit hash", o.version)
	}

	log.Printf("============= Using argument values =============")
	if o.development {
		log.Printf("Mode = Testing/Development")
	} else {
		log.Printf("Mode = Production")
	}
	log.Printf("GCR Path = %s", o.gcrPath)
	log.Printf("GCS Path = %s", o.gcsPath)
	log.Printf("Version = %s", o.version)
	return nil
}

func (o *options) addFlags() {
	flag.BoolVar(&o.development, "development", false, "[optional] Execute on development/test mode")
	flag.StringVar(&o.gcrPath, "gcrPath", "", "Target GCR path where image will be stored")
	flag.StringVar(&o.gcsPath, "gcsPath", "", "Target GCS path where YAML files will be stored")
	flag.StringVar(&o.version, "version", "", "Target version to release")
}
