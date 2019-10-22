---
title: "kf create-space"
slug: kf-create-space
url: /docs/general-info/kf-cli/commands/kf-create-space/
---
## kf create-space

Create a space

### Synopsis

Create a space

```
kf create-space SPACE [flags]
```

### Examples

```
  kf create-space my-space --container-registry gcr.io/my-project --domain myspace.example.com --build-service-account myserviceaccount
```

### Options

```
      --build-service-account string   Service account that the build pipeline will use to build containers.
      --container-registry string      Container registry built apps and sources will be stored in.
      --domain stringArray             Sets the valid domains for the space. The first provided domain will be the default.
  -h, --help                           help for create-space
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

