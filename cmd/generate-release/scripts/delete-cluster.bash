#!/usr/bin/env bash

# This has been added to work around spurious problems with deleting GKE cluster
# at the end of integration tests (usualy due to "operation in progress")

set -eux

if [ -z ${1:+x} ]; then
  echo "usage: $0 [CLUSTER_DEPLOYMENT]"
  exit 1
fi

deployment_name=$1

# retry will run the command, if it fails, it will sleep and then try again:
# args:
# 1: command: The command to run. (required)
# 2: retry_count: The number of times to attempt the command. (defaults to 3)
# 3: sleep_seconds: The number of seconds to sleep after a failure. (defaults to 1)
function retry() {
    command=${1:-x}
    retry_count=${2:-3}
    sleep_seconds=${3:-1}

    if [ "${command}" = "x" ]; then
        echo "command missing"
        exit 1
    fi

    n=0
    until [ "$n" -ge ${retry_count} ]
    do
       ${command} && break
       echo "failed..."
       n=$((n+1))
       sleep ${sleep_seconds}
    done

    if [ ${n} = ${retry_count} ];then
        echo "\"${command}\" failed too many times: (${retry_count})"
        exit 1
    fi
}

retry "gcloud deployment-manager --quiet \
       deployments delete ${deployment_name}" 15 60
