# kf Deployment

This folder contains deployment files for `kf`.

## Notes for additional resources

`kubectl -R -f` installs the files within a folder in alphabetical order.
In order to install the files with correct ordering within a folder, a three digit prefix is added.

Files with a prefix require files with smaller prefixes to be installed before they are installed.
Files with the same prefix can be installed in any order within the set sharing the same prefix.
Files without any prefix can be installed in any order and they don't have any dependencies.

A rough guide for prefixing is the following:

* `1xx` - Kubernetes namespaces
* `2xx` - RBAC roles, service accounts, bindings and gateways
* `3xx` - Custom resource definitions (CRDs)
* `4xx` - Services
* `config-*` - ConfigMaps
