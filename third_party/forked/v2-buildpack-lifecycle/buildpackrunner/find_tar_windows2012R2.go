// +build windows,windows2012R2

package buildpackrunner

import "path/filepath"

func (runner *Runner) findTar() (string, error) {
	return filepath.Join(filepath.Dir(runner.config.Path()), "tar.exe"), nil
}
