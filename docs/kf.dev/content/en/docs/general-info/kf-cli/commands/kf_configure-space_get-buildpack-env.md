---
title: "kf configure-space get-buildpack-env"
slug: kf-configure-space-get-buildpack-env
url: /docs/general-info/kf-cli/commands/kf-configure-space-get-buildpack-env/
---
## kf configure-space get-buildpack-env

Get the environment variables for buildpack builds in a space.

### Synopsis

Get the environment variables for buildpack builds in a space.

```
kf configure-space get-buildpack-env SPACE_NAME [flags]
```

### Examples

```
  kf configure-space get-buildpack-env my-space
```

### Options

```
  -h, --help   help for get-buildpack-env
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

