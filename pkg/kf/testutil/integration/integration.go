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

package integration

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
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

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/injection"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	// EnvPreCompiledKfCLI is the environment variable used to store the
	// location of a pre-compiled Kf CLI. If it's not set, then the tests will
	// compile the CLI.
	EnvPreCompiledKfCLI = "KF_CLI_PATH"

	// InternalDefaultDomain is the apps.internal domain which is configured by default for all spaces.
	InternalDefaultDomain = "apps.internal"

	// EgressMetricsEnvVar is the environment variable that when set to
	// 'true', will write benchmarking results to Stackdriver.
	EgressMetricsEnvVar = "EGRESS_TEST_RESULTS"

	// IntegrationTestID sets the test_id label on metrics. If not provided, a
	// random id is selected.
	IntegrationTestID = "INTEGRATION_TEST_ID"

	// CustomMetricLabels is a comma separated list of labels the user wants
	// included with each metric that is written.  The format is
	// key1=value1,key2=value2
	//
	// Example:
	//
	// CUSTOM_METRIC_LABELS="foo=a,bar=b,baz=c"
	CustomMetricLabels = "CUSTOM_METRIC_LABELS"

	// KeepSpaces will prevent the spaces from being cleaned up.
	KeepSpaces = "KEEP_SPACES"

	// RandomSpaceNameEnvVar will create random space names instead of
	// deterministic ones.
	RandomSpaceNameEnvVar = "RANDOM_SPACE_NAMES"

	// SkipDoctorEnvVar when set to true, will NOT run kf doctor before tests.
	SkipDoctorEnvVar = "SKIP_DOCTOR"

	// SpaceDomainEnvVar is used to override the space domain used at the integration test.
	// The default space domain is integration-tests.kf.dev, but it needs to be overridden
	// to be nip.io in BareMetal environment for tests to pass, see b/195313679.
	SpaceDomainEnvVar = "SPACE_DOMAIN"

	// DefaultSpaceDomain is the default space domain used for integration
	// tests. It is overriden by SpaceDomainEnvVar if set.
	DefaultSpaceDomain = "integration-tests.kf.dev"

	// IsBareMetalEnvVar when set to true, indicates that tests are running at a BareMetal environment.
	IsBareMetalEnvVar = "BARE_METAL"

	nfsServiceClassName = "nfs"

	nfsServicePlanExisting = "existing"
)

// ShouldSkipIntegration returns true if integration tests are being skipped.
func ShouldSkipIntegration(t *testing.T) bool {
	t.Helper()

	if skipIntegration := os.Getenv("SKIP_INTEGRATION"); skipIntegration == "true" {
		t.Skipf("Skipping %s because SKIP_INTEGRATION was true", t.Name())
		return true
	}

	if !strings.HasPrefix(t.Name(), "TestIntegration_") {
		// We want to enforce a convention so scripts can single out
		// integration tests.
		t.Fatalf("Integration tests must have the name format of 'TestIntegration_XXX`")
		return true
	}

	if testing.Short() {
		t.Skipf("Skipping %s because short tests were requested", t.Name())
		return true
	}

	return false
}

func state(t *testing.T) string {
	if t.Failed() {
		return "FAILED"
	}
	return "PASSED"
}

func logTestResults(ctx context.Context, t *testing.T, name string, f func()) {
	start := time.Now()
	defer func() {
		if os.Getenv(EgressMetricsEnvVar) == "true" {
			WriteMetric(ctx, Int64Gauge{
				Name:  name,
				Units: "ns",
				Value: int64(time.Since(start)),
				Labels: MetricLabels{
					"state": state(t),
				},
			})
		}
		Logf(t, "Test %s took %v and %s", t.Name(), time.Since(start), state(t))
	}()

	f()
}

func withSignalCaptureCancel(ctx context.Context, t *testing.T, f func(ctx context.Context)) {
	// Setup context that will allow us to cleanup if the user wants to
	// cancel the tests.
	ctx, cancel := context.WithCancel(ctx)

	// Give everything time to clean up.
	defer time.Sleep(time.Second)
	defer cancel()
	CancelOnSignal(ctx, cancel, t)

	f(ctx)
}

func withLabeledContext(ctx context.Context, t *testing.T, f func(ctx context.Context)) {
	// Parse any extra metrics the user wants written with the metrics.
	labels := MetricLabels{
		"integration_test": strings.TrimLeft(t.Name(), "TestIntegration_"),
		"test_id":          TestID(),
	}

	if customLabels := os.Getenv(CustomMetricLabels); customLabels != "" {
		for _, keyValue := range strings.Split(customLabels, ",") {
			s := strings.SplitN(keyValue, "=", 2)
			if len(s) < 2 {
				t.Fatalf("%q is malformed", keyValue)
			}
			labels[s[0]] = s[1]
		}
	}

	f(ContextWithMetricLabels(ctx, labels))
}

// WithWontFailContext returns a context that tells the kf function to NOT
// fail on an error. This is useful during cleanup operationos (e.g., deleting
// Spaces) that might not finish before the test process closes out (which
// would cause spurious errors).
func WithWontFailContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, wontFailKey{}, true)
}

func wontFailFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(wontFailKey{}).(bool)
	return v
}

type wontFailKey struct{}

// RunIntegrationTest skips the tests if testing.Short() is true (via --short
// flag). Otherwise it runs the given test.
func RunIntegrationTest(
	ctx context.Context,
	t *testing.T,
	test func(ctx context.Context, t *testing.T),
) {
	t.Helper()

	if ShouldSkipIntegration(t) {
		return
	}

	withSignalCaptureCancel(ctx, t, func(ctx context.Context) {
		// Only add a deadline if one isn't already set.
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
		}
		ctx = ContextWithStackdriverOutput(ctx)
		withLabeledContext(ctx, t, func(ctx context.Context) {
			test(ctx, t)
		})
	})
}

// RunKubeAPITest is for executing tests against the Kubernetes API directly.
func RunKubeAPITest(
	ctx context.Context,
	t *testing.T,
	test func(ctx context.Context, t *testing.T),
) {
	t.Helper()

	if ShouldSkipIntegration(t) {
		return
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	restCfg, err := clientCfg.ClientConfig()
	if err != nil {
		t.Fatalf("couldn't fetch kubeconfig: %v", err)
	}

	withSignalCaptureCancel(ctx, t, func(ctx context.Context) {
		ctx = ContextWithStackdriverOutput(ctx)
		withLabeledContext(ctx, t, func(ctx context.Context) {
			logTestResults(ctx, t, "kf_integration_test_duration", func() {
				ctx = contextWithRestConfig(ctx, restCfg)

				// Set up the autowired informers so they can be used in testing
				ctx, _ = injection.Default.SetupInformers(ctx, restCfg)

				test(ctx, t)
			})
		})
	})
}

type restConfigKey struct{}

func contextWithRestConfig(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, restConfigKey{}, cfg)
}

