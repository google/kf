---
title: "kf env"
slug: kf-env
url: /docs/general-info/kf-cli/commands/kf-env/
---
## kf env

List the names and values of the environment variables for an app

### Synopsis

The env command gets the names and values of developer managed environment variables for an application.

 This command does not include environment variables that are set by kf such as VCAP_SERVICES or set by operators for all apps on the space.

```
kf env APP_NAME [flags]
```

### Examples

```
  kf env myapp
```

### Options

```
  -h, --help   help for env
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

