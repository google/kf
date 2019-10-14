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

package gke

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/kf/pkg/kf/commands/install/kf"
	. "github.com/google/kf/pkg/kf/commands/install/util"
	"github.com/spf13/cobra"
)

const (
	GKEVersionEnvVar = "GKE_VERSION"

	machineType           = "n1-standard-4"
	imageType             = "COS"
	diskType              = "pd-standard"
	diskSize              = "100"
	numNodes              = "3"
	defaultMaxPodsPerNode = "110"
	addons                = "HorizontalPodAutoscaling,HttpLoadBalancing,Istio,CloudRun"
	defaultGKEVersion     = "1.13.7-gke.24"
)

// NewGKECommand creates a command that can install kf to GKE+Cloud Run.
// TODO: This installer is using gcloud and kubectl under the hood. We should
// really replace gcloud with google.golang.org/api and kubectl with
// k8s.io/client-go
func NewGKECommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:     "gke [subcommand]",
		Short:   "Install kf on GKE with Cloud Run (Note: this will incur GCP costs)",
		Example: "kf install gke",
		Long: `
			This interactive installer will walk you through the process of installing kf
			on GKE with Cloud Run. You MUST have gcloud and kubectl installed and
			available on the path. Note: running this will incur costs to run GKE. See
			https://cloud.google.com/products/calculator/ to get an estimate.

			To override the GKE version that's chosen, set the environment variable
			GKE_VERSION.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			// Print kubectl version
			ctx := SetContextOutput(context.Background(), cmd.ErrOrStderr())
			version, err := Kubectl(ctx, "version", "--short", "--client")
			if err != nil {
				return err
			}
			fmt.Fprintf(
				cmd.ErrOrStderr(),
				"%s\n%s\n",
				LabelColor.Sprint("kubectl version:"),
				strings.Join(version, "\n"),
			)

			// Print gcloud version
			version, err = gcloud(ctx, "version")
			if err != nil {
				return err
			}
			fmt.Fprintf(
				cmd.ErrOrStderr(),
				"%s\n%s\n",
				LabelColor.Sprint("gcloud version:"),
				strings.Join(version, "\n"),
			)

			// Ensure gcloud alpha is available
			// This is necessary because the user my have NEVER ran a `gcloud
			// alpha` (or `beta`) command before and therefore their first one
			// will ask if they really want to.
			for _, v := range []string{"alpha", "beta"} {
				// We don't care about the output if it succeeds.
				_, err := gcloud(ctx, v, "help")
				if err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx := SetContextOutput(context.Background(), cmd.ErrOrStderr())
			ctx = SetLogPrefix(ctx, "Install GKE+CR+Kf")
			ctx = SetVerbosity(ctx, verbose)

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

			// Install kf
			return kf.Install(ctx, fmt.Sprintf("gcr.io/%s", projID))
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
	ctx = SetLogPrefix(ctx, "Select Project")
	projectList, err := projects(ctx)
	if err != nil {
		return "", err
	}

	projectList = append([]string{"Create New Project"}, projectList...)

	// Fetch the desired project
	idx, result, err := SelectPrompt(
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
	ctx = SetLogPrefix(ctx, "Select Cluster")
	clusterList, err := clusters(ctx, projID)
	if err != nil {
		return "", "", err
	}

	clusterList = append([]string{"Create New GKE Cluster"}, clusterList...)

	// Fetch the desired cluster
	idx, clusterName, err := SelectPrompt(
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

func createProject(ctx context.Context) (string, error) {
	ctx = SetLogPrefix(ctx, "New Project")

	for {
		name, err := NamePrompt(
			ctx,
			LabelColor.Sprint("Project Name (not ID): "),
			RandName("kf-"),
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
	ctx = SetLogPrefix(ctx, "Create New GKE Config")

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
	gkeCfg, err := gkeClusterConfig(ctx, projID)
	if err != nil {
		return "", "", err
	}

	// Create the GKE Cluster
	if err := buildGKECluster(ctx, projID, gkeCfg); err != nil {
		return "", "", err
	}

	return gkeCfg.clusterName, gkeCfg.zone, nil
}

type gkeConfig struct {
	clusterName    string
	serviceAccount string
	zone           string
	network        string
}

func gkeClusterConfig(
	ctx context.Context,
	projID string,
) (
	cfg gkeConfig,
	err error,
) {
	ctx = SetLogPrefix(ctx, "GKE Cluster Config")

	// ClusterName
	{
		cfg.clusterName, err = NamePrompt(
			ctx,
			"Cluster Name: ",
			RandName("kf-"),
		)
		if err != nil {
			return gkeConfig{}, err
		}
	}

	// Service Account
	{
		serviceAccountName := RandName("kf-")
		cfg.serviceAccount = fmt.Sprintf(
			"%s@%s.iam.gserviceaccount.com",
			serviceAccountName,
			projID,
		)
		ok, err := SelectYesNo(
			ctx,
			fmt.Sprintf("Create service account %s?", cfg.serviceAccount),
		)
		if err != nil {
			return gkeConfig{}, err
		}
		if !ok {
			return gkeConfig{}, errors.New("chose to not create service account")
		}
		if err = createServiceAccount(
			ctx,
			serviceAccountName,
			cfg.serviceAccount,
			projID,
		); err != nil {
			return gkeConfig{}, err
		}
	}

	// Zone
	{
		availableZones, err := zones(ctx, projID)
		switch {
		case err != nil:
			return gkeConfig{}, err
		case len(availableZones) == 0:
			return gkeConfig{}, errors.New("there was a listing your zones")
		case len(availableZones) == 1:
			cfg.zone = availableZones[0]
		default:
			_, cfg.zone, err = SelectPrompt(ctx, "Zone", availableZones...)
			if err != nil {
				return gkeConfig{}, err
			}
		}
	}

	// Network
	{
		availableNetworks, err := networks(ctx, projID)
		switch {
		case err != nil:
			return gkeConfig{}, err
		case len(availableNetworks) == 0:
			// We won't use a network
		case len(availableNetworks) == 1:
			cfg.network = availableNetworks[0]
		default:
			_, cfg.network, err = SelectPrompt(ctx, "Network", availableNetworks...)
			if err != nil {
				return gkeConfig{}, err
			}
		}
	}

	return
}

func buildGKECluster(
	ctx context.Context,
	projID string,
	gkeCfg gkeConfig,
) error {
	ctx = SetLogPrefix(ctx, "Create GKE Cluster")
	args := []string{
		"beta", "container", "clusters", "create", gkeCfg.clusterName,
		"--zone", gkeCfg.zone,
		"--no-enable-basic-auth",
		"--machine-type", machineType,
		"--image-type", imageType,
		"--disk-type", diskType,
		"--disk-size", diskSize,
		"--metadata", "disable-legacy-endpoints=true",
		"--service-account", gkeCfg.serviceAccount,
		"--num-nodes", numNodes,
		"--enable-stackdriver-kubernetes",
		"--enable-ip-alias",
		"--default-max-pods-per-node", defaultMaxPodsPerNode,
		"--addons", addons,
		"--istio-config", "auth=MTLS_PERMISSIVE",
		"--enable-autoupgrade",
		"--enable-autorepair",
		"--project", projID,
	}

	if gkeClusterVersion, ok := os.LookupEnv(GKEVersionEnvVar); ok {
		args = append(args, "--cluster-version", gkeClusterVersion)
	} else {
		Logf(ctx, "Setting the GKE version to be %q, override it by setting the environment variable %s", defaultGKEVersion, GKEVersionEnvVar)
		args = append(args, "--cluster-version", defaultGKEVersion)

	}

	if gkeCfg.network != "" {
		args = append(args, "--network", gkeCfg.network)
	}

	Logf(ctx, "Creating %s GKE cluster with Cloud Run. This may take a moment", gkeCfg.clusterName)
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
	ctx = SetLogPrefix(ctx, "Target Cluster")
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
	ctx = SetLogPrefix(ctx, "Enable Service API")
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

	ok, err := SelectYesNo(ctx, fmt.Sprintf("Enable %s API?", serviceName))
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
	ctx = SetLogPrefix(ctx, "Create Service Account")
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
	ctx = SetLogPrefix(ctx, "Ensure Billing")

	enabled, err := billingEnabled(ctx, projID)
	if err != nil {
		return err
	}
	if enabled {
		// Looks good, move on
		return nil
	}

	Logf(ctx, "Looks like you need to enable billing for %s", projID)
	ok, err := SelectYesNo(ctx, fmt.Sprintf("Sync billing account for %s?", projID))
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
	_, accountID, err := SelectPrompt(ctx, "Billing Account", availableAccounts...)
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

// gcloud will run the command and block until its done.
func gcloud(ctx context.Context, args ...string) ([]string, error) {
	return Command(ctx, "gcloud", args...)
}
