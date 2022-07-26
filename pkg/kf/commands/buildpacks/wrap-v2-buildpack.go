// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildpacks

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

type logger func(string, ...interface{})

// NewWrapV2BuildpackCommand creates a WrapV2BuildpackCommand command.
func NewWrapV2BuildpackCommand() *cobra.Command {
	var (
		builderRepo  string
		launcherRepo string
		outputDir    string
		version      string
		stacks       []string
		publish      bool
	)

	buildpacksCmd := &cobra.Command{
		Use:   "wrap-v2-buildpack NAME V2_BUILDPACK_URL|PATH",
		Short: "Create a V3 buildpack that wraps a V2 buildpack.",
		Args:  cobra.ExactArgs(2),
		Example: `
		# Download buildpack from the given git URL. Uses the git CLI to
		# download the repository.
		kf wrap-v2-buildpack gcr.io/some-project/some-name https://github.com/some/buildpack

		# Creates the buildpack from the given path.
		kf wrap-v2-buildpack gcr.io/some-project/some-name path/to/buildpack

		# Creates the buildpack from the given archive file.
		kf wrap-v2-buildpack gcr.io/some-project/some-name path/to/buildpack.zip
		`,
		Long: `
		Creates a V3 buildpack that wraps a V2 buildpack.

		The resulting buildpack can then be used with other V3 buildpacks by
		creating a builder. See
		https://buildpacks.io/docs/operator-guide/create-a-builder/ for more
		information.

		A V3 buildpack is packaged as an OCI container. If the --publish flag
		is provided, then the container will be published to the corresponding
		container repository.

		This command uses other CLIs under the hood. This means the following
		CLIs need to be available on the path:
		* go
		* git
		* pack
		* cp
		* unzip

		We recommend using Cloud Shell to ensure these CLIs are available and
		the correct version.
		`,
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			buildpackName := args[0]
			buildpackRepo := args[1]
			logger := logging.FromContext(ctx)

			if strings.ContainsAny(buildpackName, ":") {
				return fmt.Errorf("buildpack name (%q) must not include a tag. Instead use --buildpack-version", buildpackName)
			}

			logger.Infof("checking for all the necessary CLIs (go, git, pack, cp and unzip)")
			if err := checkForCLIs(ctx, "go", "git", "pack", "cp", "unzip"); err != nil {
				return err
			}

			// Setup a temp dir if output dir was not provided.
			if outputDir == "" {
				tempDir, err := ioutil.TempDir("", "")
				if err != nil {
					return fmt.Errorf("failed to create temp dir: %v", err)
				}
				outputDir = tempDir

				defer func() {
					// We can't actually remove all of this because the
					// resulting $GOPATH/pkg/mod directory will have strict
					// permissions. So instead, we'll recommend the user does
					// it (they'll need sudo and we don't want to request that
					// just for clean up).
					logger.Infof("NOTE : The temp directory %s was not cleaned up.", tempDir)
				}()
			}
			if err := os.MkdirAll(filepath.Join(outputDir, "bin"), os.ModePerm); err != nil {
				return fmt.Errorf("failed to setup bin in temp dir: %v", err)
			}

			if err := createBuildpack(ctx, outputDir, buildpackName, version, stacks); err != nil {
				return err
			}
			if err := goGetRepos(ctx, outputDir, buildpackName, builderRepo, launcherRepo); err != nil {
				return err
			}

			if isDir(buildpackRepo) {
				// Assume it is a path which needs to be copied instead.
				logger.Infof("assuming path is a buildpack directory...")
				if err := copyBuildpack(ctx, outputDir, buildpackName, buildpackRepo); err != nil {
					return err
				}
			} else if isZippedFile(ctx, buildpackRepo) {
				// Zipped file, unzip it into the outputDir.
				logger.Infof("assuming path is a zip file...")
				if err := unzipBuildpack(ctx, outputDir, buildpackName, buildpackRepo); err != nil {
					return err
				}
			} else if _, err := url.Parse(buildpackRepo); err == nil {
				// Assume the repo is a git URL.
				logger.Infof("assuming path is a git URL...")
				if err := cloneBuildpackRepo(ctx, outputDir, buildpackName, buildpackRepo); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unable to figure out what to do with %s", buildpackRepo)
			}

			createZipFileFromSrc(ctx, outputDir)
			createWrapperScripts(outputDir, buildpackName)
			copyV2WrapperBuildpack(ctx, outputDir, buildpackName)

			if err := packageBuildpack(ctx, outputDir, buildpackName, version, publish); err != nil {
				return err
			}

			// Write out the resulting image name to stdout.
			fmt.Fprintln(cmd.OutOrStdout(), buildpackName)

			return nil
		},
	}

	buildpacksCmd.Flags().StringVar(&builderRepo, "builder-repo", "github.com/poy/buildpackapplifecycle/builder", "Builder repo to use.")
	buildpacksCmd.Flags().StringVar(&launcherRepo, "launcher-repo", "github.com/poy/buildpackapplifecycle/launcher", "Launcher repo to use.")
	buildpacksCmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for the buildpack data (before it's packed). If left empty, a temp dir will be used.")
	buildpacksCmd.Flags().StringVar(&version, "buildpack-version", "0.0.1", "Version of the resulting buildpack. This will be used as the image tag.")
	buildpacksCmd.Flags().StringArrayVar(&stacks, "buildpack-stacks", []string{"google"}, "Stack(s) this buildpack will be compatible with.")
	buildpacksCmd.Flags().BoolVar(&publish, "publish", false, "Publish the buildpack image.")

	return buildpacksCmd
}

