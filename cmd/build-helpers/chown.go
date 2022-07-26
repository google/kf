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
	"strconv"

	"github.com/spf13/cobra"
)

// NewChownCommand creates a command that will change the owner of a given path
// This requires the base image to have root permissions
func NewChownCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "chown PATH UID GID",
		Short: "Change the owner of PATH to UID and GID",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, uid, gid := args[0], args[1], args[2]

			uidValue, err := strconv.Atoi(uid)
			if err != nil {
				return err
			}

			gidValue, err := strconv.Atoi(gid)
			if err != nil {
				return err
			}

			return chown(path, uidValue, gidValue)
		},
	}
}

func chown(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return os.Chown(name, uid, gid)
	})
}
