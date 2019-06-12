# Concourse CI

# Pipelines

The Concourse pipelines use YAML anchors to place all injected variables at the
beginning of the document. This is done for clarity.

[1]: ./pipelines/kf-pipeline.yml
## [kf-pipeline.yml][1]

Pipeline triggered by Github pull requests.

The [kf-pipeline.yml][1] will `fly set-pipeline` a pipeline for each pull
request. For example: Github PR `#5` will correspond to pipeline `5`.


NOTE: There is currently no cleanup operation. Mergers should remember to
`fly destroy-pipeline` once a PR has been merged/closed.

[2]: ./pipelines/pr-pipeline.yml
## [pr-pipeline.yml][2]

The pipeline created for each PR. It follows this order of operations:

1. The Github PR check is sent to `pending`
2. Various checks are performed in parallel:
  1. unit and integration tests are applied ([`hack/test.sh`](../../hack/test.sh))
  2. binaries are created ([`hack/build.sh`](../../hack/build.sh))
  3. lint checks are performed ([`hack/check-linters.sh`](../../hack/check-linters.sh) and [`hack/check-go-generate.sh`](../../hack/check-go-generate.sh))
3. If all parts of step 2 are successful, the Github PR check is set to `success`.
If any step fails, the PR check is set to `failure`.

## Pipeline Variables
[3]: https://concourse-ci.org/resources.html#resource-webhook-token

The following variables should be stored in your credential manager.

| Name                 | Description                                              | kf-pipeline | pr-pipeline |
| -------------------- | -------------------------------------------------------- | ----------- | ----------- |
| github_repo          | Github repo name                                         | ✅          |             |
| github_access_token  | Github access token for setting PR status checks         | ✅          |             |
| github_webhook_token | Github webhook argument. See [this][3]                   | ✅          |             |
| pr_comment           | Comment to leave on a PR when a pipeline is created      | ✅          |             |
| ci_git_uri           | Git URI for pulling this directory                       | ✅          |             |
| ci_git_branch        | Git branch for pulling this directory. Preferably master | ✅          |             |
| ci_image_uri         | Container image for pulling/pushing the build image      | ✅          |             |
| fly_user             | Concourse user                                           | ✅          |             |
| fly_password         | Concourse password                                       | ✅          |             |
| fly_target           | Concourse target                                         | ✅          |             |
| fly_url              | Concourse server URL                                     | ✅          |             |
| fly_team             | Concourse team                                           | ✅          |             |
| service_account_json | Google Cloud Service Account Key for pushing build image | ✅          |             |
| gcp_project_id       | Google Cloud Project where test GKE cluster resides      |             | ✅          |
| k8s_cluster_name     | GKE cluster name                                         |             | ✅          |
| k8s_cluster_zone     | GKE cluster zone                                         |             | ✅          |

# Tasks

Shared tasks are placed in the `tasks` directory.

[4]: tasks/set-status.yml
## [set-status.yml][4]

Sets the status on a Github PR check.

# Images

A custom container image is used throughout the pipelines. It contains all the
binaries required by the different tasks. It can be built by the [`kf-pipeline`][1].
