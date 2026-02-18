#!/usr/bin/env bash

# XXX: This is only necessary while the operator does not have the
# functionality to install KCC.

set -eux

if [ -z ${1:+x} ] || [ -z ${2:+x} ]; then
  echo "usage: $0 [PROJECT_ID] [CLUSTER_NAME]"
  exit 1
fi

project_id=$1
cluster_name=$2

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
        echo "\"${command}\" failed too many times (${retry_count})"
        exit 1
    fi
}

tempdir="$(mktemp -d)"
trap 'rm -fr $tempdir' EXIT

cd ${tempdir}
kcc_url=$(kf dependencies url 'Config Connector')
gcloud storage cp ${kcc_url} release-bundle.tar.gz
tar zxvf release-bundle.tar.gz
kubectl apply -f operator-system/configconnector-operator.yaml

if [ ! "$(yes | gcloud beta iam service-accounts list --filter "name:${cluster_name}-sa")" ]; then
  echo "creating new GCP service-account"
  yes | gcloud beta iam service-accounts create \
          "${cluster_name}-sa" \
          --description "gcr.io admin for ${cluster_name}" \
          --display-name "${cluster_name}"

else
  echo "using existing GCP service-account"
fi

# Give service account role to access IAM
gcloud projects add-iam-policy-binding "${project_id}" \
  --member "serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" \
  --role "roles/iam.serviceAccountAdmin" \
  --format none

cat <<EOF | kubectl apply -f -
apiVersion: core.cnrm.cloud.google.com/v1beta1
kind: ConfigConnector
metadata:
  name: configconnector.core.cnrm.cloud.google.com
spec:
  mode: cluster
  googleServiceAccount: "${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"
EOF

# XXX: Give everything a moment to be created.
sleep 30

# https://cloud.google.com/config-connector/docs/how-to/advanced-install#verifying_your_installation
retry "kubectl wait -n cnrm-system --for=condition=Ready pod --all --timeout=10s" 40 20

# Workload Identity
# Link GSA with KCC KSA
# https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
# NOTE: This assumes the following:
# * KCC controller is in 'cnrm-system' namespace
# * KCC controller has the KSA 'cnrm-controller-manager'
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${project_id}.svc.id.goog[cnrm-system/cnrm-controller-manager]" \
  --format none \
  "${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"

retry "kubectl annotate serviceaccount \
  --namespace cnrm-system \
  --overwrite \
  cnrm-controller-manager \
  iam.gke.io/gcp-service-account=${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" 20 20
