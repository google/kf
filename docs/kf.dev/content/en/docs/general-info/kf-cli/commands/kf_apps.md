---
title: "kf apps"
slug: kf-apps
url: /docs/general-info/kf-cli/commands/kf-apps/
---
## kf apps

List Apps in the target space

### Synopsis

List Apps in the target space

```
kf apps [flags]
```

### Examples

```
  kf apps
```

### Options

```
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for apps
  -o, --output string                 Output format. One of: go-template|go-template-file|json|jsonpath|jsonpath-file|name|template|templatefile|yaml.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
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

