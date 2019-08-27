---
title: "kf configure-space set-default-domain"
slug: kf-configure-space-set-default-domain
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-default-domain/
---
## kf configure-space set-default-domain

Set a default domain for a space

### Synopsis

Set a default domain for a space

```
kf configure-space set-default-domain SPACE_NAME DOMAIN [flags]
```

### Examples

```
  kf configure-space set-default-domain my-space myspace.mycompany.com
```

### Options

```
  -h, --help   help for set-default-domain
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

