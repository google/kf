---
title: "kf unbind-service"
slug: kf-unbind-service
url: /docs/general-info/kf-cli/commands/kf-unbind-service/
---
## kf unbind-service

Unbind a service instance from an app

### Synopsis

Unbind removes an application's access to a service instance.

 This will delete the credential from the service broker that created the instance and update the VCAP_SERVICES environment variable for the application to remove the reference to the instance.

```
kf unbind-service APP_NAME SERVICE_INSTANCE [flags]
```

### Examples

```
  kf unbind-service myapp my-instance
```

### Options

```
      --async   Don't wait for the action to complete on the server before returning
  -h, --help    help for unbind-service
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

