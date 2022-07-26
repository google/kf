# FORKED

This repo was forked at SHA e6b576c815a1759edba5747b2da8c460c6e679f3.
This was required because the repo doesn't vendor and we hit an issue where a
downstream library started failing. See b/200837342 for more information.

This was done by doing the following:

```
git clone https://github.com/cloudfoundry/buildpackapplifecycle/
rm -rf .git
go mod init

# Remove tests and fixtures. This is necessary as there are files that will
# trigger certain checks that would potentially flag us for having keys in our
# repo.
find . | grep fixtures | xargs rm -rf
find . | grep _test.go | xargs rm -rf
go mod tidy
mkdir installer
cat < 'EOF' > installer/main.go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// Copy the launcher and builder to /workspace.
	copyFile("/workspace/launcher", "/launcher")
	copyFile("/workspace/builder", "/builder")
}

func copyFile(dst, src string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %v", err)
	}
	defer r.Close()

	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	w, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %v", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("failed to copy to destination: %v", err)
	}

	return nil
}
EOF
```

# buildpackapplifecycle

**Note**: This repository should be imported as `code.cloudfoundry.org/buildpackapplifecycle`.

The buildpack lifecycle implements the traditional Cloud Foundry deployment
strategy.

The **Builder** downloads buildpacks and app bits, and produces a droplet.

The **Launcher** runs the start command using a standard rootfs and
environment.

Read about the app lifecycle spec here: https://github.com/cloudfoundry/diego-design-notes#app-lifecycles

### Running tests

On linux or windows, please use `ginkgo -r` from this directory.

On a mac, you should use docker to run the tests on a linux machine

```
$ docker build -f Dockerfile.linux.test -t buildpackapplifecycle .
$ docker run -it buildpackapplifecycle
```
