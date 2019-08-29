// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// EnvDockerRegistry is the environment variable used to store the
	// registry we will push containers to for the integration tests. If this
	// is not set, then the integration tests are skipped.
	EnvDockerRegistry = "DOCKER_REGISTRY"

	// IntegrationTestDomain is the domain that the integration test spaces
	// are setup with.
	IntegrationTestDomain = "integration-tests.kf.dev"
)

// DockerRegistry returns the configured docker registry for the integration
// tests.
func DockerRegistry() string {
	return os.Getenv(EnvDockerRegistry)
}

// RunIntegrationTest skips the tests if testing.Short() is true (via --short
// flag) or if DOCKER_REGISTRY is not set. Otherwise it runs the given test.
func RunIntegrationTest(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	if !strings.HasPrefix(t.Name(), "TestIntegration_") {
		// We want to enforce a convention so scripts can single out
		// integration tests.
		t.Fatalf("Integration tests must have the name format of 'TestIntegration_XXX`")
	}

	t.Helper()
	if testing.Short() {
		t.Skip()
	}

	projID := os.Getenv(EnvDockerRegistry)
	if projID == "" {
		t.Skipf("%s is required for integration tests... Skipping...", EnvDockerRegistry)
	}

	start := time.Now()
	defer func() {
		state := "PASSED"
		if t.Failed() {
			state = "FAILED"
		}
		t.Logf("Test %s took %v and %s", t.Name(), time.Since(start), state)
	}()

	// Setup context that will allow us to cleanup if the user wants to
	// cancel the tests.
	ctx, cancel := context.WithCancel(context.Background())

	// Give everything time to clean up.
	defer time.Sleep(time.Second)
	defer cancel()
	CancelOnSignal(ctx, cancel, t.Log)
	t.Log()

	test(ctx, t)
}

// CancelOnSignal watches for a kill signal. If it gets one, it invokes the
// cancel function. An aditional signal will exit the process. If the given context finishes, the underlying go-routine
// finishes.
func CancelOnSignal(ctx context.Context, cancel func(), log func(args ...interface{})) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		select {
		case <-c:
			log("Signal received... Cleaning up... (Hit Ctrl-C again to quit immediately)")
			cancel()
			<-c
			os.Exit(1)
		case <-ctx.Done():
			return
		}
	}()
}

// KfTestConfig is a configuration for a Kf Test.
type KfTestConfig struct {
	Args []string
	Env  map[string]string
}

// KfTestOutput is the output from `kf`. Note, this output is while kf is
// running.
type KfTestOutput struct {
	Stdout io.Reader
	Stderr io.Reader
	Stdin  io.Writer
	Done   <-chan struct{}
}

func kf(ctx context.Context, t *testing.T, binaryPath string, cfg KfTestConfig) (KfTestOutput, <-chan error) {
	t.Helper()

	Logf(t, "kf %s\n", strings.Join(cfg.Args, " "))

	cmd := exec.CommandContext(ctx, binaryPath, cfg.Args...)
	for name, value := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to fetch Stdout pipe: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to fetch Stderr pipe: %s", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to fetch Stdin pipe: %s", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start kf: %s", err)
	}

	done := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		errs <- cmd.Wait()
		close(done)
	}()

	return KfTestOutput{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  stdin,
		Done:   done,
	}, errs
}

// KfInvoker will synchronously invoke `kf` with the given configuration.
type KfInvoker func(context.Context, *testing.T, KfTestConfig) (KfTestOutput, <-chan error)

// KfTest is a test ran by RunKfTest.
type KfTest func(ctx context.Context, t *testing.T, kf *Kf)

