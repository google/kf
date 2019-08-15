---
title: "kf restage"
slug: kf-restage
url: /docs/general-info/kf-cli/commands/kf-restage/
---
## kf restage

Restage creates a new container using the given source code and current buildpacks

### Synopsis

Restage creates a new container using the given source code and current buildpacks

```
kf restage APP_NAME [flags]
```

### Examples

```

  kf restage myapp
  
```

### Options

```
  -h, --help   help for restage
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

