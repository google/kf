---
title: "kf delete-space"
slug: kf-delete-space
url: /docs/general-info/kf-cli/commands/kf-delete-space/
---
## kf delete-space

Delete a space

### Synopsis

Delete a space and all its contents.

 This will delete a space's:

  *  Apps
  *  Service bindings
  *  Service instances
  *  RBAC roles
  *  Routes
  *  The backing Kubernetes namespace
  *  Anything else in that namespace

 NOTE: Space deletion is asynchronous and may take a long time to complete depending on the number of items in the space.

 You will be unable to make changes to resources in the space once deletion has begun.

```
kf delete-space SPACE [flags]
```

### Examples

```
  kf delete-space my-space
```

### Options

```
  -h, --help   help for delete-space
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

