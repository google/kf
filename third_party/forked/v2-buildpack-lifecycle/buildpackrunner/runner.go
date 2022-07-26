package buildpackrunner

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner/resources"
	"code.cloudfoundry.org/bytefmt"
)

const DOWNLOAD_TIMEOUT = 10 * time.Minute

type Runner struct {
	config      *buildpackapplifecycle.LifecycleBuilderConfig
	depsDir     string
	contentsDir string
	profileDir  string
}

type descriptiveError struct {
	message string
	err     error
}

type Release struct {
	DefaultProcessTypes buildpackapplifecycle.ProcessTypes `yaml:"default_process_types"`
}

func (e descriptiveError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func newDescriptiveError(err error, message string, args ...interface{}) error {
	if len(args) == 0 {
		return descriptiveError{message: message, err: err}
	}
	return descriptiveError{message: fmt.Sprintf(message, args...), err: err}
}

func New(config *buildpackapplifecycle.LifecycleBuilderConfig) *Runner {
	return &Runner{
		config: config,
	}
}

func (runner *Runner) GetDepsDir() string {
	return runner.depsDir
}

func (runner *Runner) GetContentsDir() string {
	return runner.contentsDir
}

func (runner *Runner) ProcessYML(selectedBuildpacks []string) (resources.LaunchData, error) {
	var launchYML resources.LaunchData
	var err error

	for index := range selectedBuildpacks {
		if launchYML, err = runner.MergeLaunchYML(index, launchYML); err != nil {
			return resources.LaunchData{}, err
		}
	}
	return launchYML, nil
}

func (runner *Runner) MergeLaunchYML(buildpackIndex int, procMap resources.LaunchData) (resources.LaunchData, error) {
	if runner.launchYMLExists(buildpackIndex) {
		launchYMLPath := filepath.Join(runner.depsDir, strconv.Itoa(buildpackIndex), "launch.yml")

		data, err := ioutil.ReadFile(launchYMLPath)
		if err != nil {
			return resources.LaunchData{}, err
		}

		var buildPackLaunchYML resources.LaunchData
		if err := yaml.Unmarshal(data, &buildPackLaunchYML); err != nil {
			return resources.LaunchData{}, err
		}

		procMap.Processes = resources.MergeProcesses(procMap.Processes, buildPackLaunchYML.Processes)
	}
	return procMap, nil
}

func (runner *Runner) ProcessFinalBuildpack(detectedBuildpack, detectedBuildpackDir string, procMap resources.LaunchData) (resources.LaunchData, error) {

	buildpackIndex := len(runner.config.BuildpackOrder()) - 1

	if runner.launchYMLExists(buildpackIndex) {
		var err error
		if procMap, err = runner.MergeLaunchYML(buildpackIndex, procMap); err != nil {
			return resources.LaunchData{}, err
		}
	} else {
		releaseInfo, err := runner.release(detectedBuildpackDir, map[string]string{})
		if err != nil {
			return resources.LaunchData{}, newDescriptiveError(err, buildpackapplifecycle.ReleaseFailMsg)
		}

		procMap.Processes = resources.MergeProcesses(
			procMap.Processes,
			resources.ProcDataToProcesses(releaseInfo.DefaultProcessTypes),
		)
	}

	startCommands, err := runner.readProcfile()
	if err != nil {
		return resources.LaunchData{}, newDescriptiveError(err, "Failed to read command from Procfile")
	}

	procMap.Processes = resources.MergeProcesses(
		procMap.Processes,
		resources.ProcDataToProcesses(startCommands),
	)

	return procMap, nil
}

func (runner *Runner) WriteStagingInfoYML(resultData buildpackapplifecycle.StagingResult, buildpacks []buildpackapplifecycle.BuildpackMetadata) (string, error) {
	stagingInfoYML := filepath.Join(runner.contentsDir, "staging_info.yml")
	stagingInfoFile, err := os.Create(stagingInfoYML)
	if err != nil {
		return "", err
	}
	defer stagingInfoFile.Close()

	var lastBuildpack buildpackapplifecycle.BuildpackMetadata
	if len(buildpacks) > 0 {
		lastBuildpack = buildpacks[len(buildpacks)-1]
	}

	err = json.NewEncoder(stagingInfoFile).Encode(DeaStagingInfo{
		DetectedBuildpack: lastBuildpack.Name,
		StartCommand:      resultData.ProcessTypes["web"],
	})
	if err != nil {
		return "", err
	}
	return stagingInfoYML, nil
}

func (runner *Runner) WriteResultJSON(resultData buildpackapplifecycle.StagingResult, buildpacks []buildpackapplifecycle.BuildpackMetadata) (string, error) {
	var lastBuildpack buildpackapplifecycle.BuildpackMetadata
	if len(buildpacks) > 0 {
		lastBuildpack = buildpacks[len(buildpacks)-1]
	}

	resultData = buildpackapplifecycle.UpdateStagingResult(resultData, buildpackapplifecycle.LifecycleMetadata{
		BuildpackKey:      lastBuildpack.Key,
		DetectedBuildpack: lastBuildpack.Name,
		Buildpacks:        buildpacks,
	})

	resultPath := runner.config.OutputMetadata()
	resultFile, err := os.Create(resultPath)
	if err != nil {
		return "", err
	}
	defer resultFile.Close()

	if err := json.NewEncoder(resultFile).Encode(resultData); err != nil {
		return "", err
	}
	return resultPath, nil
}

func (runner *Runner) WriteStartCommands(detectedBuildpackDir, detectedBuildpack, detectOutput string, procMap resources.LaunchData) (string, string, error) {
	resultData := resources.ConvertToResult(procMap)

	if resultData.ProcessTypes["web"] == "" {
		printError("No start command specified by buildpack or via Procfile.")
		printError("App will not start unless a command is provided at runtime.")
	}

	var buildpacks []buildpackapplifecycle.BuildpackMetadata
	if runner.config.SkipDetect() {
		buildpacks = runner.buildpacksMetadata(runner.config.BuildpackOrder())
	} else {
		buildpacks = runner.buildpacksMetadata([]string{detectedBuildpack})
		if buildpacks[0].Name == "" {
			buildpacks[0].Name = detectOutput
		}
	}

	stagingInfoPath, err := runner.WriteStagingInfoYML(resultData, buildpacks)
	if err != nil {
		return "", "", err
	}

	resultDataPath, err := runner.WriteResultJSON(resultData, buildpacks)
	if err != nil {
		return "", "", err
	}
	return resultDataPath, stagingInfoPath, nil
}

func (runner *Runner) Setup() error {
	if err := runner.makeDirectories(); err != nil {
		return newDescriptiveError(err, "Failed to set up filesystem when generating droplet")
	}

	if err := runner.downloadBuildpacks(); err != nil {
		return err
	}
	return runner.cleanCacheDir()
}

func (runner *Runner) GoLikeLightning() (string, string, error) {
	var detectedBuildpack, detectOutput, detectedBuildpackDir string
	var ok bool
	var err error

	if runner.config.SkipDetect() {
		detectedBuildpack, detectedBuildpackDir, err = runner.runSupplyBuildpacks()
		if err != nil {
			return "", "", err
		}
	} else {
		detectedBuildpack, detectedBuildpackDir, detectOutput, ok = runner.detect()
		if !ok {
			return "", "", newDescriptiveError(nil, buildpackapplifecycle.DetectFailMsg)
		}
	}

	if err := runner.runFinalize(detectedBuildpackDir); err != nil {
		return "", "", newDescriptiveError(err, buildpackapplifecycle.CompileFailMsg)
	}

	procMap, err := runner.ProcessYML(runner.config.SupplyBuildpacks())
	if err != nil {
		return "", "", err
	}

	procMap, err = runner.ProcessFinalBuildpack(detectedBuildpack, detectedBuildpackDir, procMap)
	if err != nil {
		return "", "", err
	}

	resultJSONPath, stagingInfoYMLPath, err := runner.WriteStartCommands(detectedBuildpackDir, detectedBuildpack, detectOutput, procMap)
	if err != nil {
		return "", "", err
	}

	for _, name := range []string{"tmp", "logs"} {
		if err := os.MkdirAll(filepath.Join(runner.contentsDir, name), 0755); err != nil {
			return "", "", newDescriptiveError(err, "Failed to set up droplet filesystem")
		}
	}

	appDir := filepath.Join(runner.contentsDir, "app")
	err = runner.copyApp(runner.config.BuildDir(), appDir)
	if err != nil {
		return "", "", newDescriptiveError(err, "Failed to copy compiled droplet")
	}

	tarPath, err := runner.findTar()
	if err != nil {
		return "", "", newDescriptiveError(err, "Unable to find tar executable")
	}

	err = exec.Command(tarPath, "-czf", runner.config.OutputDroplet(), "-C", runner.contentsDir, ".").Run()
	if err != nil {
		return "", "", newDescriptiveError(err, "Failed to compress droplet filesystem")
	}

	//prepare the build artifacts cache output directory
	err = os.MkdirAll(filepath.Dir(runner.config.OutputBuildArtifactsCache()), 0755)
	if err != nil {
		return "", "", newDescriptiveError(err, "Failed to create output build artifacts cache dir")
	}

	err = exec.Command(tarPath, "-czf", runner.config.OutputBuildArtifactsCache(), "-C", runner.config.BuildArtifactsCacheDir(), ".").Run()
	if err != nil {
		return "", "", newDescriptiveError(err, "Failed to compress build artifacts")
	}

	return resultJSONPath, stagingInfoYMLPath, nil
}

func (runner *Runner) Run() (string, error) {
	if err := runner.Setup(); err != nil {
		return "", err
	}

	_, stagingInfo, err := runner.GoLikeLightning()
	return stagingInfo, err
}

func (runner *Runner) CleanUp() error {
	if runner.contentsDir == "" {
		return nil
	}
	return os.RemoveAll(runner.contentsDir)
}

func (runner *Runner) launchYMLExists(index int) bool {
	launchYMLPath := filepath.Join(runner.depsDir, strconv.Itoa(index), "launch.yml")
	if _, err := os.Stat(launchYMLPath); err == nil {
		return true
	}
	return false
}

func (runner *Runner) hasLaunchYML(selectedBuildpacks []string) bool {
	for index := range selectedBuildpacks {
		if runner.launchYMLExists(index) {
			return true
		}
	}
	return false
}

func (runner *Runner) buildpacksMetadata(buildpacks []string) []buildpackapplifecycle.BuildpackMetadata {
	data := make([]buildpackapplifecycle.BuildpackMetadata, len(buildpacks))
	for i, key := range buildpacks {
		data[i].Key = key
		configPath := filepath.Join(runner.depsDir, runner.config.DepsIndex(i), "config.yml")
		if contents, err := ioutil.ReadFile(configPath); err == nil {
			configyaml := struct {
				Name    string `yaml:"name"`
				Version string `yaml:"version"`
			}{}
			if err := yaml.Unmarshal(contents, &configyaml); err == nil {
				data[i].Name = configyaml.Name
				data[i].Version = configyaml.Version
			}
		}
	}
	return data
}

func (runner *Runner) makeDirectories() error {
	if err := os.MkdirAll(filepath.Dir(runner.config.OutputDroplet()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(runner.config.OutputMetadata()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(runner.config.BuildArtifactsCacheDir(), "final"), 0755); err != nil {
		return err
	}

	for _, buildpack := range runner.config.SupplyBuildpacks() {
		if err := os.MkdirAll(runner.supplyCachePath(buildpack), 0755); err != nil {
			return err
		}
	}

	var err error
	runner.contentsDir, err = ioutil.TempDir("", "contents")
	if err != nil {
		return err
	}

	runner.depsDir = filepath.Join(runner.contentsDir, "deps")

	for i := 0; i <= len(runner.config.SupplyBuildpacks()); i++ {
		if err := os.MkdirAll(filepath.Join(runner.depsDir, runner.config.DepsIndex(i)), 0755); err != nil {
			return err
		}
	}

	runner.profileDir = filepath.Join(runner.contentsDir, "profile.d")
	if err := os.MkdirAll(runner.profileDir, 0755); err != nil {
		return err
	}

	return nil
}

func (runner *Runner) downloadBuildpacks() error {
	// Do we have a custom buildpack?
	for _, buildpackName := range runner.config.BuildpackOrder() {
		buildpackURL, err := url.Parse(buildpackName)
		if err != nil {
			return fmt.Errorf("Invalid buildpack url (%s): %s", buildpackName, err.Error())
		}
		if !buildpackURL.IsAbs() {
			continue
		}

		destination := runner.config.BuildpackPath(buildpackName)

		var downloadErr error
		if IsZipFile(buildpackURL.Path) {
			var size uint64

			zipDownloader := NewZipDownloader(runner.config.SkipCertVerify())
			size, downloadErr = zipDownloader.DownloadAndExtract(buildpackURL, destination)
			if downloadErr == nil {
				fmt.Printf("Downloaded buildpack `%s` (%s)\n", buildpackURL.String(), bytefmt.ByteSize(size))
			}
		} else {
			downloadErr = GitClone(*buildpackURL, destination)
		}
		if downloadErr != nil {
			return downloadErr
		}
	}

	return nil
}

func (runner *Runner) cleanCacheDir() error {
	neededCacheDirs := map[string]bool{
		filepath.Join(runner.config.BuildArtifactsCacheDir(), "final"): true,
	}

	for _, bp := range runner.config.SupplyBuildpacks() {
		neededCacheDirs[runner.supplyCachePath(bp)] = true
	}

	dirs, err := ioutil.ReadDir(runner.config.BuildArtifactsCacheDir())
	if err != nil {
		return err
	}

	for _, dirInfo := range dirs {
		dir := filepath.Join(runner.config.BuildArtifactsCacheDir(), dirInfo.Name())
		if !neededCacheDirs[dir] {
			err = os.RemoveAll(dir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (runner *Runner) buildpackPath(buildpack string) (string, error) {
	buildpackPath := runner.config.BuildpackPath(buildpack)

	if runner.pathHasBinDirectory(buildpackPath) {
		return buildpackPath, nil
	}

	files, err := ioutil.ReadDir(buildpackPath)
	if err != nil {
		return "", newDescriptiveError(nil, "Failed to read buildpack directory '%s' for buildpack '%s'", buildpackPath, buildpack)
	}

	if len(files) == 1 {
		nestedPath := filepath.Join(buildpackPath, files[0].Name())

		if runner.pathHasBinDirectory(nestedPath) {
			return nestedPath, nil
		}
	}

	return "", newDescriptiveError(nil, "malformed buildpack does not contain a /bin dir: %s", buildpack)
}

func (runner *Runner) pathHasBinDirectory(pathToTest string) bool {
	_, err := os.Stat(filepath.Join(pathToTest, "bin"))
	return err == nil
}

func (runner *Runner) supplyCachePath(buildpack string) string {
	return filepath.Join(runner.config.BuildArtifactsCacheDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpack))))
}

func fileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// returns buildpack path, ok
func (runner *Runner) runSupplyBuildpacks() (string, string, error) {
	if err := runner.validateSupplyBuildpacks(); err != nil {
		return "", "", err
	}
	for i, buildpack := range runner.config.SupplyBuildpacks() {
		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			return "", "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}

		err = runner.run(exec.Command(filepath.Join(buildpackPath, "bin", "supply"), runner.config.BuildDir(), runner.supplyCachePath(buildpack), runner.depsDir, runner.config.DepsIndex(i)), os.Stdout)
		if err != nil {
			return "", "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}
	}

	finalBuildpack := runner.config.BuildpackOrder()[len(runner.config.SupplyBuildpacks())]
	finalPath, err := runner.buildpackPath(finalBuildpack)
	if err != nil {
		return "", "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
	}

	return finalBuildpack, finalPath, nil
}

func (runner *Runner) validateSupplyBuildpacks() error {
	for _, buildpack := range runner.config.SupplyBuildpacks() {
		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			return newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}

		if hasSupply, err := hasSupply(buildpackPath); err != nil {
			return newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		} else if !hasSupply {
			return newDescriptiveError(err, buildpackapplifecycle.NoSupplyScriptFailMsg)
		}
	}
	return nil
}

func (runner *Runner) runFinalize(buildpackPath string) error {
	depsIdx := runner.config.DepsIndex(len(runner.config.SupplyBuildpacks()))
	cacheDir := filepath.Join(runner.config.BuildArtifactsCacheDir(), "final")

	hasFinalize, err := hasFinalize(buildpackPath)
	if err != nil {
		return newDescriptiveError(err, buildpackapplifecycle.FinalizeFailMsg)
	}

	if hasFinalize {
		hasSupply, err := hasSupply(buildpackPath)
		if err != nil {
			return newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}

		if hasSupply {
			if err := runner.run(exec.Command(filepath.Join(buildpackPath, "bin", "supply"), runner.config.BuildDir(), cacheDir, runner.depsDir, depsIdx), os.Stdout); err != nil {
				return newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
			}
		}

		if err := runner.run(exec.Command(filepath.Join(buildpackPath, "bin", "finalize"), runner.config.BuildDir(), cacheDir, runner.depsDir, depsIdx, runner.profileDir), os.Stdout); err != nil {
			return newDescriptiveError(err, buildpackapplifecycle.FinalizeFailMsg)
		}
	} else {
		if len(runner.config.SupplyBuildpacks()) > 0 {
			printError(buildpackapplifecycle.MissingFinalizeWarnMsg)
		}

		// remove unused deps sub dir
		if err := os.RemoveAll(filepath.Join(runner.depsDir, depsIdx)); err != nil {
			return newDescriptiveError(err, buildpackapplifecycle.CompileFailMsg)
		}

		if err := runner.run(exec.Command(filepath.Join(buildpackPath, "bin", "compile"), runner.config.BuildDir(), cacheDir), os.Stdout); err != nil {
			return newDescriptiveError(err, buildpackapplifecycle.CompileFailMsg)
		}
	}

	return nil
}

// returns buildpack name,  buildpack path, buildpack detect output, ok
func (runner *Runner) detect() (string, string, string, bool) {
	for _, buildpack := range runner.config.BuildpackOrder() {

		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			continue
		}

		if runner.config.SkipDetect() {
			return buildpack, buildpackPath, "", true
		}

		if err := runner.warnIfDetectNotExecutable(buildpackPath); err != nil {
			printError(err.Error())
			continue
		}

		output := new(bytes.Buffer)
		err = runner.run(exec.Command(filepath.Join(buildpackPath, "bin", "detect"), runner.config.BuildDir()), output)

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\r\n"), true
		}
	}

	return "", "", "", false
}

func (runner *Runner) readProcfile() (map[string]string, error) {
	processes := map[string]string{}

	procFile, err := ioutil.ReadFile(filepath.Join(runner.config.BuildDir(), "Procfile"))
	if err != nil {
		if os.IsNotExist(err) {
			// Procfiles are optional
			return processes, nil
		}

		return processes, err
	}

	err = yaml.Unmarshal(procFile, &processes)
	if err != nil {
		// clobber yaml parsing  error
		return processes, errors.New("invalid YAML")
	}

	return processes, nil
}

func (runner *Runner) release(buildpackDir string, startCommands map[string]string) (Release, error) {
	output := new(bytes.Buffer)

	err := runner.run(exec.Command(filepath.Join(buildpackDir, "bin", "release"), runner.config.BuildDir()), output)
	if err != nil {
		return Release{}, err
	}

	parsedRelease := Release{}

	err = yaml.Unmarshal(output.Bytes(), &parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	if len(startCommands) > 0 { // passed in a start command from the procfile
		if len(parsedRelease.DefaultProcessTypes) == 0 { // if there are no default processes
			parsedRelease.DefaultProcessTypes = startCommands // just use procfile start commands object
		} else { // otherwise overwrite default start commands of the same type with ones from the procfile
			for k, v := range startCommands {
				parsedRelease.DefaultProcessTypes[k] = v
			}
		}
	}

	return parsedRelease, nil
}

// Writes both results.json file and staging_info.yaml
func (runner *Runner) saveInfo(infoFilePath string, buildpacks []buildpackapplifecycle.BuildpackMetadata, releaseInfo Release) error {
	deaInfoFile, err := os.Create(infoFilePath)
	if err != nil {
		return err
	}
	defer deaInfoFile.Close()

	var lastBuildpack buildpackapplifecycle.BuildpackMetadata
	if len(buildpacks) > 0 {
		lastBuildpack = buildpacks[len(buildpacks)-1]
	}

	err = json.NewEncoder(deaInfoFile).Encode(DeaStagingInfo{
		DetectedBuildpack: lastBuildpack.Name,
		StartCommand:      releaseInfo.DefaultProcessTypes["web"],
	})
	if err != nil {
		return err
	}

	resultFile, err := os.Create(runner.config.OutputMetadata())
	if err != nil {
		return err
	}
	defer resultFile.Close()

	err = json.NewEncoder(resultFile).Encode(buildpackapplifecycle.NewStagingResult(
		releaseInfo.DefaultProcessTypes,
		buildpackapplifecycle.LifecycleMetadata{
			BuildpackKey:      lastBuildpack.Key,
			DetectedBuildpack: lastBuildpack.Name,
			Buildpacks:        buildpacks,
		},
	))
	if err != nil {
		return err
	}

	return nil
}

func (runner *Runner) run(cmd *exec.Cmd, output io.Writer) error {
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func printError(message string) {
	fmt.Fprintln(os.Stderr, message)
}
