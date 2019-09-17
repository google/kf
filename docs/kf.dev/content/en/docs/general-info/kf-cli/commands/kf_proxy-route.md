---
title: "kf proxy-route"
slug: kf-proxy-route
url: /docs/general-info/kf-cli/commands/kf-proxy-route/
---
## kf proxy-route

Create a proxy to a route on a local port

### Synopsis

This command creates a local proxy to a remote gateway modifying the request headers to make requests with the host set as the specified route.

 You can manually specify the gateway or have it autodetected based on your cluster.

```
kf proxy-route ROUTE [flags]
```

### Examples

```
  kf proxy-route myhost.example.com
```

### Options

```
      --gateway string   HTTP gateway to route requests to (default: autodetected from cluster)
  -h, --help             help for proxy-route
      --port int         Local port to listen on (default 8080)
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

