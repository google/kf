# Concourse CI

# Pipelines

The Concourse pipelines use YAML anchors to place all injected variables at the
beginning of the document. This is done for clarity.

[1]: ./pipelines/website-pipeline.yml
## [website-pipeline.yml][1]

Pipeline triggered by changes to the `docs/kf.dev` dir on the `master` branch.

## Pipeline Variables
[3]: https://concourse-ci.org/resources.html#resource-webhook-token

The following variables should be stored in your credential manager.

| Name                 | Description                                              | website-pipeline |
| -------------------- | -------------------------------------------------------- | ---------------- |
| ci_git_uri           | Git URI for pulling this directory                       | ✅               |
| ci_git_branch        | Git branch for pulling this directory. Preferably master | ✅               |
| ci_image_uri         | Container image for pulling/pushing the build image      | ✅               |
| service_account_json | Google Cloud Service Account Key for pushing build image | ✅               |
| gcp_project_id       | Google Cloud Project where test GKE cluster resides      |                  |

# Tasks

Shared tasks are placed in the `tasks` directory.

# Images

A custom container image is used throughout the pipelines. It contains all the
binaries required by the different tasks. It can be built by [`website-pipeline`][1].