// RunKfTest runs 'kf' for integration tests. It first compiles 'kf' and then
// launches it as a sub-process. It will set the args and environment
// variables accordingly. It will run the given test with the resulting
// STDOUT, STDERR and STDIN. It will cleanup the sub-process on completion via
// the context.
func RunKfTest(t *testing.T, test KfTest) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	CancelOnSignal(ctx, cancel, t.Log)
	kfPath := CompileKf(ctx, t)

	RunIntegrationTest(t, func(ctx context.Context, t *testing.T) {
		t.Helper()

		kf := KF(t, func(ctx context.Context, t *testing.T, cfg KfTestConfig) (KfTestOutput, <-chan error) {
			return kf(ctx, t, kfPath, cfg)
		})

		// Create the space
		spaceName := fmt.Sprintf("apps-integration-test-%d", time.Now().UnixNano())
		kf.CreateSpace(ctx, spaceName)
		defer kf.DeleteSpace(ctx, spaceName)

		// Wait for space to become ready
		// TODO(#371): create-space should wait until the space is ready.
		RetryOnPanic(ctx, t, func() {
			for _, s := range kf.Spaces(ctx) {
				if strings.HasPrefix(s, spaceName) &&
					// Ensure space is marked "Ready True"
					regexp.MustCompile(`\sTrue\s`).MatchString(s) {
					return
				}
			}
			panic(fmt.Sprintf("%s-> did not find space %s", t.Name(), spaceName))
		})

		ctx = ContextWithSpace(ctx, spaceName)

		test(ctx, t, kf)
	})
}

var (
	compileKfOnce sync.Once
	compiledKf    string
)

// CompileKf compiles the `kf` binary. It returns a string to the resulting
// binary.
func CompileKf(ctx context.Context, t *testing.T) string {
	t.Helper()
	compileKfOnce.Do(func() {
		t.Helper()
		compiledKf = Compile(ctx, t, "./cmd/kf")
	})
	return compiledKf
}

// Compile compiles a path in the repo. It returns a path to the resulting
// binary. codePath must be relative to RootDir.
func Compile(ctx context.Context, t *testing.T, codePath string) string {
	t.Helper()

	var err error
	codePath, err = filepath.Abs(filepath.Join(RootDir(ctx, t), codePath))
	if err != nil {
		t.Fatalf("failed to convert to absolute path: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get CWD: %s", err)
	}

	codePath, err = filepath.Rel(wd, codePath)
	if err != nil {
		t.Fatalf("failed to get relative code path: %s", err)
	}
	codePath = "./" + codePath

	tmpDir, err := ioutil.TempDir("", "kf_test_compile")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %s", err)
	}

	go func() {
		<-ctx.Done()
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("WARNING: failed to cleanup temp dir: %s: %s", tmpDir, err)
		}
	}()

	outputPath := filepath.Join(tmpDir, "out")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, ".")
	cmd.Dir = codePath
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go build (err=%s): %s", err, output)
	}

	return outputPath
}

// RootDir uses git to find the root of the directory. This is useful for
// tests (especially integration ones that compile things).
func RootDir(ctx context.Context, t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get Stdout pipe for RootDir: %s", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command to get RootDir %s", err)
	}

	data, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Fatalf("failed to read Stdout pipe for RootDir: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("failed to get RootDir %s", err)
	}

	return strings.TrimSpace(string(data))
}

// RetryOnPanic will retry the function if it panics (e.g., via PanicOnError)
// until the given context is cancelled or the given function succeeds without
// panicking.
func RetryOnPanic(ctx context.Context, t *testing.T, f func()) {
	var success bool
	for !success && ctx.Err() == nil {
		func() {
			defer func() {
				if err := recover(); err != nil {
					Logf(t, "got err: %s. Retrying...", err)
					return
				}
				success = true
			}()

			f()
		}()
	}
}

// PanicOnError launches a go routine and waits for either the context or an
// error. An error will result in a panic. Why a panic in a test? t.Fatal is
// not thread safe. T.Failed() is thread safe, therefore we can use that to
// prevent the panic.
func PanicOnError(ctx context.Context, t *testing.T, messagePrefix string, errs <-chan error) {
	go func() {
		t.Helper()
		select {
		case <-ctx.Done():
			return
		case err := <-errs:
			if err != nil {
				if t.Failed() {
					Logf(t, "%s: %s", messagePrefix, err)
					return
				} else {
					panic(fmt.Sprintf("%s->%s: %s", t.Name(), messagePrefix, err))
				}
			}
		}
	}()
}

