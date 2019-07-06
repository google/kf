#!/usr/bin/env bash

set -euo pipefail

if [[ -d $PWD/go-module-cache && ! -d ${GOPATH}/pkg/mod ]]; then
  mkdir -p ${GOPATH}/pkg
  ln -s $PWD/go-module-cache ${GOPATH}/pkg/mod
fi

PACKAGE_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'package')
ARTIFACTORY_DIR="${PWD}/artifactory"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

go build -ldflags='-s -w' -o bin/package github.com/cloudfoundry/libcfbuildpack/packager
bin/package ${PACKAGE_DIR}

cd ${PACKAGE_DIR}

ID=$(sed -n 's|id = \"\(.*\)\"|\1|p' buildpack.toml | head -n1)
VERSION=$(sed -n 's|version = \"\(.*\)\"|\1|p' buildpack.toml | head -n1)
TARGET="${ARTIFACTORY_DIR}/$(echo ${ID} | sed 's|\.|/|g')/${ID}/${VERSION}/${ID}-$(echo ${VERSION} | sed "s|SNAPSHOT|$(date '+%Y%m%d.%H%M%S')-1|").tgz"

mkdir -p $(dirname ${TARGET})
tar czf ${TARGET} *
