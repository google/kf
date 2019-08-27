---
title: "kf configure-space set-buildpack-env"
slug: kf-configure-space-set-buildpack-env
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-buildpack-env/
---
## kf configure-space set-buildpack-env

Set an environment variable for buildpack builds in a space.

### Synopsis

Set an environment variable for buildpack builds in a space.

```
kf configure-space set-buildpack-env SPACE_NAME ENV_VAR_NAME ENV_VAR_VALUE [flags]
```

### Examples

```
  kf configure-space set-buildpack-env my-space JDK_VERSION 11
```

### Options

```
  -h, --help   help for set-buildpack-env
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