// CombineOutputStr writes the Stdout and Stderr string. It will return when
// ctx is done.
func CombineOutputStr(ctx context.Context, t *testing.T, out KfTestOutput) []string {
	lines := CombineOutput(ctx, t, out)

	var result []string
	for {
		select {
		case <-ctx.Done():
			return result
		case <-out.Done:
			return result
		case line, ok := <-lines:
			if ok {
				result = append(result, line)
			}
		}
	}
}

// CombineOutput writes the Stdout and Stderr to a channel for each line.
func CombineOutput(ctx context.Context, t *testing.T, out KfTestOutput) <-chan string {
	lines := make(chan string)

	var wg sync.WaitGroup
	f := func(r io.Reader) {
		defer wg.Done()
		s := bufio.NewScanner(r)
		for s.Scan() {
			select {
			case <-ctx.Done():
				return
			case lines <- s.Text():
			}
		}
	}

	wg.Add(2)
	go f(out.Stdout)
	go f(out.Stderr)
	go func() {
		wg.Wait()
		close(lines)
	}()

	return lines
}

// Logf will write to Stderr if testing.Verbose is true and t.Log otherwise.
// This is so logs will stream out instead of only being displayed at the end
// of the test.
func Logf(t *testing.T, format string, i ...interface{}) {
	line := fmt.Sprintf(format, i...)
	lineWithPrefix := fmt.Sprintf("[%s] %s", t.Name(), line)

	if testing.Verbose() {
		fmt.Fprintln(os.Stderr, lineWithPrefix)
		return
	}

	t.Logf(lineWithPrefix)
}

// StreamOutput writes the output of KfTestOutput to the testing.Log if
// testing.Verbose is false and Stderr otherwise.
func StreamOutput(ctx context.Context, t *testing.T, out KfTestOutput) {
	lines := CombineOutput(ctx, t, out)

	for {
		select {
		case <-ctx.Done():
			return
		case <-out.Done:
			return
		case line, ok := <-lines:
			if ok {
				Logf(t, line)
			}
		}
	}
}

// RetryPost will post until successful, duration has been reached or context is
// done. A close function is returned for closing the sub-context.
func RetryPost(
	ctx context.Context,
	t *testing.T,
	addr string,
	duration time.Duration,
	expectedStatusCode int,
	body string,
) (*http.Response, func()) {
	ctx, cancel := context.WithTimeout(ctx, duration)

	for {
		select {
		case <-ctx.Done():
			cancel()
			t.Fatalf("context cancelled")
		default:
		}

		req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(body))
		if err != nil {
			cancel()
			t.Fatalf("failed to create request: %s", err)
		}
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			Logf(t, "failed to post (retrying...): %s", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		if resp.StatusCode != expectedStatusCode {
			Logf(t, "got %d, wanted %d (retrying...)", resp.StatusCode, expectedStatusCode)
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		return resp, func() {
			cancel()
			resp.Body.Close()
		}
	}
}

// Kf provides a DSL for running integration tests.
type Kf struct {
	t  *testing.T
	kf KfInvoker
}

// KF returns a kf.
func KF(t *testing.T, kf KfInvoker) *Kf {
	return &Kf{
		t:  t,
		kf: kf,
	}
}

// CreateQuota creates a resourcequota.
func (k *Kf) CreateQuota(ctx context.Context, quotaName string, extraArgs ...string) ([]string, error) {
	k.t.Helper()
	Logf(k.t, "creating quota %q...", quotaName)
	defer Logf(k.t, "done creating quota %q.", quotaName)

	args := []string{
		"create-quota",
		"--namespace", SpaceFromContext(ctx),
		quotaName,
	}

	output, err := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutputStr(ctx, k.t, output), <-err
}

// Quotas returns all the quotas from `kf quotas`
func (k *Kf) Quotas(ctx context.Context) ([]string, error) {
	k.t.Helper()
	Logf(k.t, "listing quotas...")
	defer Logf(k.t, "done listing quotas.")
	output, err := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"quotas",
			"--namespace", SpaceFromContext(ctx),
		},
	})

	return CombineOutputStr(ctx, k.t, output), <-err
}

