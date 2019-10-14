---
title: "kf delete"
slug: kf-delete
url: /docs/general-info/kf-cli/commands/kf-delete/
---
## kf delete

Delete an existing app

### Synopsis

This command deletes an application from kf.

 Things that won't be deleted:

  *  source code
  *  application images
  *  routes
  *  service instances

 Things that will be deleted:

  *  builds
  *  bindings

 The delete occurs asynchronously. Apps are often deleted shortly after the delete command is called, but may live on for a while if:

  *  there are still connections waiting to be served
  *  bindings fail to deprovision
  *  the cluster is in an unhealthy state

```
kf delete APP_NAME [flags]
```

### Examples

```
  kf delete myapp
```

### Options

```
      --async   Don't wait for the action to complete on the server before returning
  -h, --help    help for delete
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

