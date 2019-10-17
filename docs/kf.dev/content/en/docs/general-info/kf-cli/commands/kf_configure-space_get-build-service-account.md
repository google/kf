---
title: "kf configure-space get-build-service-account"
slug: kf-configure-space-get-build-service-account
url: /docs/general-info/kf-cli/commands/kf-configure-space-get-build-service-account/
---
## kf configure-space get-build-service-account

Get the service account that is used when building containers in the space.

### Synopsis

Get the service account that is used when building containers in the space.

```
kf configure-space get-build-service-account SPACE_NAME [flags]
```

### Examples

```
  kf configure-space get-build-service-account my-space
```

### Options

```
  -h, --help   help for get-build-service-account
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

