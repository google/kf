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

package filetools

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// MakeTempModuleOutsideGoPath makes a temporary module outside the GOPATH.
// Without this ko and go are oddly in GOPATH mode.
func MakeTempModuleOutsideGoPath(module string) {
	tmp, err := ioutil.TempDir("/tmp/", "NOTGOPATH-")
	if err != nil {
		log.Fatalf("Could not create temp GOPATH [%s]", err)
	}
	copy(module, tmp)
	if err := os.Chdir(tmp); err != nil {
		log.Fatalf("Could not chdir to the temp module [%s].", err)
	}
}

// CreateDir creates a directory.
func CreateDir(dir string) {
	os.Mkdir(dir, 0755)
}

// ReadFile - Read the content of a file and return its string representation
func ReadFile(filePath string) string {
	// Read the whole file. We do not expect huge text files
	// to be parsed with this script. Mostly configuration and
	// yaml files
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Could not read file %q", filePath)
	}
	s := string(b)
	return s
}

// WriteFile - Write the string in the file requested
func WriteFile(filePath, content string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Could not write file %q", filePath)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		log.Fatalf("Could not write data to file %q", filePath)
	}
}

// CreateTarballFile - Create a Tarball file from the directory provided
func CreateTarballFile(tarFileName, directoryPath string) {
	err := os.RemoveAll(tarFileName)
	if err != nil {
		log.Fatalf("Failed to remove %s tar temporary file", tarFileName)
	}
	// Create new tarball file and write it
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		log.Fatalf("Failed to create %s tar temporary file", tarFileName)
	}
	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	// walk through every file in the directory
	filepath.Walk(directoryPath, func(file string, fileInfo os.FileInfo, err error) error {
		// Return on any error
		if err != nil {
			return err
		}

		// Create a new header
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		// Update the name to correctly reflect the desired destination when untaring
		header.Name = filepath.ToSlash(file)
		header.Linkname, _ = os.Readlink(file)

		// write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it is not a directory, write its content to the tarball file
		if !fileInfo.IsDir() {
			// Do not process on non-regular files
			if !fileInfo.Mode().IsRegular() {
				return nil
			}
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
			data.Close()
		}

		return nil
	})
}

// CreateTempDirectory - Create a temporary directory following the provided pattern
func CreateTempDirectory(pattern string) string {
	tempDirectory, err := ioutil.TempDir("", pattern)
	if err != nil {
		log.Fatalf("Failed to create temporary directory for %q pattern", pattern)
	}
	return tempDirectory
}

// CreateTempFile - Create a temporary file following the provided pattern
func CreateTempFile(pattern string) string {
	tempFile, err := ioutil.TempFile("", pattern)
	if err != nil {
		log.Fatalf("Failed to create temporary file for %q pattern", pattern)
	}
	return tempFile.Name()
}

// CreateFileLinks - Remove target, if exists, and create a soft link to origin
func CreateFileLinks(origin, target string) {
	// Remove potential existing path
	err := os.RemoveAll(target)
	if err != nil {
		log.Fatalf("Failed to remove %q file", target)
	}
	// Generate new symbolic link
	err = os.Symlink(origin, target)
	if err != nil {
		log.Fatalf("Failed to create symbolic link from %q [%s]", origin, err)
	}
}

// RemoveAll - Remove all files and directories in the path provided
func RemoveAll(path string) {
	error := os.RemoveAll(path)
	if error != nil {
		log.Fatalf("Failed to remove %s", path)
	}
}

// copy - Copy origin to target. Directories will be copied recursively
func copy(origin, target string) {
	cmd := exec.Command("cp", "-vR", origin, target)
	log.Printf("Copy: %s Target: %s", origin, target)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("File copy failed from %s to %s", origin, target)
	}
}

// Copy - Copy files extracted from the input template to the target directory
func Copy(templateInput, target string) {
	files, err := filepath.Glob(templateInput)
	if err != nil {
		log.Fatalf("Files could not be fetched for copy from %s", templateInput)
	}
	for _, file := range files {
		copy(file, filepath.Join(target, filepath.Base(file)))
	}
}

// Move - Move files extracted from the input template to the target directory
func Move(templateInput, target string) {
	files, err := filepath.Glob(templateInput)
	if err != nil {
		log.Fatalf("Files could not be fetched for move from %s", templateInput)
	}
	for _, file := range files {
		err := os.Rename(file, filepath.Join(target, filepath.Base(file)))
		if err != nil {
			log.Fatalf("File %s could not be moved", file)
		}
		log.Printf("Move: %s Target: %s", file, filepath.Base(file))
	}
}
