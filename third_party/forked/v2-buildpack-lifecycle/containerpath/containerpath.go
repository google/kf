// +build !windows !windows2012R2

package containerpath

import "path/filepath"

type cpath struct {
	root string
}

func New(_ string) *cpath {
	return &cpath{
		root: "",
	}
}

func (c *cpath) For(path ...string) string {
	path = append([]string{c.root}, path...)
	return filepath.Clean(filepath.Join(path...))
}
