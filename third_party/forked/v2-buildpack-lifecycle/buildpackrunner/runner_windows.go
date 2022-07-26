package buildpackrunner

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func hasFinalize(buildpackPath string) (bool, error) {
	return windowsExecutableExists(filepath.Join(buildpackPath, "bin", "finalize"))
}

func hasSupply(buildpackPath string) (bool, error) {
	return windowsExecutableExists(filepath.Join(buildpackPath, "bin", "supply"))
}

func windowsExecutableExists(file string) (bool, error) {
	extensions := []string{".bat", ".exe", ".cmd"}

	for _, exe := range extensions {
		exists, err := fileExists(file + exe)
		if err != nil {
			return false, err
		} else if exists {
			return true, nil
		}
	}

	return false, nil
}

func (runner *Runner) copyApp(buildDir, stageDir string) error {
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		return err
	}
	return copyDirectory(buildDir, stageDir)
}

func (runner *Runner) warnIfDetectNotExecutable(buildpackPath string) error {
	return nil
}

func copyDirectory(srcDir, destDir string) error {
	destExists, err := fileExists(destDir)
	if err != nil {
		return err
	} else if !destExists {
		return errors.New("destination dir must exist")
	}

	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		src := filepath.Join(srcDir, f.Name())
		dest := filepath.Join(destDir, f.Name())

		if f.IsDir() {
			if err := os.MkdirAll(dest, f.Mode()); err != nil {
				return err
			}
			if err := copyDirectory(src, dest); err != nil {
				return err
			}
		} else {
			srcHandle, err := os.Open(src)
			if err != nil {
				return err
			}

			destHandle, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				srcHandle.Close()
				return err
			}

			_, err = io.Copy(destHandle, srcHandle)
			srcHandle.Close()
			destHandle.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil

}
