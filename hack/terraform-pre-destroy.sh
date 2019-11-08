#!/usr/bin/env bash

set -eux
network=${1}

delete(){
  type=${1}
  gcloud compute $type list --filter "network=$network" \
    | cut -d' ' -f1 \
    | grep -v "^NAME$" \
    | xargs --no-run-if-empty gcloud compute $type delete --quiet
}

delete firewall-rules
gcloud compute networks delete $network --quiet
