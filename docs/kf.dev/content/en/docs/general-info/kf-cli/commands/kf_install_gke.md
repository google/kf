---
title: "kf install gke"
slug: kf-install-gke
url: /docs/general-info/kf-cli/commands/kf-install-gke/
---
## kf install gke

Install kf on GKE with Cloud Run (Note: this will incur GCP costs)

### Synopsis

This interactive installer will walk you through the process of installing kf on GKE with Cloud Run. You MUST have gcloud and kubectl installed and available on the path. Note: running this will incur costs to run GKE. See https://cloud.google.com/products/calculator/ to get an estimate.

 To override the GKE version that's chosen, set the environment variable GKE_VERSION.

```
kf install gke [subcommand] [flags]
```

### Examples

```
  kf install gke
```

### Options

```
      --cluster-machine-type string   Machine type the GKE cluster will be created with (default "n1-standard-4")
      --cluster-master-ip string      GKE's master Server IP to target (public|internal) (default "public")
      --cluster-name string           GKE cluster name to use
      --cluster-network string        Network the GKE cluster is in or will be created in (default "default")
      --cluster-zone string           Zone the GKE cluster is in or will be created in (default "us-central1-a")
      --create-cluster                Create a new GKE cluster
      --create-space                  Create a new space
  -h, --help                          help for gke
      --interactive                   Make the command interactive
      --kf-version string             Kf release version to use
      --project-id string             GCP project ID to use
      --space-domain string           Kf space's default domain
      --space-name string             Kf space name to create/target use
  -v, --verbose                       Make the operation more chatty
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf install](/docs/general-info/kf-cli/commands/kf-install/)	 - Install kf

