---
title: "kf proxy"
slug: kf-proxy
url: /docs/general-info/kf-cli/commands/kf-proxy/
---
## kf proxy

Creates a proxy to an app on a local port

### Synopsis


	This command creates a local proxy to a remote gateway modifying the request
	headers to make requests route to your app.

	You can manually specify the gateway or have it autodetected based on your
	cluster.

```
kf proxy APP_NAME [flags]
```

### Examples

```
  kf proxy myapp
```

### Options

```
      --gateway string   the HTTP gateway to route requests to, if unset it will be autodetected
  -h, --help             help for proxy
      --port int         the local port to attach to (default 8080)
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

