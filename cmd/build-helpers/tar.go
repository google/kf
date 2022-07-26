// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path/filepath"

	"github.com/google/kf/v2/pkg/sourceimage"
	"github.com/spf13/cobra"
)

func NewTarCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tar DST SRC",
		Short: "Create a tar from a directory",
		Long:  "Create a tar from a directory in the same manner the Kf CLI would on a push.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dst := args[0]
			src := args[1]

			// Source paths have to be absolute.
			var err error
			src, err = filepath.Abs(src)
			if err != nil {
				return err
			}

			f, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer f.Close()

			// For now, we want to simply include every file.
			includeAllFiles := func(string) bool {
				return true
			}

			if err := sourceimage.PackageSourceTar(
				f,
				src,
				includeAllFiles,
			); err != nil {
				return err
			}

			return nil
		},
	}
}