func restConfigFromContext(ctx context.Context) *rest.Config {
	return ctx.Value(restConfigKey{}).(*rest.Config)
}

// WithRestConfig gets the rest config from the context and passes it to the callback.
func WithRestConfig(ctx context.Context, t *testing.T, callback func(cfg *rest.Config)) {
	t.Helper()

	callback(restConfigFromContext(ctx))
}

// WithKubernetes creates a Kubernetes client from the config on the context
// and passes it to the callback.
func WithKubernetes(ctx context.Context, t *testing.T, callback func(k8s *kubernetes.Clientset)) {
	t.Helper()

	WithRestConfig(ctx, t, func(cfg *rest.Config) {
		k8s, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("Creating Kubernetes: %v", err)
		}

		callback(k8s)
	})
}

// WithDynamicClient creates a dynamic Kubernetes client from the config on the
// context and passes it to the callback.
func WithDynamicClient(ctx context.Context, t *testing.T, callback func(client dynamic.Interface)) {
	t.Helper()

	WithRestConfig(ctx, t, func(cfg *rest.Config) {
		dynamic, err := dynamic.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("Creating dynamic client: %v", err)
		}

		callback(dynamic)
	})
}

// WithSpace creates a space and deletes it after the test is done.  It can be
// used within the context of a RunKubeAPITest.
func WithSpace(ctx context.Context, t *testing.T, callback func(namespace string)) {
	t.Helper()

	WithKubernetes(ctx, t, func(k8s *kubernetes.Clientset) {
		name := fmt.Sprintf("integration-%d", time.Now().UnixNano())
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}

		if _, err := k8s.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{}); err != nil {
			t.Fatalf("Creating namespace: %v", err)
		}
		defer k8s.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})

		Logf(t, "With Namespace: %s", name)
		callback(name)
	})
}

// WithNodeLabel adds a random value for the given LabelName on the 1st node in the pool.
func WithNodeLabel(ctx context.Context, labelName string, t *testing.T, callback func(labelValue string, nodeName string, k8s *kubernetes.Clientset)) {
	t.Helper()
	WithKubernetes(ctx, t, func(k8s *kubernetes.Clientset) {
		var nodeName string
		labelValue := fmt.Sprintf("value-%d", time.Now().UnixNano())
		// Should exclude master nodes, see https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#control-plane-node-isolation
		nodes, err := k8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{
			LabelSelector: "!node-role.kubernetes.io/master",
		})
		if err != nil {
			t.Fatalf("Error getting nodes: %v", err)
		}

		// Add the custom Label to 1st node in the list
		if len(nodes.Items) > 0 {
			node := nodes.Items[0]
			nodeName = node.Name
			labels := node.Labels
			labels[labelName] = labelValue
			t.Logf("Adding a Label %s=%s on the node %s", labelName, labelValue, node.Name)
			if _, err := k8s.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{}); err != nil {
				t.Fatalf("Error updating the node: %v", err)
			}
		} else {
			t.Fatalf("No nodes present in the cluster")
		}
		callback(labelValue, nodeName, k8s)

	})

}

// WithApp creates an App and deletes it after the test is done.
func WithApp(ctx context.Context, t *testing.T, kf *Kf, appName string, path string, isBroker bool, callback func(newCtx context.Context)) {
	WithAppArgs(ctx, t, kf, appName, path, isBroker, nil, callback)
}

// WithTaskApp creates an App for running task and deletes it after the test
// is done.
func WithTaskApp(ctx context.Context, t *testing.T, kf *Kf, appName string, path string, isBroker bool, callback func(newCtx context.Context)) {
	WithAppArgs(ctx, t, kf, appName, path, isBroker, []string{"--task"}, callback)
}

// WithAppArgs creates an App with provided args and deletes it after the test
// is done.
func WithAppArgs(ctx context.Context, t *testing.T, kf *Kf, appName string, path string, isBroker bool, args []string, callback func(newCtx context.Context)) {
	// Push the app then clean it up.
	kf.CachePush(ctx, appName, filepath.Join(RootDir(ctx, t), path), args...)

	if !isBroker {
		ctx = ContextWithApp(ctx, appName)
	}

	callback(ctx)
}

// CancelOnSignal watches for a kill signal. If it gets one, it invokes the
// cancel function. An aditional signal will exit the process. If the given context finishes, the underlying go-routine
// finishes.
func CancelOnSignal(ctx context.Context, cancel func(), t *testing.T) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		select {
		case <-c:
			Logf(t, "Signal received... Cleaning up... (Hit Ctrl-C again to quit immediately)")
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
	Dir  string
}

// KfTestOutput is the output from `kf`. Note, this output is while kf is
// running.
type KfTestOutput struct {
	Stdout io.Reader
	Stderr io.Reader
	Stdin  io.Writer
	Done   <-chan struct{}
}

// Kf provides a DSL for running integration tests.
type Kf struct {
	t          *testing.T
	binaryPath string
}

// NewKf creates a Kf for running tests with.
func NewKf(t *testing.T, binaryPath string) *Kf {
	return &Kf{
		t:          t,
		binaryPath: binaryPath,
	}
}

// logExecutionTime logs the start of the execution of a command along with the
// remaining deadline. Callers should call the returned func when command
// execution is complete to log the total time taken.
func logExecutionTime(ctx context.Context, t *testing.T, cfg KfTestConfig) func() {
	start := time.Now()

	if deadline, ok := ctx.Deadline(); ok {
		Logf(t, "start %s (timeout=%v)", strings.Join(cfg.Args, " "), deadline.Sub(time.Now()))
	} else {
		Logf(t, "start %s (timeout=None)", strings.Join(cfg.Args, " "))
	}

	return func() {
		Logf(t, "end %s (took=%v)", strings.Join(cfg.Args, " "), time.Since(start))
	}
}

