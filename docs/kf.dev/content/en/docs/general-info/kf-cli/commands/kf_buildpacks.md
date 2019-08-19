---
title: "kf buildpacks"
slug: kf-buildpacks
url: /docs/general-info/kf-cli/commands/kf-buildpacks/
---
## kf buildpacks

List buildpacks in current builder

### Synopsis

List the buildpacks available in the space to applications being built with buildpacks.

 Buildpack support is determined by the buildpack builder image and can change from one space to the next.

```
kf buildpacks [flags]
```

### Examples

```
  kf buildpacks
```

### Options

```
  -h, --help   help for buildpacks
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

