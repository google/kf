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

package install

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	KnativeBuildYAML      = "https://github.com/knative/build/releases/download/v0.7.0/build.yaml"
	KfNightlyBuildYAML    = "https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/releases/release-latest.yaml"
	GKEClusterVersion     = "1.13.7-gke.8"
	MachineType           = "n1-standard-4"
	ImageType             = "COS"
	DiskType              = "pd-standard"
	DiskSize              = "100"
	NumNodes              = "3"
	DefaultMaxPodsPerNode = "110"
	Addons                = "HorizontalPodAutoscaling,HttpLoadBalancing,Istio,CloudRun"
)

var (
	prefixColor = color.New(color.FgHiBlue, color.Bold)
	labelColor  = color.New(color.FgHiYellow, color.Bold)
)

// NewGKECommand creates a command that can install kf to GKE+Cloud Run.
// TODO: This installer is using gcloud and kubectl under the hood. We should
// really replace gcloud with google.golang.org/api and kubectl with
// k8s.io/client-go
func NewGKECommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "gke [subcommand]",
		Short: "Install kf to GKE with Cloud Run",
		Long: `
This interactive installer will walk you through a process to install GKE with
Cloud Run and then kf. Its really just running scripts under the hood,
therefore you have to have gcloud and kubectl installed and available in the
path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx := setContextOutput(context.Background(), cmd.ErrOrStderr())
			ctx = setLogPrefix(ctx, "Install GKE+CR+Kf")
			ctx = setVerbosity(ctx, verbose)

			// Select the desired project
			projID, err := selectProject(ctx)
			if err != nil {
				return err
			}

			// Select the desired GKE cluster
			clusterName, zone, err := selectCluster(ctx, projID)
			if err != nil {
				return err
			}

			// Target the Cluster
			if err := targetCluster(
				ctx,
				clusterName,
				zone,
				projID,
			); err != nil {
				return err
			}

			// Install Knative Build
			Logf(ctx, "install Knative Build")
			if _, err := kubectl(
				ctx,
				"apply",
				"--filename",
				KnativeBuildYAML,
			); err != nil {
				return err
			}

			// Install Service Catalog
			if err := installServiceCatalog(ctx); err != nil {
				return err
			}

			// Install kf
			Logf(ctx, "install kf")
			if _, err := kubectl(
				ctx,
				"apply",
				"--filename",
				KfNightlyBuildYAML,
			); err != nil {
				return err
			}

			// Setup kf space
			if err := setupSpace(ctx, projID); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Display the gcloud and kubectl commands",
	)

	return cmd
}

func selectProject(ctx context.Context) (string, error) {
	ctx = setLogPrefix(ctx, "Select Project")
	projectList, err := projects(ctx)
	if err != nil {
		return "", err
	}

	projectList = append([]string{"Create New Project"}, projectList...)

	// Fetch the desired project
	idx, result, err := selectPrompt(
		ctx,
		"Select Project",
		projectList...,
	)
	if err != nil {
		return "", err
	}

	// Create project
	if idx == 0 {
		result, err = createProject(ctx)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

func selectCluster(ctx context.Context, projID string) (string, string, error) {
	ctx = setLogPrefix(ctx, "Select Cluster")
	clusterList, err := clusters(ctx, projID)
	if err != nil {
		return "", "", err
	}

	clusterList = append([]string{"Create New GKE Cluster"}, clusterList...)

	// Fetch the desired cluster
	idx, clusterName, err := selectPrompt(
		ctx,
		"Select GKE Cluster",
		clusterList...,
	)
	if err != nil {
		return "", "", err
	}

	if idx == 0 {
		// Create new cluster
		return createNewCluster(ctx, projID)
	}

	// Use existing cluster
	// We need to figure out the zone
	zone, err := clusterZone(ctx, projID, clusterName)
	return clusterName, zone, err
}

func randName(prefix string, args ...interface{}) string {
	return fmt.Sprintf(prefix, args...) + strconv.FormatInt(time.Now().UnixNano(), 36)
}

var (
	nameRegexp = regexp.MustCompile(`^[a-z][0-9a-zA-Z-]{5,29}$`)
	// hostnameRegexp is from https://stackoverflow.com/a/106223
	hostnameRegexp = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
)

func createProject(ctx context.Context) (string, error) {
	ctx = setLogPrefix(ctx, "New Project")

	for {
		name, err := namePrompt(
			ctx,
			labelColor.Sprint("Project Name (not ID): "),
			randName("kf-"),
		)
		if err != nil {
			return "", err
		}

		Logf(ctx, "creating new project %s", name)
		_, err = gcloud(ctx, "-q", "projects", "create", "--name", name)
		if err != nil {
			return "", err
		}

		Logf(ctx, "fetching project ID for %s", name)
		projects, err := gcloud(
			ctx,
			"projects",
			"list",
			"--filter", fmt.Sprintf("name:%q", name),
			"--format", "value(projectId)",
		)

		if err != nil {
			return "", err
		}

		if len(projects) != 1 {
			return "", fmt.Errorf(
				"something went wrong while trying to fetch the project ID for %s: %s",
				name,
				strings.Join(projects, "\n"),
			)
		}

		projID := projects[0]
		Logf(ctx, "Created project %s with a project ID %s", name, projID)
		return projID, nil
	}
}

func hostnamePrompt(ctx context.Context, label, def string) (string, error) {
	prompt := promptui.Prompt{
		Label: labelColor.Sprint(label),
		Validate: func(input string) error {
			if !hostnameRegexp.MatchString(input) {
				return errors.New("invalid hostname")
			}
			return nil
		},
		Default: def,
	}

	name, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return name, nil
}

func namePrompt(ctx context.Context, label, def string) (string, error) {
	prompt := promptui.Prompt{
		Label: labelColor.Sprint(label),
		Validate: func(input string) error {
			if !nameRegexp.MatchString(input) {
				return errors.New("invalid name")
			}
			return nil
		},
		Default: def,
	}

	name, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return name, nil
}

func selectPrompt(
	ctx context.Context,
	label string,
	items ...string,
) (int, string, error) {
	p := promptui.Select{
		Label:             labelColor.Sprint(label),
		StartInSearchMode: true,
		Searcher:          searcher(items),
		Items:             items,
	}
	return p.Run()
}

func selectYesNo(ctx context.Context, label string) (bool, error) {
	idx, _, err := selectPrompt(ctx, label, "yes", "no")
	if err != nil {
		return false, err
	}

	return idx == 0, nil
}

// projects uses gcloud to list all the available projects.
func projects(ctx context.Context) ([]string, error) {
	Logf(ctx, "finding your projects...")
	return gcloud(ctx, "projects", "list", "--format", "value(projectId)")
}

// clusters uses gcloud to list all the available GKE clusters for a project.
func clusters(ctx context.Context, projID string) ([]string, error) {
	Logf(ctx, "finding your GKE clusters...")
	return gcloud(
		ctx,
		"--project", projID,
		"container",
		"clusters",
		"list",
		"--format", "value(name)",
	)
}

// clusterZone uses gcloud to figure out a GKE cluster's zone.
func clusterZone(ctx context.Context, projID, clusterName string) (string, error) {
	Logf(ctx, "finding your GKE cluster's zone...")
	output, err := gcloud(
		ctx,
		"--project", projID,
		"container",
		"clusters",
		"list",
		"--format", "value(location)",
		"--filter", fmt.Sprintf("name:%q", clusterName),
	)
	if err != nil {
		return "", err
	}
	return strings.Join(output, ""), nil
}

// zones uses gcloud to list all the available zones for a project ID.
func zones(ctx context.Context, projID string) ([]string, error) {
	Logf(ctx, "Finding your zones...")
	return gcloud(
		ctx,
		"compute",
		"zones",
		"list",
		"--project", projID,
		"--format", "value(name)",
	)
}

// networks uses gcloud to list all the available networks for a project ID.
func networks(ctx context.Context, projID string) ([]string, error) {
	Logf(ctx, "Finding your networks...")
	return gcloud(
		ctx,
		"compute",
		"networks",
		"list",
		"--project", projID,
		"--format", "value(name)",
	)
}

func createNewCluster(ctx context.Context, projID string) (string, string, error) {
	ctx = setLogPrefix(ctx, "Create New GKE Config")

	// Check billing for project
	if err := ensureBilling(ctx, projID); err != nil {
		return "", "", err
	}

	// Enable required services APIs
	Logf(ctx, "enabling required service APIs")
	for _, serviceName := range []string{
		"compute.googleapis.com",
		"container.googleapis.com",
	} {
		if err := enableServiceAPI(ctx, projID, serviceName); err != nil {
			return "", "", err
		}
	}

	// Grab GKE Settings from user
	clusterName, serviceAccount, zone, network, err := gkeClusterConfig(ctx, projID)
	if err != nil {
		return "", "", err
	}

	// Create the GKE Cluster
	if err := buildGKECluster(
		ctx,
		projID,
		clusterName,
		zone,
		serviceAccount,
		network,
	); err != nil {
		return "", "", err
	}

	return clusterName, zone, nil
}

func gkeClusterConfig(
	ctx context.Context,
	projID string,
) (
	clusterName string,
	serviceAccount string,
	zone string,
	network string,
	err error,
) {
	ctx = setLogPrefix(ctx, "GKE Cluster Config")

	// ClusterName
	{
		clusterName, err = namePrompt(
			ctx,
			"Cluster Name: ",
			randName("kf-"),
		)
		if err != nil {
			return "", "", "", "", err
		}
	}

	// Service Account
	{
		serviceAccountName := randName("kf-")
		serviceAccount = fmt.Sprintf(
			"%s@%s.iam.gserviceaccount.com",
			serviceAccountName,
			projID,
		)
		ok, err := selectYesNo(
			ctx,
			fmt.Sprintf("Create service account %s?", serviceAccount),
		)
		if err != nil {
			return "", "", "", "", err
		}
		if !ok {
			return "", "", "", "", errors.New("chose to not create service account")
		}
		if err = createServiceAccount(
			ctx,
			serviceAccountName,
			serviceAccount,
			projID,
		); err != nil {
			return "", "", "", "", err
		}
	}

	// Zone
	{
		availableZones, err := zones(ctx, projID)
		switch {
		case err != nil:
			return "", "", "", "", err
		case len(availableZones) == 0:
			return "", "", "", "", errors.New("there was a listing your zones")
		case len(availableZones) == 1:
			zone = availableZones[0]
		default:
			_, zone, err = selectPrompt(ctx, "Zone", availableZones...)
			if err != nil {
				return "", "", "", "", err
			}
		}
	}

	// Network
	{
		availableNetworks, err := networks(ctx, projID)
		switch {
		case err != nil:
			return "", "", "", "", err
		case len(availableNetworks) == 0:
			// We won't use a network
		case len(availableNetworks) == 1:
			network = availableNetworks[0]
		default:
			_, network, err = selectPrompt(ctx, "Network", availableNetworks...)
			if err != nil {
				return "", "", "", "", err
			}
		}
	}

	return
}

func buildGKECluster(
	ctx context.Context,
	projID string,
	clusterName string,
	zone string,
	serviceAccount string,
	network string,
) error {
	ctx = setLogPrefix(ctx, "Create GKE Cluster")
	args := []string{
		"beta", "container", "clusters", "create", clusterName,
		"--zone", zone,
		"--no-enable-basic-auth",
		"--cluster-version", GKEClusterVersion,
		"--machine-type", MachineType,
		"--image-type", ImageType,
		"--disk-type", DiskType,
		"--disk-size", DiskSize,
		"--metadata", "disable-legacy-endpoints=true",
		"--service-account", serviceAccount,
		"--num-nodes", NumNodes,
		"--enable-stackdriver-kubernetes",
		"--enable-ip-alias",
		"--default-max-pods-per-node", DefaultMaxPodsPerNode,
		"--addons", Addons,
		"--istio-config", "auth=MTLS_PERMISSIVE",
		"--enable-autoupgrade",
		"--enable-autorepair",
		"--project", projID,
	}

	if network != "" {
		args = append(args, "--network", network)
	}

	Logf(ctx, "Creating %s GKE+CR cluster. This may take a moment", clusterName)
	_, err := gcloud(
		ctx,
		args...,
	)
	return err
}

func targetCluster(
	ctx context.Context,
	clusterName string,
	zone string,
	projID string,
) error {
	ctx = setLogPrefix(ctx, "Target Cluster")
	Logf(ctx, "targeting cluster")
	_, err := gcloud(
		ctx,
		"container",
		"clusters",
		"get-credentials",
		clusterName,
		"--zone", zone,
		"--project", projID,
	)

	return err
}

func enableServiceAPI(ctx context.Context, projID, serviceName string) error {
	ctx = setLogPrefix(ctx, "Enable Service API")
	services, err := gcloud(
		ctx,
		"services",
		"list",
		"--project", projID,
		"--filter", "name:"+serviceName,
		"--format", "value(name)",
	)
	if err != nil {
		return err
	}

	if len(services) > 0 {
		// Already enabled, move on
		return nil
	}

	ok, err := selectYesNo(ctx, fmt.Sprintf("Enable %s API?", serviceName))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%s service API not enabled", serviceName)
	}

	// Enable the service API
	Logf(ctx, "enabling %s service API. This may take a moment", serviceName)
	_, err = gcloud(
		ctx,
		"-q",
		"services",
		"--project", projID,
		"enable",
		serviceName,
	)
	if err != nil {
		return err
	}

	return nil
}

func createServiceAccount(
	ctx context.Context,
	serviceAccountName string,
	serviceAccount string,
	projID string,
) error {
	ctx = setLogPrefix(ctx, "Create Service Account")
	Logf(ctx, "Creating service account %s", serviceAccount)
	if _, err := gcloud(
		ctx,
		"iam",
		"service-accounts",
		"create",
		serviceAccountName,
		"--project", projID,
	); err != nil {
		return err
	}
	if _, err := gcloud(
		ctx,
		"projects",
		"add-iam-policy-binding",
		projID,
		"--member", "serviceAccount:"+serviceAccount,
		"--role", "roles/storage.admin",
	); err != nil {
		return err
	}

	return nil
}

func ensureBilling(ctx context.Context, projID string) error {
	ctx = setLogPrefix(ctx, "Ensure Billing")

	enabled, err := billingEnabled(ctx, projID)
	if err != nil {
		return err
	}
	if enabled {
		// Looks good, move on
		return nil
	}

	Logf(ctx, "Looks like you need to enable billing for %s", projID)
	ok, err := selectYesNo(ctx, fmt.Sprintf("Sync billing account for %s?", projID))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no billing setup for %s", projID)
	}

	availableAccounts, err := billingAccounts(ctx)
	if err != nil {
		return err
	}
	_, accountID, err := selectPrompt(ctx, "Billing Account", availableAccounts...)
	if err != nil {
		return err
	}

	if err := linkBilling(ctx, projID, accountID); err != nil {
		return err
	}
	return nil
}

// billingEnabled checks to see if a project has billing enabled
func billingEnabled(ctx context.Context, projID string) (bool, error) {
	Logf(ctx, "checking if %s has billing enabled", projID)
	result, err := gcloud(
		ctx,
		"alpha",
		"billing",
		"projects",
		"describe",
		projID,
		"--format",
		"value(billingEnabled)",
	)
	if err != nil {
		return false, err
	}

	return strings.ToLower(strings.Join(result, "")) == "true", nil
}

// billingAccounts will look to see what accounts the user has to help with
// linking them.
func billingAccounts(ctx context.Context) ([]string, error) {
	Logf(ctx, "fetching billing accounts")
	accounts, err := gcloud(
		ctx,
		"alpha",
		"billing",
		"accounts",
		"list",
		"--format",
		"value(name)",
	)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, errors.New("there has to be at least one billing account setup")
	}

	return accounts, nil
}

// linkBilling links project to billing account.
func linkBilling(ctx context.Context, projID, accountID string) error {
	Logf(ctx, "linking %s to %s", projID, accountID)
	_, err := gcloud(
		ctx,
		"alpha",
		"billing",
		"projects",
		"link",
		projID,
		"--billing-account",
		accountID,
	)
	return err
}

func command(ctx context.Context, name string, args ...string) ([]string, error) {
	if verbose, ok := ctx.Value(verboseType{}).(bool); ok && verbose {
		ctx = setLogPrefix(ctx, name)
		Logf(ctx, "%s %s", name, strings.Join(args, " "))
	}

	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		Logf(ctx, string(output))
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func installServiceCatalog(ctx context.Context) error {
	ctx = setLogPrefix(ctx, "Service Catalog")
	Logf(ctx, "installing Service Catalog")
	Logf(ctx, "downloading service catalog templates")
	tempDir, err := ioutil.TempDir("", "kf-service-catalog")
	if err != nil {
		return err
	}
	defer func() {
		Logf(ctx, "cleaning up %s", tempDir)
		os.RemoveAll(tempDir)
	}()

	tmpKfPath := path.Join(tempDir, "kf")

	if _, err := git(
		ctx,
		"clone",
		"https://github.com/google/kf",
		tmpKfPath,
	); err != nil {
		return err
	}

	Logf(ctx, "applying templates")
	if _, err := kubectl(
		ctx,
		"apply",
		"-R",
		"--filename", path.Join(tmpKfPath, "third_party/service-catalog/manifests/catalog/templates"),
	); err != nil {
		return err
	}

	return nil
}

func setupSpace(ctx context.Context, projID string) error {
	ctx = setLogPrefix(ctx, "kf setup")
	ok, err := selectYesNo(ctx, "Setup kf space?")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	Logf(ctx, "Setting up kf space")
	spaceName, err := namePrompt(ctx, "Space Name: ", randName("space-"))
	if err != nil {
		return err
	}
	domain, err := hostnamePrompt(ctx, "Domain: ", "example.com")
	if err != nil {
		return err
	}

	if _, err := kf(
		ctx,
		"create-space", spaceName,
		"--domain", domain,
		"--container-registry", fmt.Sprintf("gcr.io/%s", projID),
	); err != nil {
		return err
	}

	if _, err := kf(ctx, "target", "-s", spaceName); err != nil {
		return err
	}

	return nil
}

// gcloud will run the command and block until its done.
func gcloud(ctx context.Context, args ...string) ([]string, error) {
	return command(ctx, "gcloud", args...)
}

// kubectl will run the command and block until its done.
func kubectl(ctx context.Context, args ...string) ([]string, error) {
	return command(ctx, "kubectl", args...)
}

// kf will run the command and block until its done.
func kf(ctx context.Context, args ...string) ([]string, error) {
	return command(ctx, "kf", args...)
}

// git will run the command and block until its done.
func git(ctx context.Context, args ...string) ([]string, error) {
	return command(ctx, "git", args...)
}

type loggerType struct{}

func setContextOutput(ctx context.Context, out io.Writer) context.Context {
	return context.WithValue(ctx, loggerType{}, out)
}

type loggerPrefixType struct{}

func setLogPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, loggerPrefixType{}, prefix)
}

type verboseType struct{}

func setVerbosity(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseType{}, verbose)
}

func Logf(ctx context.Context, v string, args ...interface{}) {
	out := ctx.Value(loggerType{}).(io.Writer)

	if !strings.HasSuffix(v, "\n") {
		v += "\n"
	}

	if prefix, ok := ctx.Value(loggerPrefixType{}).(string); ok {
		v = fmt.Sprintf("[%s] %s", prefixColor.Sprint(prefix), v)
	}

	fmt.Fprintf(out, v, args...)
}

func searcher(items []string) func(input string, index int) bool {
	return func(input string, index int) bool {
		item := strings.ToLower(items[index])
		input = strings.TrimSpace(strings.ToLower(input))

		return strings.Contains(item, input)
	}
}