func (k *Kf) kf(ctx context.Context, t *testing.T, cfg KfTestConfig) (out KfTestOutput, outErrs <-chan error) {
	t.Helper()
	start := time.Now()

	logDone := logExecutionTime(ctx, t, cfg)

	defer func() {
		// We have to run this on a go-routine because this call is async and
		// it likely won't be done when the function exits.
		go func() {
			<-out.Done
			logDone()
		}()
	}()

	// There are some operations that have no reason to cause the tests to
	// fail (e.g., cleaning up Spaces). Occasionally, when a test ends
	// (especially the last one) and the test is GC'ing its artifacts (e.g.,
	// Spaces, Apps, etc...) the process will close and cause these operations
	// to return an error. There is no reason this should make the test go
	// red.
	defer func() {
		wontFail := wontFailFromContext(ctx)
		if wontFail {
			err := recover()
			if err == nil {
				return
			}

			t.Logf("Test failed, however it is marked as WontFail: %v", err)

			// We have to fill in the output so downstream consumers don't
			// fail.
			out.Stdout = bytes.NewReader(nil)
			out.Stderr = bytes.NewReader(nil)

			done, cancel := context.WithCancel(context.Background())
			cancel()
			out.Done = done.Done()
		}
	}()
	fatalf := func(format string, args ...interface{}) {
		if wontFailFromContext(ctx) {
			// Use a panic so we stop execution.
			panic(fmt.Sprintf(format, args...))
		}

		t.Fatalf(format, args...)
	}

	Logf(t, "kf %s\n", strings.Join(cfg.Args, " "))

	cmd := exec.CommandContext(ctx, k.binaryPath, cfg.Args...)
	for name, value := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
	}
	cmd.Dir = cfg.Dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fatalf("failed to fetch Stdout pipe: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fatalf("failed to fetch Stderr pipe: %s", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fatalf("failed to fetch Stdin pipe: %s", err)
	}

	// According to the docs, you shouldn't invoke cmd.Wait until the STDOUT
	// and STDERR have returned EOFs. So we're going to monitor them here so
	// that we can properly mark this as done and invoke Wait.
	var wg sync.WaitGroup
	wg.Add(2)

	consumePipe := func(w io.WriteCloser, r io.Reader) {
		io.Copy(w, r)
		w.Close()
		wg.Done()
	}

	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()
	go consumePipe(stdoutW, stdout)
	go consumePipe(stderrW, stderr)

	if err := cmd.Start(); err != nil {
		fatalf("failed to start kf: %s", err)
	}

	done := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			errs <- err
		}
		close(done)

		// Measure how long he command took.
		command := cfg.Args[0]
		WriteMetric(ctx, Int64Gauge{
			Name:  "kf_command_duration",
			Units: "ns",
			Value: int64(time.Since(start)),
			Labels: MetricLabels{
				"command": command,
			},
		})
	}()

	return KfTestOutput{
		Stdout: stdoutR,
		Stderr: stderrR,
		Stdin:  stdin,
		Done:   done,
	}, errs
}

// RunSynchronous runs kf with the provided configuration and returns the
// results.
func (k *Kf) RunSynchronous(ctx context.Context, cfg KfTestConfig) (stdout, stderr []byte, err error) {
	k.t.Helper()

	defer logExecutionTime(ctx, k.t, cfg)()

	// Measure how long he command took.
	start := time.Now()
	defer func() {
		command := cfg.Args[0]
		WriteMetric(ctx, Int64Gauge{
			Name:  "kf_command_duration",
			Units: "ns",
			Value: int64(time.Since(start)),
			Labels: MetricLabels{
				"command": command,
			},
		})
	}()

	Logf(k.t, "kf %s\n", strings.Join(cfg.Args, " "))

	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.CommandContext(ctx, k.binaryPath, cfg.Args...)
	for name, value := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
	}
	cmd.Dir = cfg.Dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.Bytes()
	stderr = stderrBuf.Bytes()
	return
}

var checkClusterOnce sync.Once

// CheckCluster runs the Doctor command to ensure the K8s cluster is healthy.
// It will cache the results and therefore can be ran multiple times.
func CheckCluster(ctx context.Context, kf *Kf) {
	if os.Getenv(SkipDoctorEnvVar) == "true" {
		Logf(kf.t, "%q set to 'true', skipping cluster check...", SkipDoctorEnvVar)
		return
	}
	checkClusterOnce.Do(func() {
		kf.WaitForCluster(ctx)
	})
}

// KfTest is a test ran by RunKfTest.
type KfTest func(ctx context.Context, t *testing.T, kf *Kf)

// RunKfTest runs 'kf' for integration tests. It first compiles 'kf' and then
// launches it as a sub-process. It will set the args and environment
// variables accordingly. It will run the given test with the resulting
// STDOUT, STDERR and STDIN. It will cleanup the sub-process on completion via
// the context.
func RunKfTest(ctx context.Context, t *testing.T, test KfTest) {
	t.Helper()
	t.Parallel()

	if ShouldSkipIntegration(t) {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	CancelOnSignal(ctx, cancel, t)
	kfPath := CompileKf(ctx, t)

	RunIntegrationTest(ctx, t, func(ctx context.Context, t *testing.T) {
		t.Helper()

		kf := NewKf(t, kfPath)
		CheckCluster(ctx, kf)

		spaceName, _ := kf.CreateSpace(ctx)

		// Wait for Space to become ready
		testutil.AssertRetrySucceeds(ctx, t, -1, time.Second, func() error {
			if !kf.spaceExists(ctx, spaceName) {
				return fmt.Errorf("%s-> did not find Space %s", t.Name(), spaceName)
			}
			return nil
		})

		ctx = ContextWithSpace(ctx, spaceName)

		logTestResults(ctx, t, "kf_integration_test_duration", func() {
			test(ctx, t, kf)
		})
	})
}

var (
	compileKfOnce sync.Once
	compiledKf    string
)

// CompileKf compiles the `kf` binary or the path set by PRE_COMPILED_KF_CLI.
// It returns a string to the resulting binary.
func CompileKf(ctx context.Context, t *testing.T) string {
	t.Helper()

	// Check to see if CLI was provided by user. If so, use the given path.
	if path := os.Getenv(EnvPreCompiledKfCLI); path != "" {
		Logf(t, "%s provided. Using %s for Kf CLI path", EnvPreCompiledKfCLI, path)
		return path
	}

	compileKfOnce.Do(func() {
		Logf(t, "Compiling Kf CLI")
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

	tmpDir := CreateTempDir(ctx, t)

	outputPath := filepath.Join(tmpDir, "out")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, ".")
	cmd.Dir = codePath
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to go build (err=%s): %s", err, output)
	}

	return outputPath
}

// CreateTempDir returns a temp directory that will be cleaned up when the
// context finishes.
func CreateTempDir(ctx context.Context, t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	return tempDir
}

// RootDir uses git to find the root of the directory. This is useful for
// tests (especially integration ones that compile things).
func RootDir(ctx context.Context, t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "env", "GOMOD")
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

	return filepath.Dir(strings.TrimSpace(string(data)))
}