// copyV2WrapperBuildpack copies the wrapper V2 buildpack artifacts from {outputDir}/buildpack/bin to the stage location
// ({outputDir}/pack/{buildpackName}/bin/v2-buildpack/{hash}) where it will be packaged into
// a V3 buildpack in the next step (packageBuildpack).
func copyV2WrapperBuildpack(ctx context.Context, outputDir, buildpackName string) error {
	srcDir := filepath.Join(outputDir, "buildpack", "bin")
	dstPath := hashedPath(outputDir, buildpackName)
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir: %v", err)
	}
	if out, err := command(ctx, outputDir, "cp", "-r", srcDir, dstPath); err != nil {
		return fmt.Errorf("failed to copy: %v\n%s", err, out)
	}
	return nil
}

// createZipFileFromSrc creates a zip file of the source code of the V2 buildpack from {outputDir}/buildpack/src.
// The generated zip file is created at {outputDir}/buildpack/bin/buildpack.zip.
// This is a workaround added to mitigate a TAR issue when wrapping V2 buildpacks.
// If the name of the buildpack surpass 100 characters, we bump into a TAR issue:
// https://github.com/golang/go/issues/24821
// https://github.com/docker/for-linux/issues/484
// See context at b/202031925 and http://b/205586642
func createZipFileFromSrc(ctx context.Context, outputDir string) error {
	logger := logging.FromContext(ctx)
	// buildpackSrcDir is where the source code of the V2 buildpack is downloaded/copied to,
	// it is a stage directory where the buildpack source code will be packaged into a zip file.
	buildpackSrcDir := filepath.Join(outputDir, "buildpack", "src")
	zipFileName := "buildpack.zip"
	zipFilePath := filepath.Join(buildpackSrcDir, zipFileName)

	logger.Infof("archiving buidpack %s to %s...", buildpackSrcDir, zipFilePath)
	if out, err := command(ctx, buildpackSrcDir, "zip", "-r", zipFileName, "."); err != nil {
		return fmt.Errorf("failed to zip: %v\n%s", err, out)
	}
	// Copy the generated zip file to the {outputDir}/buildpack/bin directory. The zip file
	// will be packaged up as a V2 wrapper buildpack along with the detect/compile/release/finalize/supply
	// scripts generated in the next step (createWrapperScripts).
	dstPath := filepath.Join(outputDir, "buildpack", "bin")
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir: %v", err)
	}
	if out, err := command(ctx, buildpackSrcDir, "mv", zipFileName, dstPath); err != nil {
		return fmt.Errorf("failed to move %s file to %s: %v\n%s", zipFileName, dstPath, err, out)
	}
	return nil
}

// createWrapperScripts creates wrapper scripts on top of the V2 buildpack's scripts.
// See https://docs.cloudfoundry.org/buildpacks/understand-buildpacks.html
func createWrapperScripts(outputDir, name string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(name)))
	scriptNames := []string{"detect", "compile", "finalize", "release", "supply"}
	scriptTemplate := `#!/usr/bin/env bash
set -euo pipefail

if [ ! -d "/tmp/v2-buildpack-%s" ]; then
	unzip -q ${0%%/*}/buildpack.zip -d /tmp/v2-buildpack-%s
fi

exec /tmp/v2-buildpack-%s/bin/%s "$@"
`
	for _, s := range scriptNames {
		scriptContent := []byte(fmt.Sprintf(scriptTemplate, hash, hash, hash, s))
		if err := createScript(outputDir, s, scriptContent); err != nil {
			return err
		}
	}

	return nil
}