// DeleteQuota deletes a quota.
func (k *Kf) DeleteQuota(ctx context.Context, quotaName string) ([]string, error) {
	k.t.Helper()
	Logf(k.t, "deleting %q...", quotaName)
	defer Logf(k.t, "done deleting %q.", quotaName)
	output, err := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"delete-quota",
			"--namespace", SpaceFromContext(ctx),
			quotaName,
		},
	})
	return CombineOutputStr(ctx, k.t, output), <-err
}

// GetQuota returns information about a quota.
func (k *Kf) GetQuota(ctx context.Context, quotaName string) ([]string, error) {
	k.t.Helper()
	Logf(k.t, "getting %q...", quotaName)
	defer Logf(k.t, "done getting %q.", quotaName)
	output, err := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"quota",
			"--namespace", SpaceFromContext(ctx),
			quotaName,
		},
	})

	return CombineOutputStr(ctx, k.t, output), <-err
}

// UpdateQuota updates a quota.
func (k *Kf) UpdateQuota(ctx context.Context, quotaName string, extraArgs ...string) ([]string, error) {
	k.t.Helper()
	Logf(k.t, "updating %q...", quotaName)
	defer Logf(k.t, "done updating %q.", quotaName)

	args := []string{
		"update-quota",
		"--namespace", SpaceFromContext(ctx),
		quotaName,
	}

	output, err := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})

	return CombineOutputStr(ctx, k.t, output), <-err
}

// Push pushes an application.
func (k *Kf) Push(ctx context.Context, appName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "pushing app %q...", appName)
	defer Logf(k.t, "done pushing app %q.", appName)

	args := []string{
		"push",
		"--namespace", SpaceFromContext(ctx),
		appName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("push %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Logs displays the logs of an application.
func (k *Kf) Logs(ctx context.Context, appName string, extraArgs ...string) <-chan string {
	k.t.Helper()
	Logf(k.t, "displaying logs of app %q...", appName)
	defer Logf(k.t, "done displaying logs of app %q.", appName)

	args := []string{
		"logs",
		"--namespace", SpaceFromContext(ctx),
		appName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("logs %q", appName), errs)
	return CombineOutput(ctx, k.t, output)
}

// AppInfo is the information returned by listing an app. It is returned by
// ListApp.
type AppInfo struct {
	Name           string
	RequestedState string
	Instances      string
	Memory         string
	Disk           string
	URLs           []string
}

// Apps returns all the apps from `kf app`
func (k *Kf) Apps(ctx context.Context) map[string]AppInfo {
	Logf(k.t, "listing apps...")
	defer Logf(k.t, "done listing apps.")
	k.t.Helper()
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"apps",
			"--namespace", SpaceFromContext(ctx),
		},
	})
	PanicOnError(ctx, k.t, "apps", errs)
	apps := CombineOutputStr(ctx, k.t, output)

	if len(apps) == 0 {
		return nil
	}

	results := map[string]AppInfo{}
	for _, app := range apps[1:] {
		f := strings.Fields(app)
		if len(f) <= 1 {
			continue
		}

		name := f[0]
		requestedState := f[1]
		instances := f[2]
		memory := f[3]
		disk := f[4]
		var urls []string
		if len(f) > 5 {
			urls = strings.Split(f[5], ",")
		}

		results[name] = AppInfo{
			Name:           name,
			RequestedState: requestedState,
			Instances:      instances,
			Memory:         memory,
			Disk:           disk,
			URLs:           urls,
		}
	}

	return results
}

