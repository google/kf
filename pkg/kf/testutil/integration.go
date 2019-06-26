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
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// EnvGcpProjectID is the environment variable used to store the GCP
	// project ID for the integration tests. If this is not set, then the
	// integration tests are skipped.
	EnvGcpProjectID = "GCP_PROJECT_ID"
)

// GCPProjectID returns the configured GCP Project ID.
func GCPProjectID() string {
	return os.Getenv(EnvGcpProjectID)
}

// RunIntegrationTest skips the tests if testing.Short() is true (via --short
// flag) or if GCP_PROJECT_ID is not set. Otherwise it runs the given test.
func RunIntegrationTest(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	t.Helper()
	if testing.Short() {
		t.Skip()
	}

	projID := os.Getenv(EnvGcpProjectID)
	if projID == "" {
		t.Skipf("%s is required for integration tests... Skipping...", EnvGcpProjectID)
	}

	// Setup context that will allow us to cleanup if the user wants to
	// cancel the tests.
	ctx, cancel := context.WithCancel(context.Background())

	// Give everything time to clean up.
	defer time.Sleep(time.Second)
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		t.Log("Signal received... Cleaning up... (Hit Ctrl-C again to quit immediately)")
		cancel()
		<-c
		os.Exit(1)
	}()

	test(ctx, t)
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
	RunIntegrationTest(t, func(ctx context.Context, t *testing.T) {
		t.Helper()
		kfPath := CompileKf(ctx, t)

		kf := KF(t, func(ctx context.Context, t *testing.T, cfg KfTestConfig) (KfTestOutput, <-chan error) {
			return kf(ctx, t, kfPath, cfg)
		})

		test(ctx, t, kf)
	})
}

// CompileKf compiles the `kf` binary. It returns a string to the resulting
// binary.
func CompileKf(ctx context.Context, t *testing.T) string {
	t.Helper()
	return Compile(ctx, t, "./cmd/kf")
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
					panic(fmt.Sprintf("%s: %s", messagePrefix, err))
				}
			}
		}
	}()
}

// CombinedOutputStr writes the Stdout and Stderr string. It will return when
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

// CombinedOutput writes the Stdout and Stderr to a channel for each line.
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
	if testing.Verbose() {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(format, i...))
		return
	}

	t.Logf(format, i...)
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
func RetryPost(ctx context.Context, t *testing.T, addr string, duration time.Duration, body io.Reader) (*http.Response, func()) {
	ctx, cancel := context.WithTimeout(ctx, duration)

	for {
		select {
		case <-ctx.Done():
			cancel()
			t.Fatalf("context cancelled")
		default:
		}

		req, err := http.NewRequest(http.MethodPost, addr, body)
		if err != nil {
			cancel()
			t.Fatalf("failed to create request: %s", err)
		}
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			Logf(t, "failed to post (retrying...): %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return resp, func() {
			cancel()
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
		Args: []string{"quotas"},
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
	Name   string
	Domain string
	Reason string
}

// Apps returns all the apps from `kf app`
func (k *Kf) Apps(ctx context.Context) map[string]AppInfo {
	Logf(k.t, "listing apps...")
	defer Logf(k.t, "done listing apps.")
	k.t.Helper()
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{"apps"},
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

		domain := f[1]
		if !strings.HasPrefix(domain, "http") {
			domain = "http://" + domain
		}

		var reason string
		if len(f) > 5 {
			reason = f[5]
		}

		results[f[0]] = AppInfo{
			Name:   f[0],
			Domain: domain,
			Reason: reason,
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
			appName,
		},
	})
	PanicOnError(ctx, k.t, fmt.Sprintf("delete %q", appName), errs)
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
		results[fields[0]] = fields[1]
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
		},
	})
	PanicOnError(ctx, k.t, "routes", errs)
	return CombineOutputStr(ctx, k.t, output)
}