func createScript(outputDir string, scriptName string, content []byte) error {
	// Only create a wrapper script if the source script exists. Not all the buildpack contains all the 5
	// scripts (detect/compile/finalize/release/supply), for example, java buildpack only contains 4 scripts while
	// the go buildpack contains all 5 scripts. Wrapper script shouldn't be generated if the source script doesn't exist.
	srcPath := filepath.Join(outputDir, "buildpack", "src", "bin", scriptName)
	if _, err := os.Stat(srcPath); err == nil {
		dstPath := filepath.Join(outputDir, "buildpack", "bin", scriptName)
		if err := ioutil.WriteFile(dstPath, content, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create wrapper script (%s): %v", dstPath, err)
		}
	}
	return nil
}

func isDir(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func isZippedFile(ctx context.Context, filePath string) bool {
	logger := logging.FromContext(ctx)
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Why 512 bytes? See http://golang.org/pkg/net/http/#DetectContentType
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		logger.Infof("failed to read file: %v", err)
		return false
	}

	return http.DetectContentType(buff) == "application/zip"
}

func buildpackPath(outputDir, name string) string {
	return filepath.Join(outputDir, "pack", name)
}

func packageBuildpack(ctx context.Context, outputDir, name, version string, publish bool) error {
	logger := logging.FromContext(ctx)
	packagePath := filepath.Join(outputDir, "pack", "package.toml")

	// Write package.toml for the buildpack.
	if err := ioutil.WriteFile(packagePath, []byte(fmt.Sprintf(`
[buildpack]
uri = %q
`, name)), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write package.toml: %v", err)
	}

	commandArgs := []string{"buildpack", "package",
		name + ":" + version,
		"--config=" + packagePath,
	}

	if publish {
		commandArgs = append(commandArgs, "--publish")
	}

	logger.Infof("packaging and publishing %s...", name)
	if out, err := command(ctx, outputDir, "pack", commandArgs...); err != nil {
		return fmt.Errorf("failed to package buildpack: %v\n%s", err, out)
	}
	logger.Infof("done packaging and publishing %s.", name)

	return nil
}

// createBuildpack more or less follows the tutorial found in
// https://buildpacks.io/docs/buildpack-author-guide/create-buildpack/
func createBuildpack(ctx context.Context, outputDir, name, version string, stacks []string) error {
	logger := logging.FromContext(ctx)
	buildpackPath := buildpackPath(outputDir, name)

	// Check to see if the file already exists, if so, skip it.
	if _, err := os.Stat(buildpackPath); err == nil {
		// File exists. Move on.
		logger.Infof("%s already exists, skipping buildpack init step.", buildpackPath)
	} else if os.IsNotExist(err) {
		// File does NOT exist.
		logger.Infof("initilizing buildpack at %s...", buildpackPath)
		if out, err := command(ctx, outputDir, "pack", "buildpack", "new", name,
			"--api=0.5",
			"--path="+buildpackPath,
			"--version="+version,
			"--stacks="+strings.Join(stacks, ","),
		); err != nil {
			return fmt.Errorf("failed to initialize buildpack dir: %v\n%s", err, out)
		}
	} else {
		return fmt.Errorf("failed to check for existing file (%s): %v", buildpackPath, err)
	}

	// Update the detect script that was created when we ran `pack buildpack
	// new` to our desired behavior.
	if err := updateDetectScript(buildpackPath); err != nil {
		return err
	}
	// Update the build script that was created when we ran `pack buildpack
	// new` to our desired behavior.
	if err := updateBuildScript(buildpackPath); err != nil {
		return err
	}

	return nil
}

