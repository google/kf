---
title: Configure NFS volumes
description: "Learn how to use an NFS volume as a mounted drive."
weight: 150
---

Kf supports mounting NFS volumes using the `kf marketplace`.

## Prerequisites

* Your administrator must have [completed the NFS platform setup guide]({{< relref "nfs-platform-setup.md" >}}).

## Create an NFS service instance

Run `kf marketplace` to see available services. The built-in NFS service appears on the list if NFS is enabled on the platform.

```
Broker           Name  Namespace  Description
nfsvolumebroker  nfs              mount nfs shares
```

## Mount an external filesystem

### Create a service instance

To mount to an existing NFS service:

```sh
kf create-service nfs existing SERVICE-INSTANCE-NAME -c '{"share":"SERVER/SHARE", "capacity":"CAPACITY"}'
```

Replace variables with your values.

* <var>SERVICE-INSTANCE-NAME</var> is the name you want for this NFS volume service instance.
* <var>SERVER/SHARE</var> is the NFS address of your server and share.
* <var>CAPACITY</var> uses the [Kubernetes quantity](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) format.

Confirm that the NFS volume service appears in your list of services. You can expect output similar to this example:

```none
$ kf services
...
Listing services in Space: demo-space
Name                Type      ClassName         PlanName  Age    Ready  Reason
filestore-nfs       volume    nfs               existing  6s     True   <nil>
...
```


### Bind your service instance to an App

To bind an NFS service instance to an App, run:

```sh
kf bind-service YOUR-APP-NAME SERVICE-NAME -c '{"uid":"2000","gid":"2000","mount":"MOUNT-PATH","readonly":true}'
```

Replace variables with your values.

* <var>YOUR-APP-NAME</var> is the name of the App for which you want to use the volume service.

* <var>SERVICE-NAME</var> is the name of the volume service instance you created in the previous step.

* `uid`:<var>UID</var> and `gid`:<var>GID</var> specify the directory permissions of the mounting share.

   {{< note >}}For V2 buildpack Apps, the value for `uid` and `gid` is always 2000.
   Otherwise, the user specified by `uid` and `gid` should be the `uid`
   and `gid` of the running App process.{{< /note >}}

* <var>MOUNT-PATH</var> is the path the volume should be mounted to within your App.

* (Optional) `"readonly":true` is an optional JSON string that creates a read-only mount.
  By default, Volume Services mounts a read-write file system.

  Note: Your App automatically restarts when the NFS binding changes.

You can list all bindings in a Space using the `kf bindings` command. You will see output similar to this example:

```none
$ kf bindings
...
Listing bindings in Space: demo-space
Name                                     App           Service             Age  Ready
binding-spring-music-filestore-nfs       spring-music  filestore-nfs       71s  True
...
```

### Access the volume service from your App

To access the volume service from your App, you must know which file path to use in your code.
You can view the file path in the details of the service binding, which are visible in the environment variables for your App.

View environment variables for your App:

```sh
kf vcap-services YOUR-APP-NAME
```

Replace <var>YOUR-APP-NAME</var> with the name of your App.

The following is example output of the `kf vcap-services` command:

```sh
$ kf vcap-services YOUR-APP-NAME
{
  "nfs": [
    {
      "instance_name": "nfs-instance",
      "name": "nfs-instance",
      "label": "nfs",
      "tags": [],
      "plan": "existing",
      "credentials": {
        "capacity": "1Gi",
        "gid": 2000,
        "mount": "/test/mount",
        "share": "10.91.208.210/test",
        "uid": 2000
      },
      "volume_mounts": [
        {
          "container_dir": "/test/mount",
          "device_type": "shared",
          "mode": "rw"
        }
      ]
    }
  ]
}
```

Use the properties under `volume_mounts` for any information required by your App.

| Property | Description |
| --- | --- |
| `container_dir` | String containing the path to the mounted volume that you bound to your App. |
| `device_type` | The NFS volume release. This currently only supports shared devices. A shared device represents a distributed file system that can mount on all App instances simultaneously. |
| `mode` | String that informs what type of access your App has to NFS, either `ro` (read-only), or `rw` (read and write). |
