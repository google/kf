---
title: "kf stacks"
slug: kf-stacks
url: /docs/general-info/kf-cli/commands/kf-stacks/
---
## kf stacks

List stacks available in the space

### Synopsis

List the stacks available in the space to applications being built with buildpacks.

 Stack support is determined by the buildpack builder image so they can change from one space to the next.

```
kf stacks [flags]
```

### Examples

```
  kf stacks
```

### Options

```
  -h, --help   help for stacks
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

