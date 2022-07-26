#!/usr/bin/env bash

set -eux

# Helper function to gracefully fail a script by printing
# the error message and returning with exit code.
function fail_script() {
  [[ -n "$1" ]] && echo "ERROR: $1. Failing script."
  exit 1
}

# Helper method to apply yaml onto GKE Cluster.
# we do wait based retry and give up after 3 retries.
# Parameters:
#   1: filename to apply
function apply_yaml() {
  local file=$1
  local retry
  local max_retries=3
  for (( retry=0; retry<max_retries; retry++)) do
    if kubectl apply -f "${file}"; then
      echo "Successfully applied ${file}"
      return
    else
      sleep 5s
    fi
  done

  fail_script "Failed applying ${file}"
}

kubectl apply -f "/kf/bin/operator.yaml"
apply_yaml "/kf/bin/kfsystem.yaml"
