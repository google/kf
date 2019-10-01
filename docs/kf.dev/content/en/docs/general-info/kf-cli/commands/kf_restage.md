---
title: "kf restage"
slug: kf-restage
url: /docs/general-info/kf-cli/commands/kf-restage/
---
## kf restage

Rebuild and deploy using the last uploaded source code and current buildpacks

### Synopsis

Rebuild and deploy using the last uploaded source code and current buildpacks

```
kf restage APP_NAME [flags]
```

### Examples

```
  kf restage myapp
```

### Options

```
      --async   Don't wait for the action to complete on the server before returning
  -h, --help    help for restage
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

