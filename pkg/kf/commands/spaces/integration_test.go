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

package spaces_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	. "github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

const (
	userNameManager   = "user-manager"
	userNameDeveloper = "user-developer"
	userNameAuditor   = "user-auditor"
)

func TestIntegration_RBAC(t *testing.T) {
	appName := fmt.Sprintf("integration-rbac-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/echo"

	// This tests several things and ends up taking longer than 5 minutes.
	// TODO: Can this be parallelized?
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	integration.RunKubeAPITest(ctx, t, func(apictx context.Context, t *testing.T) {
		integration.RunKfTest(apictx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
			kf.RunCommand(ctx, "set-space-role", userNameManager, "SpaceManager")
			verifySpaceUser(ctx, t, kf, userNameManager, "User", "space-manager")

			// Space Manager can assign roles to other users
			kf.RunCommand(ctx, "set-space-role", userNameDeveloper, "SpaceDeveloper", "--as", userNameManager)
			verifySpaceUser(ctx, t, kf, userNameDeveloper, "User", "space-developer")

			kf.RunCommand(ctx, "set-space-role", userNameAuditor, "SpaceAuditor", "--as", userNameManager)
			verifySpaceUser(ctx, t, kf, userNameAuditor, "User", "space-auditor")

			// Space Developer can push App.
			withImpersonationApp(ctx, t, kf, appName, userNameDeveloper, appPath, false, func(ctx context.Context) {
				app, ok := kf.Apps(ctx)[appName]
				AssertEqual(t, "app presence", true, ok)
				AssertEqual(t, "app instances", "1", app.Instances)

				// Space Developer can ssh into App.
				helloWorld := "hello, world!"
				lines := kf.SSH(ctx, appName, "-c", "/bin/echo", "-c", helloWorld, "-T", "--as", userNameDeveloper)
				AssertContainsAll(t, strings.Join(lines, "\n"), []string{helloWorld})

				// Space Developer can tail logs of the App.
				logOutput, errs := kf.Logs(ctx, appName, "-n=30", "--as", userNameDeveloper)
				expectedLogLine := fmt.Sprintf("testing-%d", time.Now().UnixNano())
				kf.VerifyEchoLogsOutput(ctx, appName, expectedLogLine, logOutput, errs)
			})

			namespace := integration.SpaceFromContext(ctx)

			verifyManagerPermision(apictx, t, namespace)
			verifyDeveloperPermission(apictx, t, namespace)
			verifyAuditorPermission(apictx, t, namespace)
		})
	})
}

type TestInput struct {
	title          string
	space          string
	verb           string
	group          string
	resource       string
	expectedOutput bool
}

