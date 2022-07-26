// +build !windows2012R2

package buildpackrunner

import "os/exec"

func (runner *Runner) findTar() (string, error) {
	tarPath, err := exec.LookPath("tar")
	if err != nil {
		return "", err
	}
	return tarPath, nil
}
