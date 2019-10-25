---
title: "kf configure-space remove-domain"
slug: kf-configure-space-remove-domain
url: /docs/general-info/kf-cli/commands/kf-configure-space-remove-domain/
---
## kf configure-space remove-domain

Remove a domain from a space

### Synopsis

Remove a domain from a space

```
kf configure-space remove-domain [SPACE_NAME] DOMAIN [flags]
```

### Examples

```
  # Configure the space "my-space"
  kf configure-space remove-domain my-space myspace.mycompany.com
  # Configure the targeted space
  kf configure-space remove-domain myspace.mycompany.com
```

### Options

```
  -h, --help   help for remove-domain
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

