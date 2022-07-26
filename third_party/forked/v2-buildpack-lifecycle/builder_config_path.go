// +build !windows2012R2

package buildpackapplifecycle

import "path/filepath"

func (s LifecycleBuilderConfig) getPath(path string) string {
	return filepath.Clean(path)
}
