---
title: "kf install gke"
slug: kf-install-gke
url: /docs/general-info/kf-cli/commands/kf-install-gke/
---
## kf install gke

Install kf on GKE with Cloud Run (Note: this will incur GCP costs)

### Synopsis

This interactive installer will walk you through the process of installing kf on GKE with Cloud Run. You MUST have gcloud and kubectl installed and available on the path.

 Note: running this will incur costs to run GKE. See https://cloud.google.com/products/calculator/ to get an estimate.

```
kf install gke [subcommand] [flags]
```

### Examples

```
  kf install gke
```

### Options

```
      --billing-account string        Configure Billing Account
      --create-gke-cluster            Configure Create-GKE Cluster
      --create-project                Configure Create-Project
      --gke-cluster string            Configure GKE Cluster
  -h, --help                          help for gke
      --network string                Configure Network
      --new-gke-cluster-name string   Configure New GKE Cluster Name
      --new-project-name string       Configure New Project Name
      --new-space-name string         Configure New Space Name
      --project string                Configure Project
  -q, --quiet                         Non-interactive mode. This will assume yes to yes-no questions
      --space-domain string           Configure Space Domain
  -v, --verbose                       Display any commands ran in the shell
      --zone string                   Configure Zone
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf install](/docs/general-info/kf-cli/commands/kf-install/)	 - Install kf