// CombineOutputStr writes the Stdout and Stderr string. It will return when
// ctx is done.
func CombineOutputStr(ctx context.Context, t *testing.T, out KfTestOutput, errs <-chan error) []string {
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
		case err := <-errs:
			if err != nil {
				if t.Failed() || wontFailFromContext(ctx) {
					Logf(t, "%v", err)
					return nil
				} else {
					t.Fatal(err)
				}
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
				Logf(t, s.Text())
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
func StreamOutput(ctx context.Context, t *testing.T, out KfTestOutput, errs <-chan error) {
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
		case err := <-errs:
			if err != nil {
				if t.Failed() || wontFailFromContext(ctx) {
					Logf(t, "%v", err)
					return
				} else {
					t.Fatal(err)
				}
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
	t.Helper()
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
		req.Header.Add("Content-Type", "text/plain")

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

// RetryGet will send a get request until successful, duration has been reached or context is
// done. A close function is returned for closing the sub-context.
func RetryGet(
	ctx context.Context,
	t *testing.T,
	addr string,
	duration time.Duration,
	expectedStatusCode int,
) (*http.Response, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, duration)

	for {
		select {
		case <-ctx.Done():
			cancel()
			t.Fatalf("context cancelled")
		default:
		}

		req, err := http.NewRequest(http.MethodGet, addr, nil)
		if err != nil {
			cancel()
			t.Fatalf("failed to create request: %s", err)
		}
		req = req.WithContext(ctx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			Logf(t, "failed to get (retrying...): %s", err)
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

func setFlagIfEmpty(key, value string, args []string) []string {
	// Create a temp FlagSet to parse the flags. Ignore any errors.
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.String(key, "", "")
	_ = fs.Parse(args)

	if fs.Changed(key) {
		// Flag already exists.
		return args
	}

	// Didn't find the flag, add it.
	return append(args, "--"+key, value)
}

// Push pushes an App.
func (k *Kf) Push(ctx context.Context, appName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "pushing App %q...", appName)
	defer Logf(k.t, "done pushing App %q.", appName)

	args := []string{
		"push",
		"--space", SpaceFromContext(ctx),
		appName,
	}

	// We aren't pushing anything too exciting, so lets use a small container.
	args = setFlagIfEmpty("cpu-cores", "100m", args)
	args = setFlagIfEmpty("disk-quota", "100M", args)
	args = setFlagIfEmpty("memory-limit", "100M", args)

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)

	k.t.Cleanup(func() {
		k.Delete(
			WithWontFailContext(DetachContext(ctx)),
			appName,
			"--async",
		)
	})
}

// SSH runs a shell on an App.
func (k *Kf) SSH(ctx context.Context, appName string, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "sshing to App %q...", appName)
	defer Logf(k.t, "done sshing to App %q.", appName)
	args := []string{
		"ssh",
		"--space", SpaceFromContext(ctx),
		appName,
	}
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// sourceCache is a mapping between directories and built container images.
// Only vanilla buildpack built images should be cached e.g. no builds that
// require specific environment variables or the like.
var sourceCache = newAppCache()

// CachePushV2 pushes an App with V2 buildpack and caches the source or uses a
// cached version. This is incompatible with additional arguments that might
// change the semantics of push. It returns true if the cache was used.
func (k *Kf) CachePushV2(ctx context.Context, appName, source string, args ...string) bool {
	return k.cachePush(ctx, appName, source, "cflinuxfs3", args...)
}

// CachePush pushes an App with V3 buildpack and caches the source or uses a
// cached version. This is incompatible with additional arguments that might
// change the semantics of push. It returns true if the cache was used.
func (k *Kf) CachePush(ctx context.Context, appName, source string, args ...string) bool {
	return k.cachePush(ctx, appName, source, "org.cloudfoundry.stacks.cflinuxfs3", args...)
}

func (k *Kf) cachePush(ctx context.Context, appName, source, stack string, args ...string) bool {
	k.t.Helper()
	cacheCtx, cacheCancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cacheCancel()
	containerImage, ok, err := sourceCache.Load(cacheCtx, source)
	if err != nil {
		Logf(k.t, "failed to load from cache: %v", err)
	}

	// If there's a cached image, re-use that cached built image rather than
	// building.
	if ok {
		Logf(k.t, "Using cached image %s instead of rebuilding %s", containerImage, source)
		args = append(args, "--docker-image", containerImage)
		k.Push(ctx, appName, args...)
		return true
	}

	// Otherwise, push the App and wait for it to become ready. Once ready, pull
	// the image off the App and store it in the cache.
	args = append(args, "--path", source, "--stack", stack)
	k.Push(ctx, appName, args...)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			Logf(k.t, "failed while waiting for ObservedGeneration to match")
			k.t.Fatal(ctx.Err())
		case <-ticker.C:
			appJSON := k.App(ctx, appName)
			obj := map[string]interface{}{}
			if err := json.Unmarshal([]byte(appJSON), &obj); err != nil {
				k.t.Fatal("unmarshaling App:", err)
			}

			// NOTE: unstructured is used here to avoid an import cycle.

			// Make sure metadata.generation and status.observedGeneration
			// match.
			// NOTE: metadata.generation and status.observedGeneration not
			// matching implies the reconciler is still actively "working" on
			// this CR. Therefore, the conditions should NOT be trusted yet.
			{
				generation, ok, _ := unstructured.NestedFloat64(obj, "metadata", "generation")
				if !ok {
					Logf(k.t, "failed to find metadata.generation...")
					continue
				}
				observedGeneration, ok, _ := unstructured.NestedFloat64(obj, "status", "observedGeneration")
				if !ok {
					Logf(k.t, "failed to find status.observedGeneration...")
					continue
				}
				if observedGeneration != generation {
					Logf(k.t, "observedGeneration and generation don't match...")
					continue
				}

				Logf(k.t, "observedGeneration and generation match")
			}

			containerImage, ok, _ = unstructured.NestedString(obj, "status", "image")
			if containerImage == "gcr.io/kf-releases/nop:nop" {
				// We're still grabbing the placeholder's container image. We
				// should wait and try again.
				Logf(k.t, "container image is still the placeholder one...")
				continue
			}

			if ok {
				cacheCtx, cacheCancel := context.WithTimeout(ctx, 500*time.Millisecond)
				defer cacheCancel()
				Logf(k.t, "Caching source %s as %s", source, containerImage)
				if err := sourceCache.Store(cacheCtx, source, containerImage); err != nil {
					Logf(k.t, "failed to store in cache: %v", err)
				}
			}

			// Getting this far means the loop should be done.
			return false
		}
	}
	return false
}

// Logs displays the logs of an App.
func (k *Kf) Logs(ctx context.Context, appName string, extraArgs ...string) (<-chan string, <-chan error) {
	k.t.Helper()
	Logf(k.t, "displaying logs of App %q...", appName)
	defer Logf(k.t, "done displaying logs of App %q.", appName)

	args := []string{
		"logs",
		"--space", SpaceFromContext(ctx),
		appName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutput(ctx, k.t, output), errs
}

// VerifyEchoLogsOutput verifies the expected output by tailing logs of an
// echo App.
func (k *Kf) VerifyEchoLogsOutput(
	ctx context.Context,
	appName string,
	expectedLogLine string,
	logOutput <-chan string,
	errs <-chan error,
) {
	appTimeout := 90 * time.Second

	// Hit the App via the proxy. This makes sure the App is handling
	// traffic as expected and ensures the proxy works. We use the proxy
	// for two reasons:
	// 1. Test the proxy.
	// 2. Tests work even if a domain isn't setup.

	k.WithProxy(ctx, appName, func(url string) {
		// Write out an expected log line until the test dies. We do this more
		// than once because we can't guarantee much about logs.
		for i := 0; i < 10; i++ {
			resp, respCancel := RetryPost(ctx, k.t, url, appTimeout, http.StatusOK, expectedLogLine)
			resp.Body.Close()
			respCancel()
		}
		// Wait around for the log to stream out. If it doesn't after a while,
		// fail the test.
		timer := time.NewTimer(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				k.t.Fatal("context cancelled")
			case <-timer.C:
				k.t.Fatal("timed out waiting for log line")
			case line := <-logOutput:
				if line == expectedLogLine {
					return
				}
			case err := <-errs:
				k.t.Fatal(err)
			}
		}
	})
}

// VerifyLogsOutput verifies the expected output by tailing logs of an
// echo App.
func (k *Kf) VerifyLogOutput(
	ctx context.Context,
	appName string,
	expectedLogLine string,
	timeout time.Duration,
	extraArgs ...string,
) {
	logOutput, errs := k.Logs(ctx, appName, extraArgs...)

	// Wait around for the log to stream out. If it doesn't after a while,
	// fail the test.
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-ctx.Done():
			k.t.Fatal("context cancelled")
		case <-timer.C:
			k.t.Fatal("timed out waiting for log line")
		case line := <-logOutput:
			if line == expectedLogLine {
				return
			}
		case err := <-errs:
			k.t.Fatal(err)
		}
	}
}

