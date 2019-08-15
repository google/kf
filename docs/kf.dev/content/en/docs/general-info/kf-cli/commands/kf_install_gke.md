---
title: "kf install gke"
slug: kf-install-gke
url: /docs/general-info/kf-cli/commands/kf-install-gke/
---
## kf install gke

Install kf on GKE with Cloud Run (Note: this will incur GCP costs)

### Synopsis


This interactive installer will walk you through the process of installing kf
on GKE with Cloud Run. You MUST have gcloud and kubectl installed and
available on the path. Note: running this will incur costs to run GKE. See
https://cloud.google.com/products/calculator/ to get an estimate.

```
kf install gke [subcommand] [flags]
```

### Options

```
  -h, --help      help for gke
  -v, --verbose   Display the gcloud and kubectl commands
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf install](/docs/general-info/kf-cli/commands/kf-install/)	 - Install kf

