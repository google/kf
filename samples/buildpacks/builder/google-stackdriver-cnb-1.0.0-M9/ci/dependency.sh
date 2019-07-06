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

bin/dependency google-stackdriver-debugger-java "[\d]+\.[\d]+\.[\d]+" $(cat ../google-stackdriver-debugger-java/version) $(cat ../google-stackdriver-debugger-java/uri)  $(cat ../google-stackdriver-debugger-java/sha256)
commit google-stackdriver-debugger-java $(cat ../google-stackdriver-debugger-java/version)

bin/dependency google-stackdriver-profiler-java "[\d]+\.[\d]+\.[\d]+" $(cat ../google-stackdriver-profiler-java/version) $(cat ../google-stackdriver-profiler-java/uri)  $(cat ../google-stackdriver-profiler-java/sha256)
commit google-stackdriver-profiler-java $(cat ../google-stackdriver-profiler-java/version)
