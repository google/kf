#!/bin/bash

# Quit on error
set -e

if [[ ! `pwd` == *docs/kf.dev ]] ;
then
  echo 'Run this script with `hack/localpreview.sh` from the "docs/kf.dev" dir.'
  exit 1
fi

# Run `hugo serve` from the Docker image that drives CI
docker run \
  -p 1313:1313 \
  --mount type=bind,source="$(pwd)",target=/docs \
  --net=host \
  gcr.io/kf-source/website-ci-image hugo serve --enableGitInfo=false -s /docs
