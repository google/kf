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

	"github.com/google/kf/pkg/kf/cli"
	"github.com/google/kf/pkg/kf/commands/install/kf"
	"github.com/google/kf/pkg/kf/commands/install/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	GKEVersionEnvVar = "GKE_VERSION"

	imageType             = "COS"
	diskType              = "pd-standard"
	diskSize              = "100"
	numNodes              = "3"
	defaultMaxPodsPerNode = "110"
	addons                = "HorizontalPodAutoscaling,HttpLoadBalancing,Istio,CloudRun"
	defaultGKEVersion     = "1.13.11-gke.11"
)

// NewGKECommand creates a command that can install kf to GKE+Cloud Run.
// TODO: This installer is using gcloud and kubectl under the hood. We should
// really replace gcloud with google.golang.org/api and kubectl with
// k8s.io/client-go
func NewGKECommand() *cobra.Command {
	cmd := cli.NewInteractiveCommand(
		cli.CommandInfo{
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
		},
		buildGraph(),
	)
	cmd.PreRunE = preRunE
	return cmd
}

func buildGraph() *cli.InteractiveNode {
	var (
		projectIN cli.InteractiveNode
		clusterIN cli.InteractiveNode
	)

	setupProject := func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
		var projectID string
		flags.StringVar(&projectID, "project-id", "", "GCP project ID to use")

		return []*cli.InteractiveNode{&clusterIN},
			func(
				ctx context.Context,
				cmd *cobra.Command,
				args []string,
			) (context.Context, *cli.InteractiveNode, error) {
				ctx = cli.SetLogPrefix(ctx, "Setup GCP Project")
				if projectID != "" {
					// ProjectID was provided via a flag.
					return setProjectID(ctx, projectID), &clusterIN, nil
				}

				// ProjectID was not provided via a flag, fetch it from the
				// user.
				var err error
				projectID, err = selectProject(ctx)
				return setProjectID(ctx, projectID), &clusterIN, err
			}
	}

	kfInstallGraph := kf.NewInstallGraph()

	setupCluster := func(flags *pflag.FlagSet) ([]*cli.InteractiveNode, cli.Runner) {
		var (
			createCluster bool
			gkeCfg        gkeConfig
			masterIP      string
		)
		flags.StringVar(&gkeCfg.clusterName, "cluster-name", "", "GKE cluster name to use")
		flags.StringVar(&gkeCfg.zone, "cluster-zone", "us-central1-a", "Zone the GKE cluster is in or will be created in")
		flags.StringVar(&gkeCfg.network, "cluster-network", "default", "Network the GKE cluster is in or will be created in")
		flags.StringVar(&gkeCfg.machineType, "cluster-machine-type", "n1-standard-4", "Machine type the GKE cluster will be created with")
		flags.StringVar(&masterIP, "cluster-master-ip", "public", "GKE's master Server IP to target (public|internal)")
		flags.BoolVar(&createCluster, "create-cluster", false, "Create a new GKE cluster")

		return []*cli.InteractiveNode{kfInstallGraph},
			func(
				ctx context.Context,
				cmd *cobra.Command,
				args []string,
			) (context.Context, *cli.InteractiveNode, error) {
				cli.ClearDefaultsForInteractive(ctx, cmd.Flags())
				ctx = cli.SetLogPrefix(ctx, "Setup GKE Cluster")
				var err error
				projectID := getProjectID(ctx)

				switch {
				case createCluster:
					// Create a new cluster (and maybe use the provided name).
					gkeCfg, err = createNewCluster(ctx, projectID, gkeCfg)
				case gkeCfg.clusterName != "":
					// Don't create a cluster, use the provided name.
					gkeCfg, err = clusterZone(ctx, projectID, gkeCfg)
				default:
					// We don't have any information, fetch all of it from the
					// user.
					gkeCfg, err = selectCluster(ctx, projectID, gkeCfg)
				}

				if err != nil {
					return ctx, nil, err
				}

				// Target the cluster
				if err := targetCluster(ctx, projectID, masterIP, gkeCfg); err != nil {
					return nil, nil, err
				}

				return kf.SetContainerRegistry(ctx, "gcr.io/"+projectID), kfInstallGraph, nil
			}
	}

	projectIN.Setup = setupProject
	clusterIN.Setup = setupCluster

	return &projectIN
}

