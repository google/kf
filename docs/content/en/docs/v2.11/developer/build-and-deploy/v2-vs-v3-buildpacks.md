---
title: Compare V2 and V3 Buildpacks
description: >
    Learn how to decide wihch buildpacks to use.
weight: 500
---

A [buildpack](https://buildpacks.io/) converts source code into an executable, and is used to deliver a simple, reliable, and repeatable way to create containers. Kf supports both V2 and V3 buildpacks, and it is important to understand the differences between them.

## V2 buildpacks

Most Cloud Foundry applications already use V2 buildpacks. When using V2 buildpacks with Kf, the lifecycle binaries and the buildpacks are downloaded and configured from their git URLs. Kf then uses the `lifecycle` CLI to execute each buildpack against the source code.

### Pros
* Ready out of the box without pipeline or code changes.

### Cons
* Legacy buildpack supersceded by V3.
* Weaker performance and reliability. The Kf build pipeline requires more IO for V2 buildpacks.
* Fewer community resources.
* Kf only supports OSS git repos.


## V3 buildpacks

V3 buildpacks are a Cloud Native Computing Foundation (CNCF) project with a well defined [spec](https://github.com/buildpacks/spec/blob/main/buildpack.md), a CLI ([pack](https://github.com/buildpacks/pack)) and a growing community that is innovating around different languages and frameworks. Google Cloud also has its own set of [OSS buildpacks](https://github.com/GoogleCloudPlatform/buildpacks).

V3 buildpacks have two overarching OCI containers:

* Builder image
* Run image

### Builder image

The builder image is used while your source code is being built into a runnable container. The image has the necessary `detect` scripts and other utilities to compile source code.

{{< note >}} Most buildpacks don't package the compilers with the builder image. Instead, they may have the build pipelines download any dependencies required to compile.{{< /note >}}

### Run image

The run image is the base image that a container is built on. This means that it is the base image that will run when the App executes.

### Layers

V3 buildpacks use layers to compose the final container. Each buildpack included in a build is given the opportunity to manipulate the file system and environment variables of the App. This layering approach allows for buildpacks to be thinner and more generic.

V3 buildpacks are built on OCI containers. This requires that the V3 builder image be stored in a container registry that the Kf build pipeline has access to. The build pipeline uses the builder image to apply the underlying scripts to build the source code into a runnable container.

### Pros

* Google supported [builder and run image](https://github.com/GoogleCloudPlatform/buildpacks).
* Works with various CI/CD runtimes like [Cloud Build](https://cloud.google.com/functions/docs/building/pack).
* Growing community and [buildpack registry](https://registry.buildpacks.io/).

### Cons

* May require code/process updates. For example, the Java buildpack requires source code while the V2 buildpack requires a jar file.
* V3 buildpacks are newer and might require additional validation (is using community developed buildpacks).

## Kf Stacks

### View Stacks

When pushing an App, the build pipeline determines the buildpack based on the selected Stack (specified via the `--stack` flag or the manifest).

To see which Stacks are available in a Space first ensure a Space is targeted:

```sh
kf target -s myspace
```

The `kf stacks` subcommand can then be used to list the Stacks:

```sh
kf stacks
```

The output shows both V2 and V3 Stacks:

```
Getting stacks in Space: myspace
Version  Name                                Build Image                                                                                          Run Image
V2       cflinuxfs3                          cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5      cloudfoundry/cflinuxfs3@sha256:5219e9e30000e43e5da17906581127b38fa6417f297f522e332a801e737928f5
V3       kf-v2-to-v3-shim                    gcr.io/kf-releases/v2-to-v3:v2.7.0                                                                   gcr.io/buildpacks/gcp/run:v1                                                                       This is a stack added by the integration tests to assert that v2->v3 shim works
V3       google                              gcr.io/buildpacks/builder:v1                                                                         gcr.io/buildpacks/gcp/run:v1                                                                       Google buildpacks (https://github.com/GoogleCloudPlatform/buildpacks)
V3       org.cloudfoundry.stacks.cflinuxfs3  cloudfoundry/cnb:cflinuxfs3@sha256:f96b6e3528185368dd6af1d9657527437cefdaa5fa135338462f68f9c9db3022  cloudfoundry/run:full-cnb@sha256:dbe17be507b1cc6ffae1e9edf02806fe0e28ffbbb89a6c7ef41f37b69156c3c2  A large Cloud Foundry stack based on Ubuntu 18.04
```


## V2 to V3 Buildpack Migration

{{< note >}} This feature is currently experimental and subject to change.{{< /note >}}

Kf provides a V3 stack to build applications that were built with standard V2 buildpacks, using a stack named  `kf-v2-to-v3-shim`. The `kf-v2-to-v3-shim` stack is created following the [standard V3 buildpacks API](https://buildpacks.io/docs/operator-guide/create-a-stack/). A Google maintained builder image is created with each Kf release, following the [standard buildpack process](https://buildpacks.io/docs/operator-guide/create-a-builder/). The builder image aggregates a list of V3 buildpacks created by the same process used with the `kf wrap-v2-buildpack` command. The V3 buildpack images are created using the standard V2 buildpack images. It's important to note that the V3 buildpacks do not contain the binary of the referenced V2 buildpacks. Instead, the V2 buildpack images are referenced, and the bits are downloaded at App build time (by running `kf push`). 

At App build time, the V2 buildpack is downloaded from the corresponding git repository. When V3 detection runs, it delegates to the downloaded V2 detection script. For the first buildpack group that passes detection, it proceeds to the build step, which delegates the build execution to the downloaded V2 builder script. 

The following V2 buildpacks are supported in the `kf-v2-to-v3-shim` stack:

| Buildpack | Git Repository |
| ------ | ----- |
| java_buildpack | https://github.com/cloudfoundry/java-buildpack |
| dotnet_core_buildpack | https://github.com/cloudfoundry/dotnet-core-buildpack |
| nodejs_buildpack | https://github.com/cloudfoundry/nodejs-buildpack |
| go_buildpack | https://github.com/cloudfoundry/go-buildpack |
| python_buildpack | https://github.com/cloudfoundry/python-buildpack |
| binary_buildpack | https://github.com/cloudfoundry/binary-buildpack |
| nginx_buildpack | https://github.com/cloudfoundry/nginx-buildpack |

#### Option 1: Migrate Apps built with standard V2 buildpacks

To build Apps with the `kf-v2-to-v3-shim` stack, use the following command:

```sh
kf push myapp --stack kf-v2-to-v3-shim
```

The `kf-v2-to-v3-shim` stack will automatically detect the runtime with the wrapped V2 buildpacks. The resulting App image is created using the V3 standard and build pipeline, but the builder of the equivalent V2 buildpack. 

#### Option 2: Migrate Apps built with custom V2 buildpacks

Kf has a buildpack migration tool that can take a V2 buildpack and wrap it with a V3 buildpack. The wrapped buildpack can then be used anywhere V3 buildpacks are available.

```sh
kf wrap-v2-buildpack gcr.io/your-project/v2-go-buildpack https://github.com/cloudfoundry/go-buildpack --publish
```

This will create a buildpack image named `gcr.io/your-project/v2-go-buildpack`. It can then be used to create a builder by following the [create a builder](https://buildpacks.io/docs/operator-guide/create-a-builder/) docs.

This subcommand uses the following CLIs transparently:

* `go`
* `git`
* `pack`
* `unzip`
