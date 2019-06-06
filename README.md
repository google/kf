# kf

`kf` provides a `cf` like experience on top of Knative.

`kf` aims to be fully compatible with CF applications and lifecycle; it supports
logs, buildpacks, app manifests, routing, service brokers, and injected services.

At the same time, it aims to improve the operational experience by supporting
git-ops, self-healing infrastructure, containers, a service mesh, autoscaling,
scale-to-zero, improved quota management and does it all on Kubernetes using
industry-standard OSS tools (Knative, Istio, and Tekton).

## Getting started

Build a `kf` binary and follow our [install instructions](docs/install.md)
for Knative.

## How to build

**Dependencies:**

[go mod](https://github.com/golang/go/wiki/Modules#quick-start)
is used and required for dependencies

**Requirements:**

  - Golang `1.12`

**Building:**

```sh
$ ./hack/build.sh
```

**Notes:**

- `kf` CLI must be built outside of the `$GOPATH` folder unless
you explicitly use `export GO111MODULE=on`.

## Development and releasing

We use [ko](https://github.com/google/ko) for rapid development
and during the release process to build a full set of `kf` images
and installation YAML.

To update your cluster while developing run `ko apply`:

```
KO_DOCKER_REPO=gcr.io/my-repo ko apply -f config
```

This will build any images required by `config/`, upload them to the provided
registry, and apply the resulting configuration to the current cluster.
