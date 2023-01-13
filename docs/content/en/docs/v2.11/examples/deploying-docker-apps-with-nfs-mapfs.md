---
title: Deploy Docker apps with NFS UID/GID mapping.
description: >
  Learn how to deploy Docker applications with NFS UID/GID mapping.
---

This document outlines how to do UID/GID mapping for NFS volumes within a Docker container.
This may be necessary if your application uses NFS because Kubernetes assumes the UID and GID of
the NFS volume map directly into the UID/GID namespace of your container.

{{<warning>}}This method should be used as a last resort. Instead you should use an NFS client library
in your application code. This will minimize the chance for errors and will allow your app to tie
health checks directly to the health of backing storage.
{{</warning>}}

{{<note>}}This isn't supported for Kf versions before v2.11.9.{{</note>}}


To get around this limitation, Kf adds the `mapfs` binary to all continers it builds. The `mapfs`
binary creates a FUSE filesystem that maps the UID and GID of a host container into the UID and GID
of an NFS volume.

## Prerequisites

In order for these operations to work:

* Your container's OS must be Linux.
* Your container must have the coreutils `timeout`, `sh`, and `wait` installed.
* Your container must have `fusermount` installed.

## Update your Dockerfile

First, you'll need to update your Dockerfile to add the `mapfs` binary to your application:

```dockerfile
# Get the mapfs binrary from a version of Kf.
FROM gcr.io/kf-releases/fusesidecar:v2.11.14 as builder
COPY --from=builder --chown=root:vcap /bin/mapfs /bin/mapfs

# Allow users other than root to use fuse.
RUN echo "user_allow_other" >> /etc/fuse.conf
RUN chmod 644 /etc/fuse.conf

RUN chmod 750 /bin/mapfs
# Allow setuid so the mapfs binary is run as root.
RUN chmod u+s /bin/mapfs
```

## Set manifest attributes

Next, you'll have to update manifest attributes for your application.
You MUST set `args` and `entrypoint` because they'll be used by `mapfs` to launch the application.

* Set `args` to be your container's `CMD`
* Set `entrypoint` to be your container's `ENTRYPOINT`

```yaml
applications:
- name: my-docker-app
  args: ["-jar", "my-app"]
  entrypoint: "java"
  dockerfile:
    path: gcr.io/my-application-with-mapfs
```

{{<note>}}If no `entrypoint` is specified in the manifest, Kf will use `/lifecycle/entrypoint.bash`.{{</note>}}

## Deploy your application

Once your Docker image and manifest are updated, you can deploy your application and check that your NFS volume mounting correctly in the
container.

If something has gone wrong, you can debug it by getting the Deployment in Kubernetes with the same name as your application:

```
kubectl get deployment my-docker-app -n my-space -o yaml
```

Validate the `command` and `args` for the container named `user-container` look valid.