// VerifyTaskLogOutput verifies the expected output by tailing logs of a task.
func (k *Kf) VerifyTaskLogOutput(
	ctx context.Context,
	appName string,
	expectedLogLine string,
	timeout time.Duration,
) {
	k.VerifyLogOutput(ctx, appName, expectedLogLine, timeout, "--task")
}

// AppInfo is the information returned by listing an App. It is returned by
// ListApp.
type AppInfo struct {
	Name      string
	Instances string
	Memory    string
	Disk      string
	CPU       string
	URLs      []string
	Ready     string
	Reason    string
}

// TaskInfo is the information returned by listing tasks.
type TaskInfo struct {
	Name        string
	ID          string
	DisplayName string
	Status      string
	Reason      string
}

// JobInfo is the information returned by listed Jobs.
type JobInfo struct {
	Name     string
	Schedule string
	Suspend  bool
}

// SpaceUser is the information returned by listing space users.
type SpaceUser struct {
	Name  string
	Kind  string
	Roles []string
}

// Apps returns all the Apps from `kf app`
func (k *Kf) Apps(ctx context.Context) map[string]AppInfo {
	Logf(k.t, "listing Apps...")
	defer Logf(k.t, "done listing Apps.")
	k.t.Helper()
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"apps",
			"--space", SpaceFromContext(ctx),
		},
	})

	// Throw away stderr, we don't want it, but we still have to drain it
	// otherwise we could fill up a buffer and block.
	drainReader(output.Stderr)
	output.Stderr = strings.NewReader("")

	apps := CombineOutputStr(ctx, k.t, output, errs)

	if len(apps) == 0 {
		return nil
	}

	results := map[string]AppInfo{}
	for _, app := range apps[1:] {
		f := strings.Fields(app)
		if len(f) <= 1 {
			continue
		}

		var name, instances, memory, disk, CPU string
		var urls []string
		if f[2] == "(autoscaled" {
			// autoscaled App status.
			name = f[0]
			instances = strings.Join(f[1:6], " ")
			memory = f[6]
			disk = f[7]
			CPU = f[8]
			if len(f) > 9 {
				urls = strings.Split(f[9], ",")
			}
		} else {
			name = f[0]
			instances = f[1]
			memory = f[2]
			disk = f[3]
			CPU = f[4]
			if len(f) > 5 {
				urls = strings.Split(f[5], ",")
			}
		}

		results[name] = AppInfo{
			Name:      name,
			Instances: instances,
			Memory:    memory,
			Disk:      disk,
			CPU:       CPU,
			URLs:      urls,
		}
	}

	return results
}

// App gets a single App by name and returns its JSON representation.
func (k *Kf) App(ctx context.Context, name string) json.RawMessage {
	k.t.Helper()

	Logf(k.t, "getting App %s", name)
	defer Logf(k.t, "done getting App %s", name)

	stdout, stderr, err := k.RunSynchronous(ctx, KfTestConfig{
		Args: []string{
			"app",
			"--space", SpaceFromContext(ctx),
			"-o", "json",
			name,
		},
	})

	if err != nil {
		k.t.Fatalf("%s: %v", stderr, err)
	}

	if !json.Valid(stdout) {
		k.t.Fatal("App JSON was invalid:", string(stdout))
	}

	return stdout
}

// Delete deletes an App.
func (k *Kf) Delete(ctx context.Context, appName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "deleting %q...", appName)
	defer Logf(k.t, "done deleting %q.", appName)
	args := []string{
		"delete",
		"--space", SpaceFromContext(ctx),
		appName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Stop stops an App.
func (k *Kf) Stop(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "stopping %q...", appName)
	defer Logf(k.t, "done stopping %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"stop",
			"--space", SpaceFromContext(ctx),
			appName,
		},
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Start starts an App.
func (k *Kf) Start(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "starting %q...", appName)
	defer Logf(k.t, "done starting %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"start",
			"--space", SpaceFromContext(ctx),
			appName,
		},
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Restart restarts an App.
func (k *Kf) Restart(ctx context.Context, appName string) {
	k.t.Helper()
	Logf(k.t, "restarting %q...", appName)
	defer Logf(k.t, "done restarting %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"restart",
			"--space", SpaceFromContext(ctx),
			appName,
		},
	})
	StreamOutput(ctx, k.t, output, errs)
}

