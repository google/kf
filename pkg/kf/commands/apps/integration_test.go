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

package apps

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	appTimeout = 90 * time.Second
)

var currentPort int = 8086

// TestIntegration_Push pushes the echo App, lists it to ensure it can find a
// domain, uses the proxy command and then posts to it. It finally deletes the
// App.
func TestIntegration_Push(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		// For the purposes of this test the results SHOULD NOT be cached.
		kf.CachePush(ctx, appName,
			filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo"),
		)
		integration.CheckEchoApp(ctx, t, kf, appName, integration.ExpectedAddr(appName, ""))
	})
}

// TestIntegration_Push_update pushes the echo App, uses the proxy command and
// then posts to it. It then updates the App to the helloworld App by pushing
// to the same App name. It finally deletes the App.
func TestIntegration_Push_update(t *testing.T) {
	// This test needs more time because pushes the App twice.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		echoPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo")
		kf.CachePush(ctx, appName, echoPath)
		integration.CheckEchoApp(ctx, t, kf, appName, integration.ExpectedAddr(appName, ""))

		helloPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld")
		kf.CachePush(ctx, appName, helloPath)

		// BUG(730): it takes a moment after the App becomes ready to
		// reconcile the routes, the App is accessible but still points to the
		// old one.
		time.Sleep(45 * time.Second)
		integration.CheckHelloWorldApp(ctx, t, kf, appName, integration.ExpectedAddr(appName, ""))
	})
}

// TestIntegration_Push_docker pushes the echo App via a prebuilt docker
// image, lists it to ensure it can find a domain, uses the proxy command and
// then posts to it. It finally deletes the App.
func TestIntegration_Push_docker(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push", fmt.Sprint(time.Now().UnixNano()))

		// us-docker.pkg.dev/cloudrun/container/hello is a sample pre-built
		// container image for cloud run.
		kf.Push(ctx, appName, "--docker-image=us-docker.pkg.dev/cloudrun/container/hello")

		integration.CheckApp(ctx, t, kf, appName, []string{integration.ExpectedAddr(appName, "")}, func(ctx context.Context, t *testing.T, addr string) {
			resp, respCancel := integration.RetryGet(ctx, t, addr, appTimeout, http.StatusOK)
			defer resp.Body.Close()
			defer respCancel()
			testutil.AssertEqual(t, "status code", http.StatusOK, resp.StatusCode)
			integration.Logf(t, "done hitting echo App to ensure its working.")
		})
	})
}

// TestIntegration_Push_dockerfile pushes a sample dockerfile App then attempts
// to connect to it to ensure it started correctly.
func TestIntegration_Push_dockerfile(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-dockerfile", fmt.Sprint(time.Now().UnixNano()))
		appPath := filepath.Join(integration.RootDir(ctx, t), "samples", "apps", "helloworld")

		kf.Push(ctx, appName, "--path", appPath, "--dockerfile", "Dockerfile")

		integration.CheckHelloWorldApp(ctx, t, kf, appName, integration.ExpectedAddr(appName, ""))
	})
}

func TestIntegration_SSH(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-ssh", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		echoPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo")
		kf.CachePush(ctx, appName, echoPath)

		helloWorld := "hello, world!"
		lines := kf.SSH(ctx, appName, "-c", "/bin/echo", "-c", helloWorld, "-T")

		testutil.AssertContainsAll(t, strings.Join(lines, "\n"), []string{helloWorld})
	})
}

// TestIntegration_StopStart pushes the echo App and uses the proxy command
// and then posts to it. It posts to it to ensure the App is ready. It then
// stops it and ensures it can no longer reach the App. It then starts it and
// tries posting to it again. It finally deletes the App.
func TestIntegration_StopStart(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		echoPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo")
		kf.CachePush(ctx, appName, echoPath)

		// Hit the App via the proxy. This makes sure the App is handling
		// traffic as expected and ensures the proxy works. We use the proxy
		// for two reasons:
		// 1. Test the proxy.
		// 2. Tests work even if a domain isn't setup.
		integration.Logf(t, "hitting echo App to ensure it's working...")

		kf.WithProxy(ctx, appName, func(url string) {
			{
				resp, respCancel := integration.RetryPost(ctx, t, url, appTimeout, http.StatusOK, "testing")
				defer resp.Body.Close()
				defer respCancel()
				integration.Logf(t, "done hitting echo App to ensure it's working.")
			}

			integration.Logf(t, "stoping echo App...")
			kf.Stop(ctx, appName)
			integration.Logf(t, "done stopping echo App.")

			{
				integration.Logf(t, "hitting echo App to ensure it's NOT working...")
				resp, respCancel := integration.RetryPost(ctx, t, url, appTimeout, http.StatusNotFound, "testing")
				defer resp.Body.Close()
				defer respCancel()
				integration.Logf(t, "done hitting echo App to ensure it's NOT working.")
			}

			integration.Logf(t, "starting echo App...")
			kf.Start(ctx, appName)
			integration.Logf(t, "done starting echo App.")

			{
				integration.Logf(t, "hitting echo App to ensure it's working...")
				resp, respCancel := integration.RetryPost(ctx, t, url, 5*time.Minute, http.StatusOK, "testing")
				defer resp.Body.Close()
				defer respCancel()
				integration.Logf(t, "done hitting echo App to ensure it's working.")
			}
		})
	})
}

