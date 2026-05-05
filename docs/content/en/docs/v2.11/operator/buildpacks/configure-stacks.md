---
title: "Customize stacks"
---

The Stack configuration can be updated by editing the `kfsystem` Custom Resource:

```sh
kubectl edit kfsystem kfsystem
```

This example sets the Google Cloud buildpacks a V3 Stack:

```
spec:
  kf:
    config:
      spaceStacksV3:
      - name: google
        description: Google buildpacks (https://github.com/GoogleCloudPlatform/buildpacks)
        buildImage: gcr.io/buildpacks/builder:v1
        runImage: gcr.io/buildpacks/gcp/run:v1
```

This new Stack can now be pushed:

```sh
kf push myapp --stack google
```

This example configures the Ruby V2 buildpack and sets the build pipeline defaults to use V2 Stacks:

```
spec:
  kf:
    config:
      spaceDefaultToV3Stack: false
      spaceBuildpacksV2:
      - name: ruby_buildpack
        url: https://github.com/cloudfoundry/ruby-buildpack
      spaceStacksV2:
      - name: cflinuxfs5
        image: cloudfoundry/cflinuxfs5@sha256:206c68c5753abd24beb85565e9ed525d1f1be24bdab252193793b49b44d4b69e

```
