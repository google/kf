// +build windows

package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ProfileEnv(appDir, tempDir, getenvPath string, stdout io.Writer, stderr io.Writer) ([]string, error) {
	fi, err := os.Stat(tempDir)
	if err != nil {
		return nil, fmt.Errorf("invalid temp dir: %s", err.Error())
	} else if !fi.IsDir() {
		return nil, errors.New("temp dir must be a directory")
	}

	envOutputFile := filepath.Join(tempDir, "launcher.env")
	defer os.Remove(envOutputFile)

	batchFileLines := []string{
		"@echo off",
		fmt.Sprintf("cd %s", appDir),
		`(for /r %i in (..\profile.d\*) do %i)`,
		`(for /r %i in (.profile.d\*) do %i)`,
		`(if exist .profile.bat ( .profile.bat ))`,
		fmt.Sprintf("%s -output %s", getenvPath, envOutputFile),
	}

	cmd := exec.Command("cmd", "/c", strings.Join(batchFileLines, " & "))
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return []string{}, fmt.Errorf("running profile scripts failed: %s", err.Error())
	}
	out, err := ioutil.ReadFile(envOutputFile)
	if err != nil {
		return []string{}, err
	}

	cleanedVars := []string{}
	if err := json.Unmarshal(out, &cleanedVars); err != nil {
		return []string{}, fmt.Errorf("cannot unmarshal environmental variables: %s", err.Error())
	}

	return cleanedVars, nil
}
