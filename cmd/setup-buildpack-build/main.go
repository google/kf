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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/pkg/dockerutil"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/segmentio/textio"
)

func main() {
	workspace := flag.String("app", "/workspace", "the workspace for the app")
	image := flag.String("image", "", "the resulting image")
	runImage := flag.String("run-image", "", "the image to use as a stack")
	builderImage := flag.String("builder-image", "", "the image to use as the builder")
	useCredHelpers := flag.String("use-cred-helpers", "", "whether or not to use cred helpers")
	cacheVolume := flag.String("cache", "", "the name of the cache volume")
	buildpackOverrides := flag.String("buildpack", "", "custom buildpacks")

	flag.Parse()

	w := os.Stdout

	uid, gid, err := getBuildUser(*builderImage)
	if err != nil {
		fmt.Printf("couldn't fetch UID/GID from image: %v\n", err)
	}

	if uid == 0 {
		uid = 1000
		fmt.Printf("UID was 0, defaulting to %d\n", uid)
	}

	if gid == 0 {
		gid = 1000
		fmt.Printf("GID was 0, defaulting to %d\n", gid)
	}

	describe.SectionWriter(w, "Changing permissions", func(w io.Writer) {
		fmt.Fprintf(w, "chmod %d %s\n", 0744, *workspace)
		if err := os.Chmod(*workspace, 0744); err != nil {
			log.Fatal(err)
		}

		for _, dir := range []string{"/builder/home", "/layers", "/cache", *workspace} {
			fmt.Fprintf(w, "chown -R %d:%d %s\n", uid, gid, dir)

			if err := chown(dir, uid, gid); err != nil {
				log.Fatal(err)
			}
		}
	})
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Uploaded files:")
	ls(w, *workspace)
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	describe.SectionWriter(w, "Build info", func(w io.Writer) {
		fmt.Fprintf(w, "Destination image:\t%s\n", *image)
		fmt.Fprintf(w, "Stack:\t%s\n", *runImage)
		fmt.Fprintf(w, "Builder image:\t%s\n", *builderImage)
		fmt.Fprintf(w, "Use cred helpers:\t%s\n", *useCredHelpers)
		fmt.Fprintf(w, "Cache volume:\t%s\n", *cacheVolume)
		fmt.Fprintf(w, "Buildpack overrides:\t%q\n", *buildpackOverrides)
	})
	fmt.Fprintln(w)

	dockerutil.DescribeDefaultConfig(w)
}

// getBuildUser returns the UID and GID as specified by the buildpack builder
// image environment variables subject to the platform build image spec:
// https://github.com/buildpack/spec/blob/master/platform.md#build-image
func getBuildUser(builderImage string) (uid, gid int, err error) {
	imageRef, err := name.ParseReference(builderImage, name.WeakValidation)
	if err != nil {
		return 0, 0, err
	}

	image, err := remote.Image(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return 0, 0, err
	}

	cfg, err := image.ConfigFile()
	if err != nil {
		return 0, 0, err
	}

	for _, val := range cfg.Config.Env {
		parts := strings.Split(val, "=")
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		if key == "CNB_USER_ID" {
			uid, err = strconv.Atoi(value)
			if err != nil {
				return
			}
		}

		if key == "CNB_GROUP_ID" {
			gid, err = strconv.Atoi(value)
			if err != nil {
				return
			}
		}
	}

	return uid, gid, nil
}

func chown(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return os.Chown(name, uid, gid)
	})
}

func ls(w io.Writer, path string) {
	tree := textio.NewTreeWriter(w)
	tree.WriteString(filepath.Base(path))
	defer tree.Close()

	files, _ := ioutil.ReadDir(path)

	for _, f := range files {
		if f.Mode().IsDir() {
			ls(tree, filepath.Join(path, f.Name()))
		}
	}

	for _, f := range files {
		if !f.Mode().IsDir() {
			io.WriteString(textio.NewTreeWriter(tree), f.Name())
		}
	}
}
