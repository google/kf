// +build windows,windows2012R2

package buildpackapplifecycle

import (
	"path/filepath"
)

func (s LifecycleBuilderConfig) getPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(s.workingDir, path)
}
