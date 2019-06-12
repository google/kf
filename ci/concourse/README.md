# Concourse CI

# pipelines

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
  a. unit and integration tests are applied ([`hack/test.sh`](../../hack/test.sh))
  b. binaries are created ([`hack/build.sh`](../../hack/build.sh))
  c. lint checks are performed ([`hack/check-linters.sh`](../../hack/check-linters.sh) and [`hack/check-go-generate.sh`](../../hack/check-go-generate.sh))
3. If all parts of step 2 are successful, the Github PR check is set to `success`.
If any step fails, the PR check is set to `failure`.

# tasks

Shared tasks are placed in the `tasks` directory.

## set-status.yml

Sets the status on a Github PR check.