// Delete deletes an application.
func (k *Kf) Delete(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "deleting %q...", appName)
	defer Logf(k.t, "done deleting %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"delete",
			"--namespace", SpaceFromContext(ctx),
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("delete %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Stop stops an application.
func (k *Kf) Stop(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "stopping %q...", appName)
	defer Logf(k.t, "done stopping %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"stop",
			"--namespace", SpaceFromContext(ctx),
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("stop %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Start starts an application.
func (k *Kf) Start(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "starting %q...", appName)
	defer Logf(k.t, "done starting %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"start",
			"--namespace", SpaceFromContext(ctx),
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("start %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Restart restarts an application.
func (k *Kf) Restart(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "restarting %q...", appName)
	defer Logf(k.t, "done restarting %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"restart",
			"--namespace", SpaceFromContext(ctx),
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("restart %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Proxy starts a proxy for an application.
func (k *Kf) Proxy(ctx context.Context, appName string, port int) {
	k.t.Helper()
	Logf(k.t, "running proxy for %q...", appName)
	defer Logf(k.t, "done running proxy for %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"proxy",
			"--namespace", SpaceFromContext(ctx),
			appName,
			fmt.Sprintf("--port=%d", port),
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("proxy %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// SetEnv sets the environment variable for an app.
func (k *Kf) SetEnv(ctx context.Context, appName, name, value string) {
	k.t.Helper()
	Logf(k.t, "running set-env %s=%s for %q...", name, value, appName)
	defer Logf(k.t, "done running set-env %s=%s for %q.", name, value, appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"set-env",
			"--namespace", SpaceFromContext(ctx),
			appName,
			name,
			value,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("set-env %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// UnsetEnv unsets the environment variable for an app.
func (k *Kf) UnsetEnv(ctx context.Context, appName, name string) {
	k.t.Helper()
	Logf(k.t, "running unset-env %s for %q...", name, appName)
	defer Logf(k.t, "done running unset-env %s for %q.", name, appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"unset-env",
			"--namespace", SpaceFromContext(ctx),
			appName,
			name,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("unset-env %q", appName), errs)
	StreamOutput(ctx, k.t, output)
}

// Env displays the environment variables for an app.
func (k *Kf) Env(ctx context.Context, appName string) map[string]string {
	k.t.Helper()
	Logf(k.t, "running env for %q...", appName)
	defer Logf(k.t, "done running env for %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"env",
			"--namespace", SpaceFromContext(ctx),
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("env %q", appName), errs)
	envs := CombineOutputStr(ctx, k.t, output)

	if len(envs) == 0 {
		return nil
	}

	results := map[string]string{}
	for _, line := range envs[1:] {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// Remove the ':' suffix
		key := fields[0]
		key = key[:len(key)-1]
		results[key] = fields[1]
	}

	return results
}

// Doctor runs the doctor command.
func (k *Kf) Doctor(ctx context.Context) {
	k.t.Helper()
	Logf(k.t, "running doctor...")
	defer Logf(k.t, "done running doctor.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"doctor",
		},
	})
	PanicOnError(ctx, k.t, "doctor", errs)
	StreamOutput(ctx, k.t, output)
}

// Buildpacks runs the buildpacks command.
func (k *Kf) Buildpacks(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running buildpacks...")
	defer Logf(k.t, "done running buildpacks.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"buildpacks",
			"--namespace", SpaceFromContext(ctx),
		},
	})
	PanicOnError(ctx, k.t, "buildpacks", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// Stacks runs the stacks command.
func (k *Kf) Stacks(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running stacks...")
	defer Logf(k.t, "done running stacks.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"stacks",
			"--namespace", SpaceFromContext(ctx),
		},
	})
	PanicOnError(ctx, k.t, "stacks", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// CreateRoute runs the create-route command.
func (k *Kf) CreateRoute(ctx context.Context, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-route...")
	defer Logf(k.t, "done running create-route.")

	args := []string{
		"create-route",
		"--namespace", SpaceFromContext(ctx),
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "create-route", errs)
	StreamOutput(ctx, k.t, output)
}

// DeleteRoute runs the delete-route command.
func (k *Kf) DeleteRoute(ctx context.Context, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-route...")
	defer Logf(k.t, "done running delete-route.")

	args := []string{
		"delete-route",
		"--namespace", SpaceFromContext(ctx),
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "delete-route", errs)
	StreamOutput(ctx, k.t, output)
}

// Routes runs the routes command.
func (k *Kf) Routes(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running routes...")
	defer Logf(k.t, "done running routes.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"routes",
			"--namespace", SpaceFromContext(ctx),
		},
	})
	PanicOnError(ctx, k.t, "routes", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// CreateSpace runs the create-space command.
func (k *Kf) CreateSpace(ctx context.Context, space string) []string {
	k.t.Helper()
	Logf(k.t, "running create-space...")
	defer Logf(k.t, "done running create-space.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"create-space",
			"--container-registry", DockerRegistry(),
			"--domain", IntegrationTestDomain,
			space,
		},
	})
	PanicOnError(ctx, k.t, "create-space", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// DeleteSpace runs the create-space command.
func (k *Kf) DeleteSpace(ctx context.Context, space string) []string {
	k.t.Helper()
	Logf(k.t, "running delete-space...")
	defer Logf(k.t, "done running delete-space.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"delete-space",
			space,
		},
	})
	PanicOnError(ctx, k.t, "delete-space", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// Spaces returns all the spaces from `kf spaces`
func (k *Kf) Spaces(ctx context.Context) []string {
	Logf(k.t, "listing spaces...")
	defer Logf(k.t, "done listing spaces.")
	k.t.Helper()
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"spaces",
		},
	})
	PanicOnError(ctx, k.t, "spaces", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// Target runs the target command.
func (k *Kf) Target(ctx context.Context, namespace string) []string {
	k.t.Helper()
	Logf(k.t, "running target...")
	defer Logf(k.t, "done running target.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"target",
			"--space",
			namespace,
		},
	})
	PanicOnError(ctx, k.t, "target", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// Marketplace runs the marketplace command and returns the output.
func (k *Kf) Marketplace(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running marketplace...")
	defer Logf(k.t, "done running marketplace.")

	args := []string{
		"marketplace",
		"--namespace", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "marketplace", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// CreateServiceBroker runs the create-service-broker command.
func (k *Kf) CreateServiceBroker(ctx context.Context, brokerName string, url string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-service-broker...")
	defer Logf(k.t, "done running create-service-broker.")

	args := []string{
		"create-service-broker",
		"--namespace", SpaceFromContext(ctx),
		brokerName,
		url,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "create-service-broker", errs)
	StreamOutput(ctx, k.t, output)
}

// DeleteServiceBroker runs the delete-service-broker command.
func (k *Kf) DeleteServiceBroker(ctx context.Context, brokerName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-service-broker...")
	defer Logf(k.t, "done running delete-service-broker.")

	args := []string{
		"delete-service-broker",
		"--namespace", SpaceFromContext(ctx),
		brokerName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "delete-service-broker", errs)
	StreamOutput(ctx, k.t, output)
}

// CreateService runs the create-service command.
func (k *Kf) CreateService(ctx context.Context, serviceClass string, servicePlan string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-service...")
	defer Logf(k.t, "done running create-service.")

	args := []string{
		"create-service",
		"--namespace", SpaceFromContext(ctx),
		serviceClass,
		servicePlan,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "create-service", errs)
	StreamOutput(ctx, k.t, output)
}

// Services runs the services command.
func (k *Kf) Services(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running services...")
	defer Logf(k.t, "done running services.")

	args := []string{
		"services",
		"--namespace", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "services", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// DeleteService runs the delete-service command.
func (k *Kf) DeleteService(ctx context.Context, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-service...")
	defer Logf(k.t, "done running delete-service.")

	args := []string{
		"delete-service",
		"--namespace", SpaceFromContext(ctx),
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "delete-service", errs)
	StreamOutput(ctx, k.t, output)
}

// BindService runs the bind-service command.
func (k *Kf) BindService(ctx context.Context, appName string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running bind-service...")
	defer Logf(k.t, "done running bind-service.")

	args := []string{
		"bind-service",
		"--namespace", SpaceFromContext(ctx),
		appName,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "bind-service", errs)
	StreamOutput(ctx, k.t, output)
}

// Bindings runs the services command.
func (k *Kf) Bindings(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running bindings...")
	defer Logf(k.t, "done running bindings.")

	args := []string{
		"bindings",
		"--namespace", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "bindings", errs)
	return CombineOutputStr(ctx, k.t, output)
}

// UnbindService runs the unbind-service command.
func (k *Kf) UnbindService(ctx context.Context, appName string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running unbind-service...")
	defer Logf(k.t, "done running unbind-service.")

	args := []string{
		"unbind-service",
		"--namespace", SpaceFromContext(ctx),
		appName,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "unbind-service", errs)
	StreamOutput(ctx, k.t, output)
}

// VcapServices runs the vcap-services command.
func (k *Kf) VcapServices(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running vcap-services...")
	defer Logf(k.t, "done running vcap-services.")

	args := []string{
		"vcap-services",
		"--namespace", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	PanicOnError(ctx, k.t, "vcap-services", errs)
	return CombineOutputStr(ctx, k.t, output)
}

type spaceKey struct{}

// ContextWithSpace returns a context that has the space information. The
// space can be fetched via SpaceFromContext.
func ContextWithSpace(ctx context.Context, space string) context.Context {
	return context.WithValue(ctx, spaceKey{}, space)
}

// SpaceFromContext returns the space name given a context that has been setup
// via ContextWithSpace.
func SpaceFromContext(ctx context.Context) string {
	return ctx.Value(spaceKey{}).(string)
}

type brokerKey struct{}

// ContextWithBroker returns a context that has the service broker information. The
// broker name can be fetched via BrokerFromContext.
func ContextWithBroker(ctx context.Context, brokerName string) context.Context {
	return context.WithValue(ctx, brokerKey{}, brokerName)
}

// BrokerFromContext returns the broker name given a context that has been setup
// via ContextWithBroker.
func BrokerFromContext(ctx context.Context) string {
	return ctx.Value(brokerKey{}).(string)
}

type serviceInstanceKey struct{}

// ContextWithServiceInstance returns a context that has the service instance information. The
// service instance name can be fetched via ServiceInstanceFromContext.
func ContextWithServiceInstance(ctx context.Context, serviceInstanceName string) context.Context {
	return context.WithValue(ctx, serviceInstanceKey{}, serviceInstanceName)
}

// ServiceInstanceFromContext returns the service instance name given a context that has been setup
// via ContextWithServiceInstance.
func ServiceInstanceFromContext(ctx context.Context) string {
	return ctx.Value(serviceInstanceKey{}).(string)
}

type serviceClassKey struct{}

// ContextWithServiceClass returns a context that has the service class information. The
// service class name can be fetched via ServiceClassFromContext.
func ContextWithServiceClass(ctx context.Context, serviceClassName string) context.Context {
	return context.WithValue(ctx, serviceClassKey{}, serviceClassName)
}

// ServiceClassFromContext returns the service class name given a context that has been setup
// via ContextWithServiceClass.
func ServiceClassFromContext(ctx context.Context) string {
	return ctx.Value(serviceClassKey{}).(string)
}

type servicePlanKey struct{}

// ContextWithServicePlan returns a context that has the service plan information. The
// service plan name can be fetched via ServiceInstanceFromContext.
func ContextWithServicePlan(ctx context.Context, servicePlanName string) context.Context {
	return context.WithValue(ctx, servicePlanKey{}, servicePlanName)
}

// ServicePlanFromContext returns the service plan name given a context that has been setup
// via ContextWithServicePlan.
func ServicePlanFromContext(ctx context.Context) string {
	return ctx.Value(servicePlanKey{}).(string)
}

type appKey struct{}

// ContextWithApp returns a context that has the app information. The
// app name can be fetched via AppFromContext.
func ContextWithApp(ctx context.Context, appName string) context.Context {
	return context.WithValue(ctx, appKey{}, appName)
}

// AppFromContext returns the service plan name given a context that has been setup
// via ContextWithApp.
func AppFromContext(ctx context.Context) string {
	return ctx.Value(appKey{}).(string)
}

// ExpectedAddr returns the expected address for integration tests given a
// hostname and URL path.
func ExpectedAddr(hostname, urlPath string) string {
	hostnameDomain := IntegrationTestDomain
	if hostname != "" {
		hostnameDomain = fmt.Sprintf("%s.%s", hostname, IntegrationTestDomain)
	}

	return hostnameDomain + path.Join("/", urlPath)
}
