#!/usr/bin/env bash

set -euo pipefail

if [[ -d $PWD/go-module-cache && ! -d ${GOPATH}/pkg/mod ]]; then
  mkdir -p ${GOPATH}/pkg
  ln -s $PWD/go-module-cache ${GOPATH}/pkg/mod
fi

commit() {
  git commit -a -m "Dependency Upgrade: $1 $2" || true
}

cd "$(dirname "${BASH_SOURCE[0]}")/.."
git config --local user.name "$GIT_USER_NAME"
git config --local user.email ${GIT_USER_EMAIL}

go build -ldflags='-s -w' -o bin/dependency github.com/cloudfoundry/libcfbuildpack/dependency

bin/dependency gradle "[\d]+\.[\d]+\.[\d]+" $(cat ../gradle/version) $(cat ../gradle/uri)  $(cat ../gradle/sha256)
commit gradle $(cat ../gradle/version)

bin/dependency maven "[\d]+\.[\d]+\.[\d]+" $(cat ../maven/version) $(cat ../maven/uri)  $(cat ../maven/sha256)
commit maven $(cat ../maven/version)
