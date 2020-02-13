---
title: "kf configure-space set-buildpack-builder"
slug: kf-configure-space-set-buildpack-builder
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-buildpack-builder/
---
## kf configure-space set-buildpack-builder

Set the buildpack builder image.

### Synopsis

Set the buildpack builder image.

```
kf configure-space set-buildpack-builder [SPACE_NAME] BUILDER_IMAGE [flags]
```

### Examples

```
  # Configure the space "my-space"
  kf configure-space set-buildpack-builder my-space gcr.io/my-project/builder:latest
  # Configure the targeted space
  kf configure-space set-buildpack-builder gcr.io/my-project/builder:latest
```

### Options

```
  -h, --help   help for set-buildpack-builder
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

