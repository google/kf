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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/google/kf/pkg/kf/describe"
	"github.com/segmentio/textio"
)

func main() {
	uid := flag.Int("uid", 1000, "the user id of the buildpack user")
	gid := flag.Int("gid", 1000, "the group id of the buildpack user")
	workspace := flag.String("app", "/workspace", "the workspace for the app")
	image := flag.String("image", "", "the resulting image")
	runImage := flag.String("run-image", "", "the image to use as a stack")
	builderImage := flag.String("builder-image", "", "the image to use as the builder")
	useCredHelpers := flag.String("use-cred-helpers", "", "whether or not to use cred helpers")
	cacheVolume := flag.String("cache", "", "the name of the cache volume")
	buildpackOverrides := flag.String("buildpack", "", "custom buildpacks")

	flag.Parse()

	w := os.Stdout

	describe.SectionWriter(w, "Changing permissions", func(w io.Writer) {
		for _, dir := range []string{"/builder/home", "/layers", "/cache", *workspace} {
			fmt.Fprintf(w, "chown -R %d:%d %s\n", *uid, *gid, dir)

			if err := chown(dir, *uid, *gid); err != nil {
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

	describe.SectionWriter(w, "Docker config", func(w io.Writer) {
		cfg, err := readDockerConfig()
		if err != nil {
			fmt.Fprintf(w, "ERROR: %v\n", err)
			return
		}

		describe.SectionWriter(w, "Auth", func(w io.Writer) {
			fmt.Fprintln(w, "Registry\tUsername\tEmail\tPassword")
			for registry, v := range cfg.Auth {
				pass := "<blank>"
				if v.Password != "" {
					pass = "*****"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					registry,
					v.Username,
					v.Email,
					pass,
				)
			}
		})

		describe.SectionWriter(w, "Credential helpers", func(w io.Writer) {
			fmt.Fprintln(w, "Registry\tHelper")
			for registry, helper := range cfg.CredHelpers {
				fmt.Fprintf(w, "%s\t%s\n", registry, helper)
			}
		})
	})
	fmt.Fprintln(w)
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

type configFile struct {
	Auth        map[string]entry  `json:"auths"`
	CredHelpers map[string]string `json:"credHelpers"`
}

type entry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
	Email    string `json:"email"`
}

func readDockerConfig() (*configFile, error) {
	dockerDir := filepath.Join(os.Getenv("HOME"), ".docker")
	configPath := filepath.Join(dockerDir, "config.json")

	rawCfg, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	dockerConfig := configFile{}
	if err := json.Unmarshal(rawCfg, &dockerConfig); err != nil {
		return nil, err
	}

	return &dockerConfig, nil
}