// TestIntegration_IdempotentDockerPush asserts that an App can be pushed to
// several times with the --docker-image flag and have it work each time. This
// test is important because when an App is pushed with --docker-image, it
// doesn't really need a build and therefore the App might not have any
// properties that change.
func TestIntegration_IdempotentDockerPush(t *testing.T) {
	// This test needs more time because pushes the App multiple times.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-idempotent-docker", fmt.Sprint(time.Now().UnixNano()))

		kf.Push(ctx, appName, "--docker-image=us-docker.pkg.dev/cloudrun/container/hello")
		kf.Push(ctx, appName, "--docker-image=us-docker.pkg.dev/cloudrun/container/hello")

		// Getting this far implies success
	})
}

// TestIntegration_Delete pushes an App and then deletes it.
func TestIntegration_Delete(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-delete", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// simplies replies with the same body that was posted.
		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo"))

		// List the apps and make sure we can find the App.
		integration.Logf(t, "ensuring App is there...")
		_, ok := kf.Apps(ctx)[appName]
		testutil.AssertEqual(t, "App presence", true, ok)
		integration.Logf(t, "done ensuring App is there.")

		// Delete the App.
		// This is run synchronously so that the test can ensure the App is deleted afterwards.
		kf.Delete(ctx, appName)

		// Make sure the App is gone
		// List the apps and make sure we can find the App.
		integration.Logf(t, "ensuring App is gone from list...")
		_, ok = kf.Apps(ctx)[appName]
		testutil.AssertEqual(t, "App exists", false, ok)
		integration.Logf(t, "done ensuring App is gone.")
	})
}

// TestIntegration_Envs pushes the env sample App. It sets two variables while
// pushing, and another via SetEnv. It then reads them via Env. It then unsets
// one via Unset and reads them again via Env.
func TestIntegration_Envs(t *testing.T) {
	// This test needs more time because pushes the App and then manipulates
	// it in a way that will result in a new deployment.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-envs", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the envs App which
		// returns the set environment variables via JSON. Set two environment
		// variables (ENV1 and ENV2).
		kf.CachePush(ctx, appName,
			filepath.Join(integration.RootDir(ctx, t), "./samples/apps/envs"),
			"--env", "ENV1=VALUE1",
			"--env=ENV2=VALUE2",
		)

		integration.CheckEnvApp(ctx, t, kf, appName, map[string]string{
			"ENV1": "VALUE1", // Set on push
			"ENV2": "VALUE2", // Set on push
		}, integration.ExpectedAddr(appName, ""))

		// Unset the environment variables ENV1.
		kf.UnsetEnv(ctx, appName, "ENV1")

		// Overwrite ENV2 via set-env
		kf.SetEnv(ctx, appName, "ENV2", "OVERWRITE2")

		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		integration.AssertVars(ctx, t, kf, appName, map[string]string{
			"ENV2": "OVERWRITE2", // Set on push and overwritten via set-env
		}, []string{"ENV1"})
	})
}

// TestIntegration_NodeSelector sets a label on a Node in cluster,
// adds the same label on a space and then push a App in this space.
// It verifies the pod lands on the correct node and also verifies
// nodeSelector are present in the podSpec.
func TestIntegration_NodeSelector(t *testing.T) {
	var labelName = "testLabel"
	integration.RunKubeAPITest(context.Background(), t, func(apictx context.Context, t *testing.T) {
		integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
			integration.WithNodeLabel(apictx, labelName, t, func(labelValue string, nodeName string, k8s *kubernetes.Clientset) {

				appName := v1alpha1.GenerateName("integration-nodeselector", fmt.Sprint(time.Now().UnixNano()))

				// Add nodeselector on the space
				kf.ConfigureSpace(ctx, "set-nodeselector", labelName, labelValue)

				// Push an App and then clean it up. This pushes the echo App which
				// simplies replies with the same body that was posted.
				kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo"))

				// Verify the application and then check the podSpec to make sure nodeSelector was correctly set.
				checkNodeSelector(ctx, t, kf, appName, integration.SpaceFromContext(ctx), k8s, nodeName, map[string]string{labelName: labelValue}, integration.ExpectedAddr(appName, ""))
			})
		})
	})
}

