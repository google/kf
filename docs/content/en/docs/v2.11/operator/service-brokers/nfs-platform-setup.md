---
title: Set up NFS platform
---

Kf supports Kubernetes native NFS, and exposes them with a dedicated `nfsvolumebroker` service broker for developers to consume. This broker has an `nfs` service offering which has a service plan named `existing`.

Use `kf marketplace` to see the service offering:

```none
$ kf marketplace
...
Broker           Name  Namespace  Description
nfsvolumebroker  nfs              mout nfs shares
...
```

Use `kf marketplace -s nfs` to see the service plan:

```none
$ kf marketplace -s nfs
...
Broker           Name      Free  Description
nfsvolumebroker  existing  true  mount existing nfs server
...
```

## Requirements

You need an NFS volume that can be accessed by your Kubernetes cluster. For example [Cloud Filestore](https://cloud.google.com/filestore) which Google's managed NFS solution that provides access to clusters in the same gcloud project.

## Prepare NFS

If you have an existing NFS service, you can use that. If you want a Google managed NFS service, [create a Filestore instance](https://cloud.google.com/filestore/docs/creating-instances) and Kf will automaticlaly configure the cluster to use it.

Warning: You only need to create the NFS instance. Kf will create relevant Kubernetes objects, including PersistentVolume and PersistentVolumeClaims. Do not manually mount the volume.

## What's next?

* [Configure NFS volumes]({{< relref "nfs-getting-started" >}})
