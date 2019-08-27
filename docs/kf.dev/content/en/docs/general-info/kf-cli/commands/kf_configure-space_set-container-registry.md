---
title: "kf configure-space set-container-registry"
slug: kf-configure-space-set-container-registry
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-container-registry/
---
## kf configure-space set-container-registry

Set the container registry used for builds.

### Synopsis

Set the container registry used for builds.

```
kf configure-space set-container-registry SPACE_NAME REGISTRY [flags]
```

### Examples

```
  kf configure-space set-container-registry my-space gcr.io/my-project
```

### Options

```
  -h, --help   help for set-container-registry
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf configure-space](/docs/general-info/kf-cli/commands/kf-configure-space/)	 - Set configuration for a space