// TestIntegration_Logs pushes the echo App, tails
// it's logs and then posts to it. It then waits for the expected logs. It
// finally deletes the App.
func TestIntegration_Logs(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-logs", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		kf.CachePush(ctx, appName, filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo"))

		logOutput, errs := kf.Logs(ctx, appName, "-n=30")
		expectedLogLine := fmt.Sprintf("testing-%d", time.Now().UnixNano())
		kf.VerifyEchoLogsOutput(ctx, appName, expectedLogLine, logOutput, errs)
	})
}

// TestIntegration_HealthType_process pushes a bash for-loop that writes out
// logs. It does not listen to a port.
func TestIntegration_HealthType_process(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-process", fmt.Sprint(time.Now().UnixNano()))

		expectedLogLine := fmt.Sprintf("testing-%d", time.Now().UnixNano())
		kf.Push(ctx, appName,
			"--health-check-type", "process",
			"--buildpack", "binary_buildpack",
			"--command", fmt.Sprintf("for i in $(seq 1 60); do echo %s && sleep 1; done", expectedLogLine),
		)

		logOutput, errs := kf.Logs(ctx, appName, "-n=30")

		// Wait around for the log to stream out. If it doesn't after a while,
		// fail the test.
		timer := time.NewTimer(30 * time.Second)
		for {
			select {
			case <-ctx.Done():
				t.Fatal("context cancelled")
			case <-timer.C:
				t.Fatal("timed out waiting for log line")
			case line := <-logOutput:
				if line == expectedLogLine {
					return
				}
			case err := <-errs:
				t.Fatal(err)
			}
		}
	})
}

// TestIntegration_LogsNoContainer tests that the logs command exits in a
// reasonable amount of time when logging an application that doesn't have a
// container (scaled to 0).
func TestIntegration_LogsNoContainer(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-logs-noc", fmt.Sprint(time.Now().UnixNano()))

		output, errs := kf.Logs(ctx, appName, "--recent")

		timer := time.NewTimer(15 * time.Second)
		for {
			select {
			case <-timer.C:
				t.Fatal("expected kf logs to exit")
			case _, ok := <-output:
				if !ok {
					// Success
					return
				}

				// Not closed, command still going.
			case err := <-errs:
				t.Fatal(err)
			}
		}
	})
}

// TestIntegration_Push_SigtermV2Buildpack builds the sigterm App with a v3
// buildpack and asserts it receives SIGTERM signal on stop.
func TestIntegration_Push_SigtermV3Buildpack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push-sigterm-v3", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the sigterm App which
		// waits for a SIGTERM and logs when received.
		appSrcPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/sigterm")
		kf.CachePush(ctx, appName, appSrcPath)
		integration.CheckSigtermApp(ctx, t, kf, appName)
	})
}

// TestIntegration_Push_SigtermV2Buildpack builds the sigterm App with a v2
// buildpack and asserts it receives SIGTERM signal on stop.
func TestIntegration_Push_SigtermV2Buildpack(t *testing.T) {
	// This test needs more time because pushes with V2 buildpacks.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push-sigterm-v2", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the sigterm App which
		// waits for a SIGTERM and logs when received.
		appSrcPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/sigterm")
		kf.CachePushV2(ctx, appName, appSrcPath)
		integration.CheckSigtermApp(ctx, t, kf, appName)
	})
}

// TestIntegration_Push_binaryBuildpack builds the hello world App and then
// pushes it, lists it to ensure it can find a domain, uses the proxy command
// and then posts to it. It finally deletes the App.
func TestIntegration_Push_binaryBuildpack(t *testing.T) {
	// This test needs more time because pushes with V2 buildpacks.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-push-bin", fmt.Sprint(time.Now().UnixNano()))

		// Push an App and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		// For the purposes of this test the results SHOULD NOT be cached.
		kf.Push(ctx, appName,
			"--buildpack", "binary_buildpack",
			"--command", "python -m SimpleHTTPServer $PORT",
		)
		integration.CheckApp(ctx, t, kf, appName, []string{integration.ExpectedAddr(appName, "")},
			func(ctx context.Context, t *testing.T, addr string) {
				// Just check to ensure we got a 200
				integration.RetryGet(ctx, t, addr, appTimeout, http.StatusOK)
			})
	})
}

