---
title: Set up a custom domain
description: Learn to set up a DNS domain apps can use on your cluster.
---

All Kf Apps that serve HTTP traffic to users or applications
outside of the cluster must be associated with a domain name.

Kf has three locations where domains can be configured.
Ordered by precedence, they are:

1. Apps
2. Spaces
3. The `config-defaults` ConfigMap in the `kf` Namespace

## Edit the `config-defaults` ConfigMap

The `config-defaults` ConfigMap holds cluster-wide settings for Kf and can be edited by cluster administrators.
The values in the ConfigMap are read by the Spaces controller and modify their configuration.
Domain values are reflected in the Space's `status.networkConfig.domains` field.

To modify Kf cluster's domain, edit the `config-defaults` ConfigMap in the `kf` Namespace:

```sh
kubectl edit configmap config-defaults -n kf
```

Add or update the entry for the `spaceClusterDomain` key like the following:

```yaml 
spaceClusterDomain: my-domain.com
```

To validate the configuration was updated correctly, check the domain value in a Space:

```sh
kf space SPACE_NAME -o "jsonpath={.status.networkConfig.domains[]['domain']}"
```

The output will look similar to:

```none 
Getting Space some-space
some-space.my-domain.com
```

Each Space prefixes the cluster domains with its own name.
This prevents conflicts between Apps.

{{< warning >}}
Updating the `spaceClusterDomain` in `config-defaults` will immediately be
reflected on all Spaces and Apps that haven't overridden the domain.
{{< /warning >}}

## Assign Space domains

Spaces are the authoritative location for domain configuration.
You can assign domains and sub-domains to each Space for developers to use.
The field for configuring domains is `spec.networkConfig.domains`.

Use `kf space` to view the domains assigned to a Space:

```sh
kf space SPACE_NAME
```

In the output, the `Spec` field contains specific configuration for the Space
and the `Status` field reflects configuration for the Space with cluster-wide
defaults appended to the end:

```none 
...
Spec:
  Network Config:
    Domains:
      Domain: my-space.mycompany.com
...
Status:
  Network Config:
    Domains:
      Domain: my-space.mycompany.com
      Domain: my-space.prod.us-east1.kf.mycompany.com
```

{{< note >}} Apps will use the **first** domain the status if developers don't specify a domain.{{< /note >}}


### Add or remove domains using the CLI

The `kf` CLI supports mutations on Space domains. Each command outputs
a diff between the old and new configurations.

Add a new domain with `kf configure-space append-domain`:

```sh
kf configure-space append-domain SPACE_NAME myspace.mycompany.com
```

Add or make an existing domain the default with `kf configure-space set-default-domain`:

```sh
kf configure-space set-default-domain SPACE_NAME myspace.mycompany.com
```

And finally, remove a domain:

```sh
kf configure-space remove-domain SPACE_NAME myspace.mycompany.com
```

{{< note >}} You can override a cluster domains by inserting a new domain on the `spec` with the same `domain`.
This ensures the space won't be updated if the cluster domain changes.{{< /note >}}


## Use Apps to specify domains

Apps can specify domains as part of their configuration.
Routes are mapped to Apps during `kf push` using the following logic:

```none 
let current_routes  = The set of routes already on the app
let manifest_routes = The set of routes defined by the manifest
let flag_routes     = The set of routes supplied by the --route flag(s)
let no_route        = Whether the manifest has no-route:true or --no-route is set
let random_route    = Whether the manifest has random-route:true or --random-route is set

let new_routes = Union(current_routes, manifest_routes, flag_routes)

if new_routes.IsEmpty() then
  if random_route then
    new_routes.Add(CreateRandomRoute())
  else
    new_routes.Add(CreateDefaultRoute())
  end
end

if no_route then
  new_routes.RemoveAll()
end

return new_routes
```

If an App doesn't specify a Route, or requests a random Route, the first domain
on the Space is used. If the first domain on a Space changes, all Apps in the
Space using the default domain are updated to reflect it.

## Customize domain templates

Kf supports variable substitution in domains. Substitution allows a single
cluster-wide domain to be customized per-Space and to react to changes to the
ingress IP. Substitution is performed on variables with the syntax `$(VARIABLE_NAME)`
that occur in a domain.

| Variable             | Description |
| ---                  | ---         |
| `CLUSTER_INGRESS_IP` | The IPV4 address of the cluster ingress. |
| `SPACE_NAME`         | The name of the Space. |

{{< note >}} Use the `SPACE_NAME` variable in domains to allow developers in different
Spaces to avoid DNS conflicts if they push Apps with the same name.{{< /note >}}

### Examples

The following examples demonstrate how domain variables can be used to support
a variety of different organizational structures and cluster patterns.

* Using a wildcard DNS service like [nip.io](https://nip.io/):

  ```none
  $(SPACE_NAME).$(CLUSTER_INGRESS_IP).nip.io
  ```

* Domain for an organization with centrally managed DNS:

  ```none 
  $(SPACE_NAME).cluster-name.example.com
  ```

* Domain for teams who manage their own DNS:

  ```none 
  cluster-name.$(SPACE_NAME).example.com
  ```

* Domain for a cluster with warm failover and external circuit breaker:

  ```none 
  $(SPACE_NAME)-failover.cluster-name.example.com
  ```

## Differences between Kf and CF

* Kf Spaces prefix the cluster-wide domain with the Space name.
* Kf does not check for domain conflicts on user-specified routes.

