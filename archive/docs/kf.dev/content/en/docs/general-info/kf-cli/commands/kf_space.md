---
title: "kf space"
slug: kf-space
url: /docs/general-info/kf-cli/commands/kf-space/
---
## kf space

Show space info

### Synopsis

Get detailed information about a specific space and its configuration.

 The output of this command is similar to what you'd get by running:

  kubectl describe space.kf.dev SPACE

```
kf space SPACE [flags]
```

### Examples

```
  kf space my-space
```

### Options

```
  -h, --help   help for space
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

