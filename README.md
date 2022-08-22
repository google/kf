# Kf

This is not an officially supported Google product

## Getting started the manual way

Follow the install instructions at go/kf-docs to create a GKE cluster,
install Kf into it, and deploy an app with the `kf` CLI.

## Deploy a local Kf install to a new cluster

If you need to set up a new development cluster run the following command:

```sh
./hack/deploy-dev-release.sh
```

It will fetch all your local sources and kick off a Cloud Build that builds
a version of Kf, creates a GKE cluster and installs the Kf version onto it.

## Iterative development

**Building the CLI:**

```sh
$ ./hack/build.sh
```

**Installing Kf server-side components:**

We use [ko](https://github.com/google/ko) for rapid development
and during the release process to build a full set of `kf` images
and installation YAML. Run the following to stage local changes on
a targeted cluster:

```sh
$ ./hack/ko-apply.sh
```

This will build any images required by `config/`, upload them to the provided
registry, and apply the resulting configuration to the current cluster.