// WithProxy starts a proxy for an App and invokes the given function with its
// address. The proxy is closed when the function exits.
func (k *Kf) WithProxy(ctx context.Context, appName string, f func(addr string)) {
	k.t.Helper()
	Logf(k.t, "running proxy for %q...", appName)
	defer Logf(k.t, "done running proxy for %q.", appName)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"proxy",
			"--space", SpaceFromContext(ctx),
			appName,
			"--port=0",
		},
	})

	// Display stderr only and capture stdout. Stdout has the address.
	stdout := output.Stdout
	output.Stdout = strings.NewReader("")

	// NOTE: We don't want the StreamOutput to fail on a different go routine,
	// so we are NOT going to pass it the errors channel.
	go StreamOutput(ctx, k.t, output, nil)

	select {
	case err := <-errs:
		if k.t.Failed() || wontFailFromContext(ctx) {
			Logf(k.t, "Logging instead of failing: %v", err)
			return
		} else {
			k.t.Fatal(err)
		}
	case addr := <-CombineOutput(ctx, k.t, KfTestOutput{
		Stderr: strings.NewReader(""),
		Stdout: stdout,
	}):
		// Assert that we received a valid url from proxy. Proxy may fail and
		// output "FAIL" which should not be interpreted as the url.
		if u, err := url.Parse(addr); err != nil || u.Host == "" {
			k.t.Fatalf("proxy returned invalid url: %q, %v", addr, err)
			return
		}
		f(addr)
	}
}

// WithProxyRoute starts a proxy for a route and invokes the given function
// with its address. The proxy is closed when the function exits.
func (k *Kf) WithProxyRoute(ctx context.Context, routeHost string, f func(addr string)) {
	k.t.Helper()
	Logf(k.t, "running proxy for %q...", routeHost)
	defer Logf(k.t, "done running proxy for %q.", routeHost)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"proxy-route",
			"--space", SpaceFromContext(ctx),
			routeHost,
			"--port=0",
		},
	})

	// Display stderr only and capture stdout. Stdout has the address.
	stdout := output.Stdout
	output.Stdout = strings.NewReader("")

	// NOTE: We don't want the StreamOutput to fail on a different go routine,
	// so we are NOT going to pass it the errors channel.
	go StreamOutput(ctx, k.t, output, nil)

	select {
	case err := <-errs:
		k.t.Fatal(err)
	case addr := <-CombineOutput(ctx, k.t, KfTestOutput{
		Stderr: strings.NewReader(""),
		Stdout: stdout,
	}):
		f(addr)
	}
}

// SetEnv sets the environment variable for an App.
func (k *Kf) SetEnv(ctx context.Context, appName, name, value string) {
	k.t.Helper()
	Logf(k.t, "running set-env %s=%s for %q...", name, value, appName)
	defer Logf(k.t, "done running set-env %s=%s for %q.", name, value, appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"set-env",
			"--space", SpaceFromContext(ctx),
			appName,
			name,
			value,
		},
	})
	StreamOutput(ctx, k.t, output, errs)
}

// UnsetEnv unsets the environment variable for an App.
func (k *Kf) UnsetEnv(ctx context.Context, appName, name string) {
	k.t.Helper()
	Logf(k.t, "running unset-env %s for %q...", name, appName)
	defer Logf(k.t, "done running unset-env %s for %q.", name, appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"unset-env",
			"--space", SpaceFromContext(ctx),
			appName,
			name,
		},
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Env displays the environment variables for an App.
func (k *Kf) Env(ctx context.Context, appName string) map[string]string {
	k.t.Helper()
	Logf(k.t, "running env for %q...", appName)
	defer Logf(k.t, "done running env for %q.", appName)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"env",
			"--space", SpaceFromContext(ctx),
			appName,
		},
	})
	envs := CombineOutputStr(ctx, k.t, output, errs)

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

// WaitForCluster runs doctor repeatedly until the cluster is ready or the
// operation times out.
func (k *Kf) WaitForCluster(ctx context.Context) {
	k.t.Helper()
	k.Doctor(ctx, "--delay", "5s", "--retries", "12")
}

