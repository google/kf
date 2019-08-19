---
title: "kf spaces"
slug: kf-spaces
url: /docs/general-info/kf-cli/commands/kf-spaces/
---
## kf spaces

List all kf spaces

### Synopsis

List spaces and their statuses for the currently targeted cluster.

 The output of this command is similar to what you'd get by running:

  kubectl get spaces.kf.dev

```
kf spaces [flags]
```

### Examples

```
  kf spaces
```

### Options

```
  -h, --help   help for spaces
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

