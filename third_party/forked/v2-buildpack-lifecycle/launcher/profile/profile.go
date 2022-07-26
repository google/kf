// +build !windows

package profile

import (
	"io"
)

func ProfileEnv(appDir, tempDir, getenvPath string, stdout io.Writer, stderr io.Writer) ([]string, error) {
	panic("not implemented for non-Windows OS")
}
