#!/bin/bash

# Quit on error
set -e

if [[ ! `pwd` == *docs/kf.dev ]] ;
then
  echo 'Run this script with `hack/build-branch.sh` from the "docs/kf.dev" dir.'
  exit 1
fi

if [[ -z "$GOOGLE_PROJECT" ]];
then
  echo GOOGLE_PROJECT is required.
  exit 1
fi

if [[ -z "$VERSION" ]];
then
  echo VERSION is required and should correspond to the branch or tag you are building.
  exit 1
fi

promote_flag="--promote"
if [[ -z "$DEFAULT" ]] || [[ "$DEFAULT" != "true" ]];
then
  echo This version will be deployed to App Engine but it will not be the default. Set DEFAULT=true to make this the default version.
  promote_flag="--no-promote"
fi

# Build site with Hugo
export NODE_PATH=$NODE_PATH:`npm root -g`
hugo

# Move generated contents and app.yaml to tempdir
tmp_dir=$(mktemp -d -t ci-XXXXXXXXXX)
cp app.yaml $tmp_dir/app.yaml
ln -s `pwd`/public $tmp_dir/www

pushd $tmp_dir
  ls -la
  gcloud app deploy \
    --project $GOOGLE_PROJECT \
    --version $VERSION \
    $promote_flag \
    --quiet
popd