func updateDetectScript(buildpackPath string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte("v2-buildpack")))
	detectPath := filepath.Join(buildpackPath, "bin", "detect")
	if err := ioutil.WriteFile(detectPath, []byte(fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

script_path="${0%%/*}"

# TODO: Enviroment variables
${script_path}/v2-buildpack/%s/bin/detect ./

exit 0
`, hash)), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create new detect script (%s): %v", detectPath, err)
	}

	return nil
}

func updateBuildScript(buildpackPath string) error {
	buildPath := filepath.Join(buildpackPath, "bin", "build")
	if err := ioutil.WriteFile(buildPath, []byte(`#!/usr/bin/env bash
set -euo pipefail

layersdir="$1"
env_dir="$2/env"
plan_path="$3"
script_path="${0%/*}"
shim_layer="${layersdir}/shim"
shim_cache_layer="${layersdir}/shim_cache"

mkdir -p ${shim_layer}/{bin,droplet} ${shim_cache_layer}

echo -e 'cache = true\nlaunch = true' > "${layersdir}/shim_cache.toml"
echo -e 'launch = true' > "$layersdir/shim.toml"

# A /tmp directory is necessary because some buildpacks use /tmp
# which causes cross-device links to be made because Tekton mounts the
# /workspace directory.
cp -r . /tmp/app/
cd /tmp/app

# Many of the V2 buildpacks expect this environment variable to be set...
export CF_STACK=cflinuxfs3

${script_path}/v2-lifecycle/builder \
  -buildDir=/tmp/app \
  -buildpacksDir=${script_path}/v2-buildpack \
  -outputDroplet=${shim_layer}/droplet/droplet.tar.gz \
  -outputMetadata=${shim_layer}/bin/result.json \
  -buildpackOrder=v2-buildpack \
  -buildArtifactsCacheDir=${shim_cache_layer}/cache \
  -outputBuildArtifactsCache=${shim_cache_layer}/output-cache \
  -skipDetect=true

cp ${script_path}/v2-lifecycle/launcher ${shim_layer}/bin

pushd ${shim_layer}/droplet
  tar -xzf droplet.tar.gz && rm -f droplet.tar.gz
popd

cat << EOF > ${shim_layer}/bin/entrypoint.sh
#!/usr/bin/env bash
set -e

cd "\${0%/*}"/../droplet

# TODO: Set user and home dir (app should be at /home/vcap/app), if it's possible...

if [[ "\$@" == "" ]]; then
  # ${shim_layer}/bin/launcher "/home/vcap/app" "" ""
  exec ${shim_layer}/bin/launcher "${shim_layer}/droplet/app" "" ""
else
  # ${shim_layer}/bin/launcher "/home/vcap/app" "\$@" ""
  exec ${shim_layer}/bin/launcher "${shim_layer}/droplet/app" "\$@" ""
fi
EOF

chmod a+x ${shim_layer}/bin/entrypoint.sh

cat > "$layersdir/launch.toml" <<EOF
[[processes]]
type = "web"
command = "${shim_layer}/bin/entrypoint.sh"
EOF

`), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create new build script (%s): %v", buildPath, err)
	}

	return nil
}

// goGetRepos does a `go get` on each of the given repo. It then moves the
// contents to the output directory where the compiled artifacts will be
// included in the OCI container.
func goGetRepos(ctx context.Context, outputDir, buildpackName string, repos ...string) error {
	logger := logging.FromContext(ctx)
	for _, repo := range repos {
		logger.Infof("Downloading repo (%s) to %s", repo, outputDir)
		if out, err := goCommand(ctx, outputDir, "get", repo); err != nil {
			return fmt.Errorf("failed to download repo: %v\n%s", err, out)
		}
	}

	return moveBinToBuildpack(ctx, outputDir, buildpackName)
}

func cp(ctx context.Context, srcPath, dstPath string) error {
	logger := logging.FromContext(ctx)
	// Check to see if the file already exists, if so, skip it.
	if _, err := os.Stat(dstPath); err == nil {
		// File exists. Move on.
		logger.Infof("%s already exists, moving on.", dstPath)
		return nil
	} else if os.IsNotExist(err) {
		// File does NOT exist.
		logger.Infof("copying %s to %s", srcPath, dstPath)
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("failed to check for existing file (%s): %v", dstPath, err)
	}
}

// moveBinToBuildpack moves the artifacts into a directory that will be
// included in the OCI container AND won't have any risk of name collisions
// with other scripts/binaries (e.g., detect).
func moveBinToBuildpack(ctx context.Context, outputDir, buildpackName string) error {
	logger := logging.FromContext(ctx)

	// Binaries can be found in both $GOPATH/bin or $GOPATH/bin/linux_*. So
	// look at both.
	binDir := filepath.Join(outputDir, "bin")
	buildpackV2Bin := filepath.Join(buildpackPath(outputDir, buildpackName), "bin", "v2-lifecycle")

	if err := os.MkdirAll(buildpackV2Bin, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create dir (%s): %v", buildpackV2Bin, err)
	}

	files, err := ioutil.ReadDir(binDir)
	if err != nil {
		return fmt.Errorf("failed to list directory: %v", err)
	}

	for _, f := range files {
		if !f.IsDir() {
			dstPath := filepath.Join(buildpackV2Bin, f.Name())
			if err := cp(ctx, filepath.Join(binDir, f.Name()), dstPath); err != nil {
				return err
			}
			continue
		} else if !strings.HasPrefix(f.Name(), "linux_") {
			// This shouldn't really happen because only linux_* directories
			// should exist here, but just in case we'll skip it.
			continue
		}

		logger.Infof("found dir %s while normalizing bin locations", f.Name())
		linuxDir := filepath.Join(binDir, f.Name())

		innerFiles, err := ioutil.ReadDir(linuxDir)
		if err != nil {
			return fmt.Errorf("failed to list directory: %v", err)
		}

		for _, ff := range innerFiles {
			dstPath := filepath.Join(buildpackV2Bin, ff.Name())
			if err := cp(ctx, filepath.Join(linuxDir, ff.Name()), dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func unzipBuildpack(
	ctx context.Context,
	outputDir string,
	buildpackName string,
	src string,
) error {
	logger := logging.FromContext(ctx)
	dstPath := filepath.Join(outputDir, "buildpack", "src")

	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir: %v", err)
	}

	// XXX: For now, we're doing the lazy/safe thing and simply shelling out
	// to unzip.
	logger.Infof("unarchiving buidpack %s to %s...", src, dstPath)
	if out, err := command(ctx, outputDir, "unzip", src, "-d", dstPath); err != nil {
		return fmt.Errorf("failed to unzip: %v\n%s", err, out)
	}

	return nil
}

func copyBuildpack(
	ctx context.Context,
	outputDir string,
	buildpackName string,
	src string,
) error {
	logger := logging.FromContext(ctx)
	dstPath := filepath.Join(outputDir, "buildpack", "src")

	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir: %v", err)
	}

	// XXX: For now, we're doing the lazy/safe thing and simply shelling out
	// to cp.
	logger.Infof("copying buidpack dir %s to %s...", src, dstPath)
	if out, err := command(ctx, src, "cp", "-r", "./", dstPath); err != nil {
		return fmt.Errorf("failed to copy: %v\n%s", err, out)
	}

	return nil
}

func cloneBuildpackRepo(
	ctx context.Context,
	outputDir string,
	buildpackName string,
	repo string,
) error {
	logger := logging.FromContext(ctx)
	dstPath := filepath.Join(outputDir, "buildpack", "src")
	if _, err := os.Stat(dstPath); err == nil {
		// File exists. Move on.
		logger.Infof("%s already exists, moving on.", dstPath)
		return nil
	} else if os.IsNotExist(err) {
		// File does NOT exist.
		logger.Infof("cloning %s to %s...", repo, dstPath)
		if out, err := command(ctx, outputDir,
			"git", "clone", repo, dstPath,
		); err != nil {
			return fmt.Errorf("failed to clone repo: %v\n%s", err, out)
		}

		logger.Infof("updating submodules (if any)...")
		if out, err := command(ctx, dstPath,
			"git", "submodule", "update",
			"--init",
			"--recursive",
		); err != nil {
			return fmt.Errorf("failed to update submodules: %v\n%s", err, out)
		}
		return nil
	} else {
		return fmt.Errorf("failed to check for existing file (%s): %v", dstPath, err)
	}
}

func checkForCLIs(ctx context.Context, clis ...string) error {
	logger := logging.FromContext(ctx)
	for _, cli := range clis {
		logger.Infof("Checking for %s CLI...", cli)
		if _, err := exec.LookPath(cli); err != nil {
			return fmt.Errorf("CLI %s was not found in the path: %v\nNOTE: It is highly recommmended that this command be ran on cloud shell to ensure the correct CLI versions are available.", cli, err)
		}
	}
	return nil
}

func command(ctx context.Context, outputDir string, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = outputDir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func goCommand(ctx context.Context, outputDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = outputDir
	cmd.Env = []string{
		"GOPROXY=https://proxy.golang.org",
		"GOSUMDB=sum.golang.org",
		"GO111MODULE=on",
		"CGO_ENABLED=0",
		"GOOS=linux",
		"GOCACHE=" + filepath.Join(outputDir, "gocache"),
		"GOPATH=" + outputDir,
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// copyFile does more or less what `cp` does. As of now (when this function
// was committed), the Go stdlib did not offer a way to copy files across file
// systems.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file: %v", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to open dst file: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %v", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat: %v", err)
	}
	if err := os.Chmod(dst, si.Mode()); err != nil {
		return fmt.Errorf("failed to chmod: %v", err)
	}

	return nil
}

func hashedPath(outputDir, buildpackName string) string {
	// The builder will look for the buildpack at a very specific path. The
	// path is the hex value of the name hashed.
	hash := fmt.Sprintf("%x", md5.Sum([]byte("v2-buildpack")))
	return filepath.Join(buildpackPath(outputDir, buildpackName), "bin", "v2-buildpack", hash)
}
