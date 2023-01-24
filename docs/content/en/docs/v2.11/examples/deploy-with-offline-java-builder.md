---
title: Deploy With Offline Java Buildpack
description: >
  Learn how to install and use an offline Java Buildpack (provided by Cloud Foundry) to compile your Java apps.
---

This document shows how to use an offline Java Buildpack to deploy your
applications.

Cloud Foundry's [Java
buildpack](https://github.com/cloudfoundry/java-buildpack/tree/main) uses a
number of large dependencies. At the time of writing, ~800 MB of sources are
pulled into the builder during the buildpack's execution. Much of this data is
brought in from the internet which is convenient as it keeps the buildpack
itself small but introduces a great deal of data transfer.

Java builds can be optimized to reduce outside network ingress and improve
performance by hosting your own Java buildpack compiled in an [offline
mode](https://github.com/cloudfoundry/java-buildpack/blob/main/docs/buildpack-modes.md#offline-mode).
In its offline mode, Cloud Foundry's Java buildpack downloads packages that may
be used into the cache when creating the builder. This avoids pulling
dependencies from the internet at runtime and makes the builder image self
contained.

## Before You Begin

You will need a cluster with Kf installed and access to the Kf CLI

Additionally, you will need access to the following software:

- **`git`**: Git is required to clone a repository.
- **`ruby`**: Ruby is required to create the Java bulidpack.
- **`bundle`**: Ruby package manager to install Java buildpack dependencies.

## Compile the Java Buildpack in Offline mode

Follow Cloud Foundry's instructions to compile the Java Buildpack in Offline
mode:
https://github.com/cloudfoundry/java-buildpack/blob/main/README.md#offline-package.

These instructions will generate a `.zip` extension file containing the
buildpack and its dependencies.

## Deploy the Java Buildpack

### Self Hosted on Kf (Recommended)

Once you have built the Java Buildpack, you can host it on your Kf cluster using
a staticfile buildpack.

1. Create a new directory for your static file Kf app i.e.
   `java-buildpack-staticapp`
2. Create a new manifest in that directory `manifest.yml` with the contents:

```yaml
---
applications:
  - name: java-buildpack-static
```

3. Create an emptyfile named `Staticfile` (case sensitive) in the directory.
   This indicates that this is a static app to the staticfile buildpack which
   will create a small image containing the contents of the directory + an
   [nginx](https://www.nginx.com/) installation to host your files.
4. Copy the buildpack you created in the previous step into the directory as
   `java-buildpack-offline-<hash>.zip`. Your directory structure at this point
   should resemble:

```
/
├── java-buildpack-offline-fe26136c.zip
├── manifest.yml
└── Staticfile
```

6. Run `kf push` to deploy your app, this will trigger a build.
7. When the push finishes, run `kf apps` and take note of the URL your app is
   running on. You will reference this in the next step.

See [internal routing]({{< relref "routes-and-domains#internal-routing" >}}) to
construct an internal link to your locally hosted buildpack. The URL should
resemble `<your route>/java-buildpack-offline-<hash>.zip`. Take note of this URL
as you will reference it later in your application manifests or in your
cluster's buildpack configuration.

### Served from Google Cloud Storage

See the Google Cloud documentation on creating public objects in a Cloud Storage
bucket: https://cloud.google.com/storage/docs/access-control/making-data-public.

To find the URL where your buildpack is hosted and that you will reference in
your manifests, see: https://cloud.google.com/storage/docs/access-public-data.

## Using an Offline Buildpack

### Whole Cluster

You can apply your buildpack to the whole cluster by updating the cluster's
configuration. See [configuring stacks]({{<relref "configure-stacks.md">}}) to
learn how to register your buildpack URL for an entire cluster.

### In App Manifest

To use your offline buildpack in a specific app, update your application
manifest to specify the buildpack explicitly i.e.

```
---
applications:
- name: my-app
  buildpacks:
    - http://<your host>/java-buildpack-offline-<hash>.zip
```
