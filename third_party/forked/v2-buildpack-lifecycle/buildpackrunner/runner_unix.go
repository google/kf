// +build !windows

package buildpackrunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func hasFinalize(buildpackPath string) (bool, error) {
	return fileExists(filepath.Join(buildpackPath, "bin", "finalize"))
}

func hasSupply(buildpackPath string) (bool, error) {
	return fileExists(filepath.Join(buildpackPath, "bin", "supply"))
}

func (runner *Runner) copyApp(buildDir, stageDir string) error {
	return runner.run(exec.Command("cp", "-a", buildDir, stageDir), os.Stdout)
}

func (runner *Runner) warnIfDetectNotExecutable(buildpackPath string) error {
	fileInfo, err := os.Stat(filepath.Join(buildpackPath, "bin", "detect"))
	if err != nil {
		return err
	}

	if fileInfo.Mode()&0111 != 0111 {
		fmt.Println("WARNING: buildpack script '/bin/detect' is not executable")
	}

	return nil
}
