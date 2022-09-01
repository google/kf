# How to add specific buildPacks to your kf app manifest file

## Description

Buildpacks can support pinning by get tags instead of automatically sourcing the latest buildpack from a git repository.

## Instructions

Note: If a space was created before doing this process, then the space will not automatically update itself

Create a new file that includes the existing buildPacks, as well as the new buildPacks and tags that are to be included

`touch patch-file.yaml`

Add new buildPacks as follows and use a git tag to specify which version of the buildPacks the app should use, if not the latest

For example from the Go buildpack which can be found here https://github.com/cloudfoundry/go-buildpack a user can specify a tag using the following format:

url: https://github.com/cloudfoundry/go-buildpack.git#v1.9.48

Populate the patch-file.yaml using the following format

```
data:
  buildpacksV2: |
    - name: [buildpack name]
      url: [buildpack url, specifying tag if appropriate]
```
For example, to pin Go Lang buildpack to version 1.9.48 
```
data:
  buildpacksV2: |
    - name: go_buildpack_v1.9.48
      url: https://github.com/cloudfoundry/go-buildpack.git#v1.9.48
```

Use kubectl to apply the changes from that patch file directly to the existing `./config/config-defaults.yaml` file 

`kubectl patch configmap config-defaults --patch-file ./patch-file.yaml`

Note: 
1. The configmap name in this case is config-defaults
2. Applying the patch-file.yaml file will replace all the buildPacks under data/buildpacksV2 path which means you have to add all the buildPacks your spaces and apps will be needing.

To check the config map has been successfully updated run 

`kubectl describe configmap config-defaults -n kf`

Here you should see the amended list of buildPacks under data/buildpacksV2, newly created spaces will have the updated buildPacks.