// Doctor runs the doctor command.
func (k *Kf) Doctor(ctx context.Context, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running doctor...")
	defer Logf(k.t, "done running doctor.")
	args := []string{"doctor"}
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Buildpacks runs the buildpacks command.
func (k *Kf) Buildpacks(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running buildpacks...")
	defer Logf(k.t, "done running buildpacks.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"buildpacks",
			"--space", SpaceFromContext(ctx),
		},
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// Stacks runs the stacks command.
func (k *Kf) Stacks(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running stacks...")
	defer Logf(k.t, "done running stacks.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"stacks",
			"--space", SpaceFromContext(ctx),
		},
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// CreateRoute runs the create-route command.
func (k *Kf) CreateRoute(ctx context.Context, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-route...")
	defer Logf(k.t, "done running create-route.")

	args := []string{
		"create-route",
		"--space", SpaceFromContext(ctx),
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// MapRoute runs the map-route command.
func (k *Kf) MapRoute(ctx context.Context, appName, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running map-route...")
	defer Logf(k.t, "done running map-route.")

	args := []string{
		"map-route",
		"--space", SpaceFromContext(ctx),
		appName,
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// UnmapRoute runs the unmap-route command.
func (k *Kf) UnmapRoute(ctx context.Context, appName, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running unmap-route...")
	defer Logf(k.t, "done running unmap-route.")

	args := []string{
		"unmap-route",
		"--space", SpaceFromContext(ctx),
		appName,
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// DeleteRoute runs the delete-route command.
func (k *Kf) DeleteRoute(ctx context.Context, domain string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-route...")
	defer Logf(k.t, "done running delete-route.")

	args := []string{
		"delete-route",
		"--space", SpaceFromContext(ctx),
		domain,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Routes runs the routes command.
func (k *Kf) Routes(ctx context.Context) []string {
	k.t.Helper()
	Logf(k.t, "running routes...")
	defer Logf(k.t, "done running routes.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"routes",
			"--space", SpaceFromContext(ctx),
		},
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

func (k *Kf) Jobs(ctx context.Context, appName string) []JobInfo {
	return k.jobLists(ctx, appName, "jobs")
}

func (k *Kf) JobSchedules(ctx context.Context, appName string) []JobInfo {
	return k.jobLists(ctx, appName, "job-schedules")
}

func (k *Kf) jobLists(ctx context.Context, appName, cmd string) []JobInfo {
	k.t.Helper()
	Logf(k.t, "running %s...", cmd)
	defer Logf(k.t, "done running %s.", cmd)
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			cmd,
			"--space", SpaceFromContext(ctx),
		},
	})

	// Throw away stderr, we don't want it, but we still have to drain it
	// otherwise we could fill up a buffer and block.
	drainReader(output.Stderr)
	output.Stderr = strings.NewReader("")

	jobs := CombineOutputStr(ctx, k.t, output, errs)

	if len(jobs) == 0 {
		return nil
	}

	results := []JobInfo{}

	for _, task := range jobs[1:] {
		f := strings.Fields(task)
		if len(f) <= 1 {
			continue
		}

		var name, schedule, suspend string
		name = f[0]
		schedule = f[1]
		suspend = f[2]

		var suspendBool bool
		if suspend == "<nil>" {
			suspendBool = false
		} else {
			suspendBool = true
		}

		results = append(results, JobInfo{
			Name:     name,
			Schedule: schedule,
			Suspend:  suspendBool,
		})
	}
	return results
}

// Tasks runs the tasks command with given App name.
func (k *Kf) Tasks(ctx context.Context, appName string) []TaskInfo {
	k.t.Helper()
	Logf(k.t, "running tasks...")
	defer Logf(k.t, "done running tasks.")
	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"tasks",
			appName,
			"--space", SpaceFromContext(ctx),
		},
	})

	// Throw away stderr, we don't want it, but we still have to drain it
	// otherwise we could fill up a buffer and block.
	drainReader(output.Stderr)
	output.Stderr = strings.NewReader("")

	tasks := CombineOutputStr(ctx, k.t, output, errs)

	if len(tasks) == 0 {
		return nil
	}

	results := []TaskInfo{}

	for _, task := range tasks[1:] {
		f := strings.Fields(task)
		if len(f) <= 1 {
			continue
		}

		var name, id, displayname, status, reason string
		name = f[0]
		id = f[1]
		displayname = f[2]
		status = f[5]
		reason = f[6]

		results = append(results, TaskInfo{
			Name:        name,
			ID:          id,
			DisplayName: displayname,
			Status:      status,
			Reason:      reason,
		})
	}
	return results
}

// SpaceUsers returns a list of space users within the current space.
func (k *Kf) SpaceUsers(ctx context.Context) []SpaceUser {
	k.t.Helper()
	Logf(k.t, "running space-users...")
	defer Logf(k.t, "done running space-users.")

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"space-users",
			"--space", SpaceFromContext(ctx),
		},
	})

	// Throw away stderr, we don't want it, but we still have to drain it
	// otherwise we could fill up a buffer and block.
	drainReader(output.Stderr)
	output.Stderr = strings.NewReader("")

	spaceUsers := CombineOutputStr(ctx, k.t, output, errs)

	if len(spaceUsers) == 0 {
		return nil
	}

	results := []SpaceUser{}

	for _, spaceUser := range spaceUsers[1:] {
		f := strings.Fields(spaceUser)
		if len(f) <= 1 {
			continue
		}
		var name, kind string
		var roles []string
		name = f[0]
		kind = f[1]
		roles = strings.Split(strings.Trim(f[2], "[]"), " ")

		results = append(results, SpaceUser{
			Name:  name,
			Kind:  kind,
			Roles: roles,
		})
	}
	return results
}

// CreateSpace runs the create-space command.
func (k *Kf) CreateSpace(ctx context.Context) (string, []string) {
	k.t.Helper()

	Logf(k.t, "running create-space...")
	defer Logf(k.t, "done running create-space.")

	// Build the space name.
	space := fmt.Sprintf("s-%x", md5.Sum([]byte(k.t.Name())))

	if os.Getenv(RandomSpaceNameEnvVar) == "true" {
		// Replace the deterministic Space name with a random one.
		space = fmt.Sprintf("%s-%d", space, rand.Uint64())
	}

	// First check to see if the Space already exists
	if k.spaceExists(ctx, space) {
		// Already exists, move on.
		Logf(k.t, "Space %s already exists... Skipping create...", space)
		return space, nil
	}

	domain := os.Getenv(SpaceDomainEnvVar)
	if domain == "" {
		domain = DefaultSpaceDomain
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: []string{
			"create-space",
			"--domain", domain,
			space,
		},
	})

	if os.Getenv(KeepSpaces) != "true" {
		k.t.Cleanup(func() {
			k.DeleteSpace(context.Background(), space)
		})
	}

	return space, CombineOutputStr(ctx, k.t, output, errs)
}

// DeleteSpace runs the delete-space command.
// This command is run asynchronously since every test uses a unique space.
func (k *Kf) DeleteSpace(ctx context.Context, space string) []string {
	k.t.Helper()
	Logf(k.t, "running delete-space...")
	defer Logf(k.t, "done running delete-space.")
	output, errs := k.kf(
		WithWontFailContext(DetachContext(ctx)),
		k.t,
		KfTestConfig{
			Args: []string{
				"delete-space",
				space,
				"--async",
			},
		})
	return CombineOutputStr(ctx, k.t, output, errs)
}

var readyTrueRegexp = regexp.MustCompile(`\sTrue\s`)

func (k *Kf) spaceExists(ctx context.Context, spaceName string) bool {
	for _, s := range k.Spaces(ctx) {
		if strings.HasPrefix(s, spaceName) &&
			// Ensure Space is marked "Ready True"
			readyTrueRegexp.MatchString(s) {
			return true
		}
	}

	return false
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
	return CombineOutputStr(ctx, k.t, output, errs)
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
	return CombineOutputStr(ctx, k.t, output, errs)
}

// Marketplace runs the marketplace command and returns the output.
func (k *Kf) Marketplace(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running marketplace...")
	defer Logf(k.t, "done running marketplace.")

	args := []string{
		"marketplace",
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// CreateServiceBroker runs the create-service-broker command.
func (k *Kf) CreateServiceBroker(ctx context.Context, brokerName string, username string, password string, url string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-service-broker...")
	defer Logf(k.t, "done running create-service-broker.")

	args := []string{
		"create-service-broker",
		"--space", SpaceFromContext(ctx),
		brokerName,
		username,
		password,
		url,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// DeleteServiceBroker runs the delete-service-broker command.
func (k *Kf) DeleteServiceBroker(ctx context.Context, brokerName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-service-broker...")
	defer Logf(k.t, "done running delete-service-broker.")

	args := []string{
		"delete-service-broker",
		"--space", SpaceFromContext(ctx),
		brokerName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// CreateService runs the create-service command.
func (k *Kf) CreateService(ctx context.Context, serviceClass string, servicePlan string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-service...")
	defer Logf(k.t, "done running create-service.")

	args := []string{
		"create-service",
		"--space", SpaceFromContext(ctx),
		serviceClass,
		servicePlan,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// CreateVolumeService runs the create-service command and create a VolumeServiceInstance.
func (k *Kf) CreateVolumeService(ctx context.Context, serviceInstanceName string, extraArgs ...string) {
	k.CreateService(ctx, nfsServiceClassName, nfsServicePlanExisting, serviceInstanceName, extraArgs...)
}

// CreateUserProvidedService runs the create-user-provided-service command.
func (k *Kf) CreateUserProvidedService(ctx context.Context, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running create-user-provided-service...")
	defer Logf(k.t, "done running create-user-provided-service.")

	args := []string{
		"create-user-provided-service",
		"--space", SpaceFromContext(ctx),
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// UpdateUserProvidedService runs the update-user-provided-service command.
func (k *Kf) UpdateUserProvidedService(ctx context.Context, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running update-user-provided-service...")
	defer Logf(k.t, "done running update-user-provided-service.")

	args := []string{
		"update-user-provided-service",
		"--space", SpaceFromContext(ctx),
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Services runs the services command.
func (k *Kf) Services(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running services...")
	defer Logf(k.t, "done running services.")

	args := []string{
		"services",
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// DeleteService runs the delete-service command.
func (k *Kf) DeleteService(ctx context.Context, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running delete-service...")
	defer Logf(k.t, "done running delete-service.")

	args := []string{
		"delete-service",
		"--space", SpaceFromContext(ctx),
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// BindService runs the bind-service command.
func (k *Kf) BindService(ctx context.Context, appName string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running bind-service...")
	defer Logf(k.t, "done running bind-service.")

	args := []string{
		"bind-service",
		"--space", SpaceFromContext(ctx),
		appName,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// Bindings runs the services command.
func (k *Kf) Bindings(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running bindings...")
	defer Logf(k.t, "done running bindings.")

	args := []string{
		"bindings",
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	return CombineOutputStr(ctx, k.t, output, errs)
}

// UnbindService runs the unbind-service command.
func (k *Kf) UnbindService(ctx context.Context, appName string, serviceInstanceName string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running unbind-service...")
	defer Logf(k.t, "done running unbind-service.")

	args := []string{
		"unbind-service",
		"--space", SpaceFromContext(ctx),
		appName,
		serviceInstanceName,
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
}

// VcapServices runs the vcap-services command.
func (k *Kf) VcapServices(ctx context.Context, extraArgs ...string) []byte {
	k.t.Helper()
	Logf(k.t, "running vcap-services...")
	defer Logf(k.t, "done running vcap-services.")

	args := []string{
		"vcap-services",
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})

	// Display stderr only and capture stdout. Stdout has the data.
	stdout := output.Stdout
	output.Stdout = strings.NewReader("")

	// NOTE: We don't want the StreamOutput to fail on a different go routine,
	// so we are NOT going to pass it the errors channel.
	go StreamOutput(ctx, k.t, output, nil)

	var out []string
	for {
		select {
		case err := <-errs:
			k.t.Fatal(err)
			return nil
		case data, ok := <-CombineOutput(ctx, k.t, KfTestOutput{
			Stderr: strings.NewReader(""),
			Stdout: stdout,
		}):
			if !ok {
				return []byte(strings.Join(out, "\n"))
			}
			out = append(out, data)
		}
	}
}

// ConfigureSpace runs any configure-space sub-command with the
// space configured on the context.
func (k *Kf) ConfigureSpace(ctx context.Context, extraArgs ...string) []string {
	k.t.Helper()
	Logf(k.t, "running configure-space...")
	defer Logf(k.t, "done running configure-space.")

	args := []string{
		"configure-space",
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})

	return CombineOutputStr(ctx, k.t, output, errs)
}

// Scale runs the scale command.
func (k *Kf) Scale(ctx context.Context, appName string, scale string) {
	k.t.Helper()
	Logf(k.t, "running scale...")
	defer Logf(k.t, "done running scale")

	args := []string{
		"scale",
		appName,
		"--instances",
		scale,
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: args,
	})
	StreamOutput(ctx, k.t, output, errs)
}

// WrapV2Buildpack runs the wrap-v2-buildpack command.
func (k *Kf) WrapV2Buildpack(ctx context.Context, buildpackName, v2URL string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running wrap-v2-buildpack...")
	defer Logf(k.t, "done running wrap-v2-buildpack")

	args := []string{
		"wrap-v2-buildpack",
		buildpackName,
		v2URL,
	}
	args = append(args, extraArgs...)

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: args,
	})
	StreamOutput(ctx, k.t, output, errs)
}

// RunCommand runs the provided commanded with provided args.
func (k *Kf) RunCommand(ctx context.Context, cmd string, extraArgs ...string) {
	k.t.Helper()
	Logf(k.t, "running %s...", cmd)
	defer Logf(k.t, "done running %s.", cmd)

	args := []string{
		cmd,
		"--space", SpaceFromContext(ctx),
	}

	output, errs := k.kf(ctx, k.t, KfTestConfig{
		Args: append(args, extraArgs...),
	})
	StreamOutput(ctx, k.t, output, errs)
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

// ContextWithApp returns a context that has the App information. The
// App name can be fetched via AppFromContext.
func ContextWithApp(ctx context.Context, appName string) context.Context {
	return context.WithValue(ctx, appKey{}, appName)
}

// AppFromContext returns the service plan name given a context that has been setup
// via ContextWithApp.
func AppFromContext(ctx context.Context) string {
	return ctx.Value(appKey{}).(string)
}

var testIDOnce sync.Once
var testID string

// TestID returns a ID. This is either populated by
// IntegrationTestID (if provided), or a random value.
func TestID() string {
	testIDOnce.Do(func() {
		testID = os.Getenv(IntegrationTestID)
		if testID == "" {
			testID = fmt.Sprintf("rand-%d", rand.Uint64())
		}
	})
	return testID
}

// ExpectedAddr returns the expected address for integration tests given a
// hostname and URL path.
func ExpectedAddr(hostname, urlPath string) string {
	domain := os.Getenv(SpaceDomainEnvVar)
	hostnameDomain := domain
	if hostname != "" {
		hostnameDomain = fmt.Sprintf("%s.%s", hostname, domain)
	}
	if len(urlPath) == 0 {
		return hostnameDomain
	}
	return hostnameDomain + path.Join("/", urlPath)
}

// unionMaps is a duplicate from v1alpha. We can't pull it in because it would
// result in a circular dependency.
func unionMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func drainReader(r io.Reader) {
	go func() {
		// Copying from the reader to the nopWriter will read every byte from
		// the reader UNTIL it returns an EOF. At this point the io.Copy
		// function will gracefully exit.
		io.Copy(nopWriter{}, r)
	}()
}

type nopWriter struct{}

func (nopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}
