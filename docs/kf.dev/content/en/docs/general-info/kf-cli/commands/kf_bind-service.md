---
title: "kf bind-service"
slug: kf-bind-service
url: /docs/general-info/kf-cli/commands/kf-bind-service/
---
## kf bind-service

Bind a service instance to an app

### Synopsis

Bind a service instance to an app

```
kf bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME] [flags]
```

### Examples

```
  kf bind-service myapp mydb -c '{"permissions":"read-only"}'
```

### Options

```
      --async                 Don't wait for the action to complete on the server before returning
  -b, --binding-name string   Name to expose service instance to app process with (default: service instance name)
  -h, --help                  help for bind-service
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

