---
title: "kf configure-space"
slug: kf-configure-space
url: /docs/general-info/kf-cli/commands/kf-configure-space/
---
## kf configure-space

Set configuration for a space

### Synopsis

The configure-space sub-command allows operators to configure individual fields on a space.

 In Kf, almost all configuration is at the space level as opposed to being globally set on the cluster.

 NOTE: The space is queued for reconciliation every time changes are made via this command. If you want to configure spaces in automation it's better to use kubectl.

```
kf configure-space [subcommand] [flags]
```

### Options

```
  -h, --help   help for configure-space
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
* [kf configure-space append-domain](/docs/general-info/kf-cli/commands/kf-configure-space-append-domain/)	 - Append a domain for a space
* [kf configure-space delete-quota](/docs/general-info/kf-cli/commands/kf-configure-space-delete-quota/)	 - Remove all quotas for the space
* [kf configure-space get-build-service-account](/docs/general-info/kf-cli/commands/kf-configure-space-get-build-service-account/)	 - Get the service account that is used when building containers in the space.
* [kf configure-space get-buildpack-builder](/docs/general-info/kf-cli/commands/kf-configure-space-get-buildpack-builder/)	 - Get the buildpack builder used for builds.
* [kf configure-space get-buildpack-env](/docs/general-info/kf-cli/commands/kf-configure-space-get-buildpack-env/)	 - Get the environment variables for buildpack builds in a space.
* [kf configure-space get-container-registry](/docs/general-info/kf-cli/commands/kf-configure-space-get-container-registry/)	 - Get the container registry used for builds.
* [kf configure-space get-domains](/docs/general-info/kf-cli/commands/kf-configure-space-get-domains/)	 - Get domains associated with the space.
* [kf configure-space get-execution-env](/docs/general-info/kf-cli/commands/kf-configure-space-get-execution-env/)	 - Get the space-wide environment variables.
* [kf configure-space quota](/docs/general-info/kf-cli/commands/kf-configure-space-quota/)	 - Show quota info for a space
* [kf configure-space remove-domain](/docs/general-info/kf-cli/commands/kf-configure-space-remove-domain/)	 - Remove a domain from a space
* [kf configure-space set-build-service-account](/docs/general-info/kf-cli/commands/kf-configure-space-set-build-service-account/)	 - Set the service account to use when building containers
* [kf configure-space set-buildpack-builder](/docs/general-info/kf-cli/commands/kf-configure-space-set-buildpack-builder/)	 - Set the buildpack builder image.
* [kf configure-space set-buildpack-env](/docs/general-info/kf-cli/commands/kf-configure-space-set-buildpack-env/)	 - Set an environment variable for buildpack builds in a space.
* [kf configure-space set-container-registry](/docs/general-info/kf-cli/commands/kf-configure-space-set-container-registry/)	 - Set the container registry used for builds.
* [kf configure-space set-default-domain](/docs/general-info/kf-cli/commands/kf-configure-space-set-default-domain/)	 - Set a default domain for a space
* [kf configure-space set-env](/docs/general-info/kf-cli/commands/kf-configure-space-set-env/)	 - Set a space-wide environment variable.
* [kf configure-space unset-buildpack-env](/docs/general-info/kf-cli/commands/kf-configure-space-unset-buildpack-env/)	 - Unset an environment variable for buildpack builds in a space.
* [kf configure-space unset-env](/docs/general-info/kf-cli/commands/kf-configure-space-unset-env/)	 - Unset a space-wide environment variable.
* [kf configure-space update-quota](/docs/general-info/kf-cli/commands/kf-configure-space-update-quota/)	 - Update the quota for a space

