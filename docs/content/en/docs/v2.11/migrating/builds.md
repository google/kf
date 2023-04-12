---
title: "Build Container Images"
linkTitle: "Build Container Images"
weight: 100
description: Learn how to build container images that can be used with Kf and Kubernetes.
---

The first step you'll want to take when migrating from Kf to another container runtime is to change the way your software is built so you build containers rather than relying on Kf to do it on your behalf.

Best practices for container runtimes differ from Cloud Foundry's application best practices so you may need to change some organizational processes as well as technical processes.

## Best practices

*   **Images should be small**. Images need to be pulled over a network potentially every time the container starts. Lager images cost more to store, transfer, and require bigger disks on the machines you run the image from.
*   **Images should have few dependencies.** More dependencies creates larger images and a larger attack surface. A base image like [distroless](https://github.com/GoogleContainerTools/distroless) or [alpine](https://www.alpinelinux.org/) can reduce your size while maintaining critical components like security certificates.
*   **Rebuild rather than patch images**. Images are built in layers so patching either increases the size of the image by applying the patches to a new layer or replaces lower layers which were used to produce upper layers which is unsafe.
*   **Images should be built reproducibly.** [Reproducible builds](https://reproducible-builds.org/) not only give confidence in supply chain security, but also reduce storage cost by allowing better deduplication of layers shared between images.
*   **Containers should be non-root.** A root process in a container will have the same privilege as root on the host if it escapes the container. Non-root containers mean both a privilege escalation exploit and a container escape exploit are necessary to compromise the host.
*   **Reference images by SHA, not tag.** Referencing images by tag, or the implicit tag `latest`, means that your builds or runtimes may not be reproducible.


## Differences from Cloud Foundry and Kf

In Cloud Foundry and Kf it's common to `push` in every environment which produces a separate build and image. This is a common source of errors if an application development team hasn't pinned the dependencies for their application correctly.

In Cloud Foundry and Kf this pushing provides an additional purpose, a platform operations team can change buildpacks or base images in an environment to patch applications on the fly. This has always been risky, but large base images like `cflinuxfs3`, come with many security vulnerabilities.

**Things to consider:**

*   Applications should be built with the dependencies they need for all environments. They shouldn't depend on the environment like the [Spring Autoreconfiguration buildpack](https://github.com/cloudfoundry/java-buildpack-auto-reconfiguration).
*   Teams should be able to patch and promote an image quickly up through environments as the preferred method of patching.
*   Building on smaller base images than `cflinuxfs3` may leave teams without tools they expect to use when using `ssh` to debug containers or may expose hidden dependencies.
*   Cloud Foundry and Kf add a [launcher process](https://github.com/cloudfoundry/buildpackapplifecycle/tree/main/launcher) responsible for reading Procfiles and certain environment variables. Applications will need to stop using these, or you should create a drop-in replacement launcher.
*   Operations teams should monitor deployed images for vulnerabilities.


## Approaches

There are two major approaches to move from V2 buildpacks to building containers.


### Cloud Native Buildpacks

[Cloud Native Buildpacks (CNBs)](https://buildpacks.io/) are a joint effort by Pivotal and Heroku to bring a similar buildpack based development experience to containers. The project belongs to the CNCF and has a wide variety of buildpacks available.

The Cloud Foundry Foundation created [Paketo](https://paketo.io/) a set of buildpacks that work together which are the next generation of the v2 buildpacks that Cloud Foundry uses. They're also based on Ubuntu meaning migration should be less risky for CF applications.

It's possible for an operations team to customize buildpack pipelines with [company specific base images or buildpacks](https://buildpacks.io/docs/operator-guide/).

Images built with buildpacks are designed to be rebased on newer operating system layers, similar to what is done in Cloud Foundry. If building new images and promoting them is too much of an organizational hurdle, this may be a good option for patching large-scale vulnerabilities across a fleet of deployed applications.

Cloud Native Buildpacks can also achieve similar performance to Cloud Foundry builds using [caches](https://buildpacks.io/docs/buildpack-author-guide/create-buildpack/caching/), which are separate container images created by the pipeline.

Developers can run the Cloud Native Buildacks locally to build and test images as they'd appear in production to reduce the number of failed builds.


### Dockerfile

[Dockerfiles](https://docs.docker.com/engine/reference/builder/) are the second major option for producing container images. Dockerfiles specify a set of commands that are executed in a container. Each command creates a new layer that captures the delta from the layer before. The end result is an image.

Dockerfiles are very flexible and can be used to reliably reproduce nearly any build process. In fact, the same process Kf uses to execute v2 buildpacks can be turned into a Dockerfile.

Historically, Dockerfiles needed a root daemon on the host machine and additional permissions for each user to connect to it, but there are now alternatives that can run in userspace rootless like [Podman](https://podman.io/).

Building Dockerfiles that rely on untrusted images with a Docker daemon running as root is risky. You should either restrict the network on build machines to only allow pulling trusted images and/or use a rootless container build engine that can execute Dockerfiles like [Kaniko](https://github.com/GoogleContainerTools/kaniko) or [Buildah](https://github.com/containers/buildah/blob/main/README.md). Kf uses Kaniko, but in certain circumstances it can be slow if it has to handle many files.

To reduce the need for developers to author Dockerfiles, a platform operations team may create steps in their build pipeline that containerize the existing output that would normally be fed to Cloud Foundry or Kf.

Rebasing containers created by Dockerfiles is possible with [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane.md) if you know which image was used to produce the container, but it's risky unless you can guarantee all items in the new image are ABI compatible with the previous layer.


## Migration path

Kf supports deploying applications from container images. You can migrate to building images rather than intermediate formats like JARs and change your deployment pipelines to use those images instead.

This will also reduce the resources consumed on your Kf clusters by build processes and the need to store source containers.

## Additional resources

*   Learn how to pull large images faster on GKE using [image streaming](https://cloud.google.com/kubernetes-engine/docs/how-to/image-streaming).
*   Learn about how reproducible builds [help your security posture](https://reproducible-builds.org/). 
*   Learn how to [attach debugging tools to existing Kubernetes pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#ephemeral-container) to reduce the footprint of your images. 