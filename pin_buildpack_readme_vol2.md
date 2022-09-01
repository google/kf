# How to add specific buildpacks to your kf app manifest file

## Description

Buildpacks can support pinning by using tags instead of automatically sourcing the latest buildpack from a git repository

<br />

## Instructions

Buildpacks can be added using the `kfsystem` operator and `kubectl patch`

Add new buildpacks as follows and use a git tag to specify which version of the buildpack the app should use. Otherwise the buildpack will default to the latest version

For example, to pin Golang buildpack to version 1.9.48 use

```
kubectl patch \
kfsystem kfsystem \
--type='json' \
-p='[{"op":"add","path":"./config/config-defaults.yaml/data/spaceBuildpacksV2","value":{"name":"go_buildpack_v1.9.48"}}]'
```

This command will add the following to the `./config/config-defaults.yaml` file

```
data:
  SpaceBuildpacksV2: |
    - name: go_buildpack_v1.9.48
      url: https://github.com/cloudfoundry/go-buildpack.git#v1.9.48
```