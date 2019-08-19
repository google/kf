---
title: "kf install"
slug: kf-install
url: /docs/general-info/kf-cli/commands/kf-install/
---
## kf install

Install kf

### Synopsis

Installs kf into a new Kubernetes cluster, optionally creating the cluster.

 WARNING: No checks are done on a cluster before installing a new version of kf. This means that if you target a cluster with a later version of kf then you can downgrade the system.

```
kf install [subcommand] [flags]
```

### Options

```
  -h, --help   help for install
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience
* [kf install gke](/docs/general-info/kf-cli/commands/kf-install-gke/)	 - Install kf on GKE with Cloud Run (Note: this will incur GCP costs)

