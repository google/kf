#!/usr/bin/env bash

set -euo pipefail

if [[ -d $PWD/go-module-cache && ! -d ${GOPATH}/pkg/mod ]]; then
  mkdir -p ${GOPATH}/pkg
  ln -s $PWD/go-module-cache ${GOPATH}/pkg/mod
fi

commit() {
  git commit -a -m "Dependency Upgrade: $1 $2" || true
}

version() {
  local PATTERN="([0-9]+)\.([0-9]+)\.([0-9]+)-(.*)"

  for VERSION in $(cat $1); do
      if [[ ${VERSION} =~ ${PATTERN} ]]; then
        echo "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${BASH_REMATCH[3]}"
        return
      else
        >2 echo "version is not semver"
        exit 1
      fi
    done
}

cd "$(dirname "${BASH_SOURCE[0]}")/.."
git config --local user.name "$GIT_USER_NAME"
git config --local user.email ${GIT_USER_EMAIL}

go build -ldflags='-s -w' -o bin/dependency github.com/cloudfoundry/libcfbuildpack/dependency

bin/dependency tomcat "7\.[\d]+\.[\d]+" $(cat ../tomcat-7/version) $(cat ../tomcat-7/uri)  $(cat ../tomcat-7/sha256)
commit tomcat $(cat ../tomcat-7/version)

bin/dependency tomcat "8\.[\d]+\.[\d]+" $(cat ../tomcat-8/version) $(cat ../tomcat-8/uri)  $(cat ../tomcat-8/sha256)
commit tomcat $(cat ../tomcat-8/version)

bin/dependency tomcat "9\.[\d]+\.[\d]+" $(cat ../tomcat-9/version) $(cat ../tomcat-9/uri)  $(cat ../tomcat-9/sha256)
commit tomcat $(cat ../tomcat-9/version)

bin/dependency tomcat-access-logging-support "[\d]+\.[\d]+\.[\d]+" $(version ../tomcat-access-logging-support/version) $(cat ../tomcat-access-logging-support/uri)  $(cat ../tomcat-access-logging-support/sha256)
commit tomcat-access-logging-support $(version ../tomcat-access-logging-support/version)

bin/dependency tomcat-lifecycle-support "[\d]+\.[\d]+\.[\d]+" $(version ../tomcat-lifecycle-support/version) $(cat ../tomcat-lifecycle-support/uri)  $(cat ../tomcat-lifecycle-support/sha256)
commit tomcat-lifecycle-support $(version ../tomcat-lifecycle-support/version)

bin/dependency tomcat-logging-support "[\d]+\.[\d]+\.[\d]+" $(version ../tomcat-logging-support/version) $(cat ../tomcat-logging-support/uri)  $(cat ../tomcat-logging-support/sha256)
commit tomcat-logging-support $(version ../tomcat-logging-support/version)
