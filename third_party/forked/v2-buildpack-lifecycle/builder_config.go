package buildpackapplifecycle

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type LifecycleBuilderConfig struct {
	*flag.FlagSet
	workingDir     string
	ExecutablePath string
}

const (
	lifecycleBuilderBuildDirFlag                  = "buildDir"
	lifecycleBuilderOutputDropletFlag             = "outputDroplet"
	lifecycleBuilderOutputMetadataFlag            = "outputMetadata"
	lifecycleBuilderOutputBuildArtifactsCacheFlag = "outputBuildArtifactsCache"
	lifecycleBuilderBuildpacksDirFlag             = "buildpacksDir"
	lifecycleBuilderBuildpacksDownloadDirFlag     = "buildpacksDownloadDir"
	lifecycleBuilderBuildArtifactsCacheDirFlag    = "buildArtifactsCacheDir"
	lifecycleBuilderBuildpackOrderFlag            = "buildpackOrder"
	lifecycleBuilderSkipDetect                    = "skipDetect"
	lifecycleBuilderSkipCertVerify                = "skipCertVerify"
)

var lifecycleBuilderDefaults = map[string]string{
	lifecycleBuilderBuildDirFlag:                  "/tmp/app",
	lifecycleBuilderOutputDropletFlag:             "/tmp/droplet",
	lifecycleBuilderOutputMetadataFlag:            "/tmp/result.json",
	lifecycleBuilderOutputBuildArtifactsCacheFlag: "/tmp/output-cache",
	lifecycleBuilderBuildpacksDirFlag:             "/tmp/buildpacks",
	lifecycleBuilderBuildpacksDownloadDirFlag:     "/tmp/buildpackdownloads",
	lifecycleBuilderBuildArtifactsCacheDirFlag:    "/tmp/cache",
}

func NewLifecycleBuilderConfig(buildpacks []string, skipDetect bool, skipCertVerify bool) LifecycleBuilderConfig {
	flagSet := flag.NewFlagSet("builder", flag.ExitOnError)

	flagSet.String(
		lifecycleBuilderBuildDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildDirFlag],
		"directory containing raw app bits",
	)

	flagSet.String(
		lifecycleBuilderOutputDropletFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputDropletFlag],
		"file where compressed droplet should be written",
	)

	flagSet.String(
		lifecycleBuilderOutputMetadataFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputMetadataFlag],
		"directory in which to write the app metadata",
	)

	flagSet.String(
		lifecycleBuilderOutputBuildArtifactsCacheFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputBuildArtifactsCacheFlag],
		"file where compressed contents of new cached build artifacts should be written",
	)

	flagSet.String(
		lifecycleBuilderBuildpacksDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	flagSet.String(
		lifecycleBuilderBuildpacksDownloadDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildpacksDownloadDirFlag],
		"directory to download buildpacks to",
	)

	flagSet.String(
		lifecycleBuilderBuildArtifactsCacheDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildArtifactsCacheDirFlag],
		"directory where previous cached build artifacts should be extracted",
	)

	flagSet.String(
		lifecycleBuilderBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	flagSet.Bool(
		lifecycleBuilderSkipDetect,
		skipDetect,
		"skip buildpack detect",
	)

	flagSet.Bool(
		lifecycleBuilderSkipCertVerify,
		skipCertVerify,
		"skip SSL certificate verification",
	)

	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	return LifecycleBuilderConfig{
		FlagSet:        flagSet,
		workingDir:     wd,
		ExecutablePath: "/tmp/lifecycle/builder",
	}
}

func (s LifecycleBuilderConfig) Path() string {
	return s.getPath(s.ExecutablePath)
}

func (s LifecycleBuilderConfig) Args() []string {
	argv := []string{}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s=%s", flag.Name, flag.Value.String()))
	})

	return argv
}

func (s LifecycleBuilderConfig) Validate() error {
	var validationError ValidationError

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		value := flag.Value.String()
		if value == "" {
			validationError = validationError.Append(fmt.Errorf("missing flag: -%s", flag.Name))
		}
	})

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (s LifecycleBuilderConfig) BuildDir() string {
	return s.getPath(s.Lookup(lifecycleBuilderBuildDirFlag).Value.String())
}

func (s LifecycleBuilderConfig) BuildpackPath(buildpackName string) string {
	baseDir := s.BuildpacksDir()
	buildpackURL, err := url.Parse(buildpackName)
	if err == nil && buildpackURL.IsAbs() {
		baseDir = s.BuildpacksDownloadDir()
	}
	return filepath.Join(baseDir, fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
}

func (s LifecycleBuilderConfig) BuildpackOrder() []string {
	buildpackOrder := s.Lookup(lifecycleBuilderBuildpackOrderFlag).Value.String()
	return strings.Split(buildpackOrder, ",")
}

func (s LifecycleBuilderConfig) SupplyBuildpacks() []string {
	numBuildpacks := len(s.BuildpackOrder())
	if !s.SkipDetect() || numBuildpacks == 0 {
		return []string{}
	}
	return s.BuildpackOrder()[0 : numBuildpacks-1]
}

func (s LifecycleBuilderConfig) DepsIndex(i int) string {
	numBuildpacks := len(s.SupplyBuildpacks()) + 1
	padDigits := int(math.Log10(float64(numBuildpacks))) + 1
	indexFormat := fmt.Sprintf("%%0%dd", padDigits)
	return fmt.Sprintf(indexFormat, i)
}

func (s LifecycleBuilderConfig) BuildpacksDir() string {
	return s.getPath(s.Lookup(lifecycleBuilderBuildpacksDirFlag).Value.String())
}

func (s LifecycleBuilderConfig) BuildpacksDownloadDir() string {
	return s.getPath(s.Lookup(lifecycleBuilderBuildpacksDownloadDirFlag).Value.String())
}

func (s LifecycleBuilderConfig) BuildArtifactsCacheDir() string {
	return s.getPath(s.Lookup(lifecycleBuilderBuildArtifactsCacheDirFlag).Value.String())
}

func (s LifecycleBuilderConfig) OutputDroplet() string {
	return s.getPath(s.Lookup(lifecycleBuilderOutputDropletFlag).Value.String())
}

func (s LifecycleBuilderConfig) OutputMetadata() string {
	return s.getPath(s.Lookup(lifecycleBuilderOutputMetadataFlag).Value.String())
}

func (s LifecycleBuilderConfig) OutputBuildArtifactsCache() string {
	return s.getPath(s.Lookup(lifecycleBuilderOutputBuildArtifactsCacheFlag).Value.String())
}

func (s LifecycleBuilderConfig) SkipCertVerify() bool {
	return s.Lookup(lifecycleBuilderSkipCertVerify).Value.String() == "true"
}

func (s LifecycleBuilderConfig) SkipDetect() bool {
	return s.Lookup(lifecycleBuilderSkipDetect).Value.String() == "true"
}

type ValidationError []error

func (ve ValidationError) Append(err error) ValidationError {
	switch err := err.(type) {
	case ValidationError:
		return append(ve, err...)
	default:
		return append(ve, err)
	}
}

func (ve ValidationError) Error() string {
	var buffer bytes.Buffer

	for i, err := range ve {
		if err == nil {
			continue
		}
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(err.Error())
	}

	return buffer.String()
}

func (ve ValidationError) Empty() bool {
	return len(ve) == 0
}