// TestIntegration_ClusterLocal pushes 2 Apps and test the ability of
// communicating using the cluster.local domain.
func TestIntegration_ClusterLocal(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName1 := v1alpha1.GenerateName("integration-clusterLocalApp1", fmt.Sprint(time.Now().UnixNano()))
		appName2 := v1alpha1.GenerateName("integration-clusterLocalApp2", fmt.Sprint(time.Now().UnixNano()))

		// Push Apps and then clean it up. This pushes the echo Apps which
		// reply with the same body that was posted.
		appSrcPath := filepath.Join(integration.RootDir(ctx, t), "./samples/apps/helloworld")
		kf.CachePush(ctx, appName1, appSrcPath)
		kf.CachePush(ctx, appName2, appSrcPath)

		appEndpoint2 := fmt.Sprintf("%s.%s.svc.cluster.local", appName2, integration.SpaceFromContext(ctx))
		lines := kf.SSH(ctx, appName1, "-c", "curl", "-c", appEndpoint2, "-T")

		helleMessage := "hello from " + appName2
		fmt.Println("Lines: " + strings.Join(lines, "\n"))
		testutil.AssertContainsAll(t, strings.Join(lines, "\n"), []string{helleMessage})
	})
}

// TestIntegration_PushTaskWithRouteSetting pushes the echo App
// as a Task with route setting and ensures
// kf ignores the route setting and doesn't return an error.
func TestIntegration_PushTaskWithRouteSetting(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		appName := v1alpha1.GenerateName("integration-pushTaskWithRouteSetting", fmt.Sprint(time.Now().UnixNano()))

		// Push an App as a Task with random-route and then clean it up. This pushes the echo App which
		// replies with the same body that was posted.
		// For the purposes of this test the results SHOULD NOT be cached.
		kf.CachePush(ctx, appName,
			filepath.Join(integration.RootDir(ctx, t), "./samples/apps/echo"),
			"--task",
			"--random-route",
		)
	})
}

func TestIntegration_Push_With_Manifest(t *testing.T) {
	integration.RunKubeAPITest(context.Background(), t, func(apictx context.Context, t *testing.T) {
		integration.UpdateConfigMapBuildpack(apictx, t, func(apictx context.Context, t *testing.T) {
			integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
				appName := v1alpha1.GenerateName("integration-push-manifest", fmt.Sprint(time.Now().UnixNano()))
				//appName := "manifest-with-buildpack"
				fmt.Println("Application Name: ", appName)

				// Push an App and then clean it up. This pushes the echo App which
				// replies with the same body that was posted.
				// For the purposes of this test the results SHOULD NOT be cached.
				kf.Push(ctx, appName,
					"--path",
					filepath.Join(integration.RootDir(ctx, t), "./samples/apps/manifest-with-buildpack"),
				)
				//integration.CheckHelloWorldApp(ctx, t, kf, appName, integration.ExpectedAddr(appName, ""))
			})
		})
	})
}

// checkNodeSelector verifies the application and then compares the expected nodeSelectos with podSpec
func checkNodeSelector(
	ctx context.Context,
	t *testing.T,
	kf *integration.Kf,
	appName string,
	spaceName string,
	k8s *kubernetes.Clientset,
	expectedNodeName string,
	expectedNodeSelector map[string]string,
	expectedRoutes ...string,
) {
	integration.Logf(t, "Checking the application %s deployed correctly", appName)
	integration.CheckApp(ctx, t, kf, appName, expectedRoutes, func(ctx context.Context, t *testing.T, addr string) {

		integration.Logf(t, "Checking the podSpec has correct nodeSelectors")
		// Check if the podSpec has nodeSelector.
		pods, err := k8s.CoreV1().Pods(spaceName).List(ctx, v1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s, app.kubernetes.io/component=%s", appName, "app-server"),
		})

		if err != nil {
			t.Fatalf("Error getting pods: %v", err)
		}

		if len(pods.Items) > 0 {
			pod := pods.Items[0]
			testutil.AssertEqual(t, "nodeName", expectedNodeName, pod.Spec.NodeName)
			testutil.AssertEqual(t, "nodeSelectors", expectedNodeSelector, pod.Spec.NodeSelector)
			integration.Logf(t, "PASS: Found the correct nodeSelectos on the pod %s", pod.Name)
		} else {
			t.Fatalf("No pod found for the App=%s", appName)
		}
	})
}
