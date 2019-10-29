---
title: "kf configure-space get-execution-env"
slug: kf-configure-space-get-execution-env
url: /docs/general-info/kf-cli/commands/kf-configure-space-get-execution-env/
---
## kf configure-space get-execution-env

Get the space-wide environment variables.

### Synopsis

Get the space-wide environment variables.

```
kf configure-space get-execution-env [SPACE_NAME] [flags]
```

### Examples

```
  # Configure the space "my-space"
  kf configure-space get-execution-env my-space
  # Configure the targeted space
  kf configure-space get-execution-env
```

### Options

```
  -h, --help   help for get-execution-env
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