func preRunE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Print kubectl version
	ctx := cli.SetContextOutput(context.Background(), cmd.ErrOrStderr())
	version, err := util.Kubectl(ctx, "version", "--short", "--client")
	if err != nil {
		return err
	}
	fmt.Fprintf(
		cmd.ErrOrStderr(),
		"%s\n%s\n",
		cli.LabelColor.Sprint("kubectl version:"),
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
		cli.LabelColor.Sprint("gcloud version:"),
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
}

func selectProject(ctx context.Context) (string, error) {
	ctx = cli.SetLogPrefix(ctx, "Select Project")
	projectList, err := projects(ctx)
	if err != nil {
		return "", err
	}

	projectList = append([]string{"Create New Project"}, projectList...)

	// Fetch the desired project
	idx, result, err := cli.SelectPrompt(
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

func selectCluster(ctx context.Context, projID string, cfg gkeConfig) (gkeConfig, error) {
	clusterList, err := clusters(ctx, projID)
	if err != nil {
		return gkeConfig{}, err
	}

	clusterList = append([]string{"Create New GKE Cluster"}, clusterList...)

	// Fetch the desired cluster
	idx, clusterName, err := cli.SelectPrompt(
		ctx,
		"Select GKE Cluster",
		clusterList...,
	)
	if err != nil {
		return gkeConfig{}, err
	}

	if idx == 0 {
		// Create new cluster
		return createNewCluster(ctx, projID, cfg)
	}
	cfg.clusterName = clusterName

	// Use existing cluster
	// We need to figure out the zone
	cfg, err = clusterZone(ctx, projID, cfg)
	return cfg, err
}

func createProject(ctx context.Context) (string, error) {
	ctx = cli.SetLogPrefix(ctx, "New Project")

	for {
		name, err := cli.NamePrompt(
			ctx,
			cli.LabelColor.Sprint("Project Name (not ID): "),
			cli.RandName("kf-"),
		)
		if err != nil {
			return "", err
		}

		cli.Logf(ctx, "creating new project %s", name)
		_, err = gcloud(ctx, "-q", "projects", "create", "--name", name)
		if err != nil {
			return "", err
		}

		cli.Logf(ctx, "fetching project ID for %s", name)
		projects, err := gcloud(
			ctx,
			"projects",
			"list",
			"--filter", fmt.Sprintf("name~^%s$", name),
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
		cli.Logf(ctx, "Created project %s with a project ID %s", name, projID)
		return projID, nil
	}
}

// projects uses gcloud to list all the available projects.
func projects(ctx context.Context) ([]string, error) {
	cli.Logf(ctx, "finding your projects...")
	return gcloud(ctx, "projects", "list", "--format", "value(projectId)")
}

// clusters uses gcloud to list all the available GKE clusters for a project.
func clusters(ctx context.Context, projID string) ([]string, error) {
	cli.Logf(ctx, "finding your GKE clusters...")
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
func clusterZone(ctx context.Context, projID string, cfg gkeConfig) (gkeConfig, error) {
	if cfg.zone != "" {
		return cfg, nil
	}

	cli.Logf(ctx, "finding your GKE cluster's zone...")
	output, err := gcloud(
		ctx,
		"--project", projID,
		"container",
		"clusters",
		"list",
		"--format", "value(location)",
		"--filter", fmt.Sprintf("name~^%s$", cfg.clusterName),
	)
	if err != nil {
		return gkeConfig{}, err
	}
	cfg.zone = strings.Join(output, "")
	return cfg, nil
}

// zones uses gcloud to list all the available zones for a project ID.
func zones(ctx context.Context, projID string) ([]string, error) {
	cli.Logf(ctx, "Finding your zones...")
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
	cli.Logf(ctx, "Finding your networks...")
	return gcloud(
		ctx,
		"compute",
		"networks",
		"list",
		"--project", projID,
		"--format", "value(name)",
	)
}

// machineTypes uses gcloud to list all the available machine for a project ID and zone.
func machineTypes(ctx context.Context, projID, zone string) ([]string, error) {
	cli.Logf(ctx, "Finding your machine types...")
	return gcloud(
		ctx,
		"compute",
		"machine-types",
		"list",
		"--project", projID,
		"--zones", zone,
		"--format", "value(name)",
	)
}

func createNewCluster(ctx context.Context, projID string, cfg gkeConfig) (gkeConfig, error) {
	ctx = cli.SetLogPrefix(ctx, "Create New GKE Config")

	// Check billing for project
	if err := ensureBilling(ctx, projID); err != nil {
		return gkeConfig{}, err
	}

	// Enable required services APIs
	cli.Logf(ctx, "enabling required service APIs")
	for _, serviceName := range []string{
		"compute.googleapis.com",
		"container.googleapis.com",
	} {
		if err := enableServiceAPI(ctx, projID, serviceName); err != nil {
			return gkeConfig{}, err
		}
	}

	// Grab GKE Settings from user
	var err error
	cfg, err = gkeClusterConfig(ctx, projID, cfg)
	if err != nil {
		return gkeConfig{}, err
	}

	// Create the GKE Cluster
	if err := buildGKECluster(ctx, projID, cfg); err != nil {
		return gkeConfig{}, err
	}

	return cfg, nil
}

type gkeConfig struct {
	clusterName    string
	serviceAccount string
	zone           string
	network        string
	machineType    string
}

func gkeClusterConfig(
	ctx context.Context,
	projID string,
	cfg gkeConfig,
) (
	gkeConfig,
	error,
) {
	var err error
	ctx = cli.SetLogPrefix(ctx, "GKE Cluster Config")

	// ClusterName
	if cfg.clusterName == "" {
		cfg.clusterName, err = cli.NamePrompt(
			ctx,
			"Cluster Name: ",
			cli.RandName("kf-"),
		)
		if err != nil {
			return gkeConfig{}, err
		}
	}

	// Service Account
	{
		serviceAccountName := cli.RandName("kf-")
		cfg.serviceAccount = fmt.Sprintf(
			"%s@%s.iam.gserviceaccount.com",
			serviceAccountName,
			projID,
		)
		ok, err := cli.SelectYesNo(
			ctx,
			fmt.Sprintf("Create service account %s?", cfg.serviceAccount),
			true,
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
	if cfg.zone == "" {
		availableZones, err := zones(ctx, projID)
		switch {
		case err != nil:
			return gkeConfig{}, err
		case len(availableZones) == 0:
			return gkeConfig{}, errors.New("there was a listing your zones")
		case len(availableZones) == 1:
			cfg.zone = availableZones[0]
		default:
			_, cfg.zone, err = cli.SelectPrompt(ctx, "Zone", availableZones...)
			if err != nil {
				return gkeConfig{}, err
			}
		}
	}

	// Network
	if cfg.network == "" {
		availableNetworks, err := networks(ctx, projID)
		switch {
		case err != nil:
			return gkeConfig{}, err
		case len(availableNetworks) == 0:
			// We won't use a network
		case len(availableNetworks) == 1:
			cfg.network = availableNetworks[0]
		default:
			_, cfg.network, err = cli.SelectPrompt(ctx, "Network", availableNetworks...)
			if err != nil {
				return gkeConfig{}, err
			}
		}
	}

	// Machine Type
	if cfg.machineType == "" {
		availableMachineTypes, err := machineTypes(ctx, projID, cfg.zone)
		switch {
		case err != nil:
			return gkeConfig{}, err
		case len(availableMachineTypes) == 1:
			cfg.machineType = availableMachineTypes[0]
		default:
			_, cfg.machineType, err = cli.SelectPrompt(ctx, "Machine Type (minimum recommended: 'n1-standard-4')", availableMachineTypes...)
			if err != nil {
				return gkeConfig{}, err
			}
		}
	}

	return cfg, nil
}

func buildGKECluster(
	ctx context.Context,
	projID string,
	gkeCfg gkeConfig,
) error {
	ctx = cli.SetLogPrefix(ctx, "Create GKE Cluster")
	args := []string{
		"beta", "container", "clusters", "create", gkeCfg.clusterName,
		"--zone", gkeCfg.zone,
		"--no-enable-basic-auth",
		"--machine-type", gkeCfg.machineType,
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
		cli.Logf(ctx, "Setting the GKE version to be %q, override it by setting the environment variable %s", defaultGKEVersion, GKEVersionEnvVar)
		args = append(args, "--cluster-version", defaultGKEVersion)

	}

	if gkeCfg.network != "" {
		args = append(args, "--network", gkeCfg.network)
	}

	cli.Logf(ctx, "Creating %s GKE cluster with Cloud Run. This may take a moment", gkeCfg.clusterName)
	_, err := gcloud(
		ctx,
		args...,
	)
	return err
}

func targetCluster(
	ctx context.Context,
	projectID string,
	masterIP string,
	cfg gkeConfig,
) error {
	ctx = cli.SetLogPrefix(ctx, "Target Cluster")
	cli.Logf(ctx, "targeting cluster")

	args := []string{
		"container",
		"clusters",
		"get-credentials",
		cfg.clusterName,
		"--zone", cfg.zone,
		"--project", projectID,
	}

	switch strings.ToLower(masterIP) {
	case "internal":
		args = append(args, "--internal-ip")
	case "public":
		// NOP
	case "":
		// Unset by user via flag
		_, selection, err := cli.SelectPrompt(ctx, "GKE Master Server IP", "public", "internal")
		if err != nil {
			return err
		}

		if selection == "internal" {
			args = append(args, "--internal-ip")
		}
	default:
		// Invalid selection
		return fmt.Errorf("invalid --cluster-master-ip select. Must be 'public' or 'internal'")
	}

	_, err := gcloud(ctx, args...)

	return err
}

func enableServiceAPI(ctx context.Context, projID, serviceName string) error {
	ctx = cli.SetLogPrefix(ctx, "Enable Service API")
	services, err := gcloud(
		ctx,
		"services",
		"list",
		"--project", projID,
		"--filter", fmt.Sprintf("name~^%s$", serviceName),
		"--format", "value(name)",
	)
	if err != nil {
		return err
	}

	if len(services) > 0 {
		// Already enabled, move on
		return nil
	}

	ok, err := cli.SelectYesNo(ctx, fmt.Sprintf("Enable %s API?", serviceName), true)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%s service API not enabled", serviceName)
	}

	// Enable the service API
	cli.Logf(ctx, "enabling %s service API. This may take a moment", serviceName)
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
	ctx = cli.SetLogPrefix(ctx, "Create Service Account")
	cli.Logf(ctx, "Creating service account %s", serviceAccount)
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
	ctx = cli.SetLogPrefix(ctx, "Ensure Billing")

	enabled, err := billingEnabled(ctx, projID)
	if err != nil {
		return err
	}
	if enabled {
		// Looks good, move on
		return nil
	}

	cli.Logf(ctx, "Looks like you need to enable billing for %s", projID)
	ok, err := cli.SelectYesNo(ctx, fmt.Sprintf("Sync billing account for %s?", projID), true)
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
	_, accountID, err := cli.SelectPrompt(ctx, "Billing Account", availableAccounts...)
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
	cli.Logf(ctx, "checking if %s has billing enabled", projID)
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
	cli.Logf(ctx, "fetching billing accounts")
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
	cli.Logf(ctx, "linking %s to %s", projID, accountID)
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
	return util.Command(ctx, "gcloud", args...)
}

type projectIDType struct{}

func setProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDType{}, projectID)
}

func getProjectID(ctx context.Context) string {
	projectID, _ := ctx.Value(projectIDType{}).(string)
	return projectID
}
