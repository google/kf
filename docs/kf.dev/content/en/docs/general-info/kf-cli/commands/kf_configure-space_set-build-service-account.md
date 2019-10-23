---
title: "kf configure-space set-build-service-account"
slug: kf-configure-space-set-build-service-account
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-build-service-account/
---
## kf configure-space set-build-service-account

Set the service account to use when building containers

### Synopsis

Set the service account to use when building containers

```
kf configure-space set-build-service-account [SPACE_NAME] SERVICE_ACCOUNT [flags]
```

### Examples

```
  # Configure the space "my-space"
  kf configure-space set-build-service-account my-space myserviceaccount
  # Configure the targeted space
  kf configure-space set-build-service-account myserviceaccount
```

### Options

```
  -h, --help   help for set-build-service-account
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