func verifyManagerPermision(ctx context.Context, t *testing.T, namespace string) {
	tests := []TestInput{
		{title: "SpaceManager can not create Apps in space", space: namespace, verb: "create", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceManager can not update Apps in space", space: namespace, verb: "update", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceManager can not delete Apps in space", space: namespace, verb: "delete", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceManager can not create spaces in cluster", space: "", verb: "create", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceManager can not update spaces in cluster", space: "", verb: "update", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceManager can not patch spaces in cluster", space: "", verb: "patch", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceManager can not delete spaces in cluster", space: "", verb: "delete", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceManager can not list secrets in space", space: namespace, verb: "list", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not get secrets in space", space: namespace, verb: "get", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not watch secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not create secrets in space", space: namespace, verb: "list", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not update secrets in space", space: namespace, verb: "get", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not patch secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceManager can not delete secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
	}

	for _, test := range tests {
		verifyAccessAs(ctx, t, test.space, test.verb, test.group, test.resource, userNameManager, test.title, test.expectedOutput)
	}
}

func verifyDeveloperPermission(ctx context.Context, t *testing.T, namespace string) {
	tests := []TestInput{
		{title: "SpaceDeveloper manages all kf resources in space", space: namespace, verb: "*", group: "kf.dev", resource: "*", expectedOutput: true},
		{title: "SpaceDeveloper manages networking policies in space", space: namespace, verb: "*", group: "networking.k8s.io", resource: "networkpolicies", expectedOutput: true},
		{title: "SpaceDeveloper gets rolebindings in space", space: namespace, verb: "get", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceDeveloper lists rolebindings in space", space: namespace, verb: "list", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceDeveloper watches rolebindings in space", space: namespace, verb: "watch", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceDeveloper creates upload.kf.dev resources in space", space: namespace, verb: "create", group: "upload.kf.dev", resource: "*", expectedOutput: true},
		{title: "SpaceDeveloper proxies upload.kf.dev resources in space", space: namespace, verb: "proxy", group: "upload.kf.dev", resource: "*", expectedOutput: true},
		{title: "SpaceDeveloper gets spaces in cluster", space: "", verb: "get", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceDeveloper lists spaces in cluster", space: "", verb: "list", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceDeveloper watches spaces in cluster", space: "", verb: "watch", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceDeveloper gets clusterservicebrokers in cluster", space: "", verb: "get", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceDeveloper lists clusterservicebrokers in cluster", space: "", verb: "list", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceDeveloper watches clusterservicebrokers in cluster", space: "", verb: "watch", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceDeveloper creates secrets in space", space: namespace, verb: "create", group: "", resource: "secrets", expectedOutput: true},
		{title: "SpaceDeveloper updates secrets in space", space: namespace, verb: "update", group: "", resource: "secrets", expectedOutput: true},
		{title: "SpaceDeveloper patches secrets in space", space: namespace, verb: "patch", group: "", resource: "secrets", expectedOutput: true},
		{title: "SpaceDeveloper deletes secrets in space", space: namespace, verb: "delete", group: "", resource: "secrets", expectedOutput: true},
		{title: "SpaceDeveloper gets taskruns in cluster", space: namespace, verb: "get", group: "tekton.dev", resource: "taskruns", expectedOutput: true},
		{title: "SpaceDeveloper lists taskruns in cluster", space: namespace, verb: "list", group: "tekton.dev", resource: "taskruns", expectedOutput: true},
		{title: "SpaceDeveloper watches taskruns in cluster", space: namespace, verb: "watch", group: "tekton.dev", resource: "taskruns", expectedOutput: true},
		{title: "SpaceDeveloper can not update rolebindings in space", space: namespace, verb: "update", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceDeveloper can not patch rolebindings in space", space: namespace, verb: "patch", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceDeveloper can not create rolebindings in space", space: namespace, verb: "create", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceDeveloper can not delete rolebindings in space", space: namespace, verb: "delete", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceDeveloper can not list secrets in space", space: namespace, verb: "list", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceDeveloper can not get secrets in space", space: namespace, verb: "get", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceDeveloper can not watch secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceDeveloper can not create spaces in cluster", space: "", verb: "create", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceDeveloper can not update spaces in cluster", space: "", verb: "update", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceDeveloper can not patch spaces in cluster", space: "", verb: "patch", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceDeveloper can not delete spaces in cluster", space: "", verb: "delete", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceDeveloper can not list Apps in other space", space: "kf", verb: "list", group: "kf.dev", resource: "apps", expectedOutput: false},
	}

	for _, test := range tests {
		verifyAccessAs(ctx, t, test.space, test.verb, test.group, test.resource, userNameDeveloper, test.title, test.expectedOutput)
	}
}

func verifyAuditorPermission(ctx context.Context, t *testing.T, namespace string) {
	tests := []TestInput{
		{title: "SpaceAuditor gets Apps in space", space: namespace, verb: "get", group: "kf.dev", resource: "apps", expectedOutput: true},
		{title: "SpaceAuditor lists Apps in space", space: namespace, verb: "list", group: "kf.dev", resource: "apps", expectedOutput: true},
		{title: "SpaceAuditor watches Apps in space", space: namespace, verb: "watch", group: "kf.dev", resource: "apps", expectedOutput: true},
		{title: "SpaceAuditor gets rolebindings in space", space: namespace, verb: "get", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceAuditor lists rolebindings in space", space: namespace, verb: "list", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceAuditor watches rolebindings in space", space: namespace, verb: "watch", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: true},
		{title: "SpaceAuditor gets spaces in cluster", space: "", verb: "get", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceAuditor lists spaces in cluster", space: "", verb: "list", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceAuditor watches spaces in cluster", space: "", verb: "watch", group: "kf.dev", resource: "spaces", expectedOutput: true},
		{title: "SpaceAuditor gets clusterservicebrokers in cluster", space: "", verb: "get", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceAuditor lists clusterservicebrokers in cluster", space: "", verb: "list", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceAuditor watches clusterservicebrokers in cluster", space: "", verb: "watch", group: "kf.dev", resource: "clusterservicebrokers", expectedOutput: true},
		{title: "SpaceAuditor can not update rolebindings in space", space: namespace, verb: "update", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceAuditor can not patch rolebindings in space", space: namespace, verb: "patch", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceAuditor can not create rolebindings in space", space: namespace, verb: "create", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceAuditor can not delete rolebindings in space", space: namespace, verb: "delete", group: "rbac.authorization.k8s.io", resource: "rolebindings", expectedOutput: false},
		{title: "SpaceAuditor can not create spaces in cluster", space: "", verb: "create", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceAuditor can not update spaces in cluster", space: "", verb: "update", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceAuditor can not patch spaces in cluster", space: "", verb: "patch", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceAuditor can not delete spaces in cluster", space: "", verb: "delete", group: "kf.dev", resource: "spaces", expectedOutput: false},
		{title: "SpaceAuditor can not create Apps in space", space: namespace, verb: "create", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceAuditor can not update Apps in space", space: namespace, verb: "update", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceAuditor can not delete Apps in space", space: namespace, verb: "delete", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceAuditor can not list Apps in other space", space: "kf", verb: "list", group: "kf.dev", resource: "apps", expectedOutput: false},
		{title: "SpaceAuditor can not list secrets in space", space: namespace, verb: "list", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not get secrets in space", space: namespace, verb: "get", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not watch secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not create secrets in space", space: namespace, verb: "list", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not update secrets in space", space: namespace, verb: "get", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not patch secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
		{title: "SpaceAuditor can not delete secrets in space", space: namespace, verb: "watch", group: "", resource: "secrets", expectedOutput: false},
	}

	for _, test := range tests {
		verifyAccessAs(ctx, t, test.space, test.verb, test.group, test.resource, userNameAuditor, test.title, test.expectedOutput)
	}
}

func verifyAccessAs(ctx context.Context, t *testing.T, namespace, verb, group, resource, username, testTitle string, expectedOutput bool) {
	withImpersonationKubernetes(ctx, t, username, func(k8s *kubernetes.Clientset) {
		kfAllSelfSubjectAccessReview := v1.SelfSubjectAccessReview{
			Spec: v1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &v1.ResourceAttributes{
					Namespace: namespace,
					Verb:      verb,
					Group:     group,
					Resource:  resource,
				},
			},
		}

		result, err := k8s.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &kfAllSelfSubjectAccessReview, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("error when checking for access: %v", err)
		}
		AssertEqual(t, testTitle, expectedOutput, result.Status.Allowed)
		if result.Status.Allowed != expectedOutput {
			t.Fatalf("test failed, reason: %v", result.Status.Reason)
		}
	})
}

// withImpersonationApp creates an App with impersonation.
func withImpersonationApp(
	ctx context.Context,
	t *testing.T,
	kf *integration.Kf,
	appName string,
	username string,
	path string,
	isBroker bool,
	callback func(newCtx context.Context),
) {
	integration.WithAppArgs(ctx, t, kf, appName, path, isBroker, []string{"--as", username}, callback)
}

// withImpersonationKubernetes creates an impersonated Kubernetes client from
// the config on the context and passes it to the callback.
func withImpersonationKubernetes(ctx context.Context, t *testing.T, username string, callback func(k8s *kubernetes.Clientset)) {
	t.Helper()

	integration.WithRestConfig(ctx, t, func(cfg *rest.Config) {
		p := &config.KfParams{
			Impersonate: transport.ImpersonationConfig{
				UserName: username,
			},
		}

		cfg.Wrap(config.NewImpersonatingRoundTripperWrapper(p))
		k8s, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			t.Fatalf("Creating Kubernetes: %v", err)
		}

		callback(k8s)
	})
}

func verifySpaceUser(ctx context.Context, t *testing.T, kf *integration.Kf, username, kind, expectedRole string) {
	spaceUser := getSpaceUser(ctx, t, kf, username, kind)
	AssertEqual(t, "space user name", username, spaceUser.Name)
	AssertEqual(t, "space role", expectedRole, spaceUser.Roles[0])
}

func getSpaceUser(ctx context.Context, t *testing.T, kf *integration.Kf, username, kind string) (founduser integration.SpaceUser) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	spaceUsers := kf.SpaceUsers(ctx)

	if len(spaceUsers) == 0 {
		t.Fatalf("No space user is found")
	}

	for _, spaceUser := range spaceUsers {

		if spaceUser.Name == username && spaceUser.Kind == kind {
			founduser = spaceUser
			return
		}
	}
	t.Fatalf("No space user of name %q (kind %q) is found.", username, kind)
	return
}
