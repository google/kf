#!/usr/bin/env bash

set -x
set -e
set -u

# https://github.com/knative/docs/blob/master/install/Knative-with-GKE.md#creating-a-kubernetes-cluster

cluster=${1}
zone=${2}
project=${3}

gcloud container clusters get-credentials ${cluster} --zone ${zone} --project ${project}

function pod_status_count() {
    kubectl get pods --namespace ${1} | grep ${2} | wc -l
}

function wait_for_namespace() {
    namespace=${1}
    run_count=${2}
    complete_count=${3}

    while true; do

        if [[ "$(pod_status_count ${namespace} Running)" != "${run_count}" ]]; then
            continue
        fi

        if [[ "$(pod_status_count ${namespace} Completed)" != "${complete_count}" ]]; then
            continue
        fi

    break

    done
}

set +e
kubectl get clusterrolebinding cluster-admin-binding &>/dev/null
if [[ $? -ne 0 ]]; then
    kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole=cluster-admin \
        --user=$(gcloud config get-value core/account)
fi
set -e

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.4.0/istio-crds.yaml
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.4.0/istio.yaml

set +e
kubectl get namespace default --output=json | jq -r -S '.metadata.labels | to_entries | .[] | " \(.key)=\(.value)"' 2>/dev/null | grep -o istio-injection=enabled &>/dev/null
if [[ $? -ne 0 ]]; then
    kubectl label namespace default istio-injection=enabled
fi
set -e

wait_for_namespace istio-system 13 1

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.4.0/serving.yaml \
    --filename https://github.com/knative/build/releases/download/v0.4.0/build.yaml \
    --filename https://github.com/knative/eventing/releases/download/v0.4.0/in-memory-channel.yaml \
    --filename https://github.com/knative/eventing/releases/download/v0.4.0/release.yaml \
    --filename https://github.com/knative/eventing-sources/releases/download/v0.4.0/release.yaml \
    --filename https://github.com/knative/serving/releases/download/v0.4.0/monitoring.yaml \
    --filename https://raw.githubusercontent.com/knative/serving/v0.4.0/third_party/config/build/clusterrole.yaml

wait_for_namespace knative-serving 4 0
wait_for_namespace knative-build 2 0
wait_for_namespace knative-eventing 4 0
wait_for_namespace knative-sources 1 0
wait_for_namespace knative-monitoring 9 0
