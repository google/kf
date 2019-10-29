---
title: "kf configure-space append-domain"
slug: kf-configure-space-append-domain
url: /docs/general-info/kf-cli/commands/kf-configure-space-append-domain/
---
## kf configure-space append-domain

Append a domain for a space

### Synopsis

Append a domain for a space

```
kf configure-space append-domain [SPACE_NAME] DOMAIN [flags]
```

### Examples

```
  # Configure the space "my-space"
  kf configure-space append-domain my-space myspace.mycompany.com
  # Configure the targeted space
  kf configure-space append-domain myspace.mycompany.com
```

### Options

```
  -h, --help   help for append-domain
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

