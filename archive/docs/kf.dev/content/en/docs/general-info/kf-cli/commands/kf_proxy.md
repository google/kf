---
title: "kf proxy"
slug: kf-proxy
url: /docs/general-info/kf-cli/commands/kf-proxy/
---
## kf proxy

Create a proxy to an app on a local port

### Synopsis

This command creates a local proxy to a remote gateway modifying the request headers to make requests route to your app.

 You can manually specify the gateway or have it autodetected based on your cluster.

```
kf proxy APP_NAME [flags]
```

### Examples

```
  kf proxy myapp
```

### Options

```
      --gateway string   HTTP gateway to route requests to (default: autodetected from cluster)
  -h, --help             help for proxy
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

