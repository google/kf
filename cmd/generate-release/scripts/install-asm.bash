#!/usr/bin/env bash

# This is a script written based on
# https://cloud.google.com/service-mesh/docs/gke-install-existing-cluster
#
# NOTE: It uses a local and altered version of the script. There are more
# notes in the script as to what was changed.

set -ex

PROJECT_ID=$1
CLUSTER_NAME=$2
CLUSTER_LOCATION=$3

if
    [ -z "${PROJECT_ID}" ] ||
        [ -z "${CLUSTER_NAME}" ] ||
        [ -z "${CLUSTER_LOCATION}" ]
then
    echo "usage: $0 [PROJECT_ID] [CLOUDSDK_CONTAINER_CLUSTER] [CLOUDSDK_COMPUTE_ZONE]"
    exit 1
fi

# Necessary to download ASM tarball
apt-get update -y
apt-get install -y jq

# Install gcloud CLI only if it hasn't already been provided
if ! command -v gcloud &> /dev/null
then
    apt-get install -y gcloud
else 
    echo "gcloud CLI is already available, skipping installation"
fi

curl https://storage.googleapis.com/csm-artifacts/asm/asmcli_1.16 >asmcli
chmod +x asmcli

gcloud container clusters get-credentials "${CLUSTER_NAME}" \
    --project="${PROJECT_ID}" \
    --zone="${CLUSTER_LOCATION}"

kubectl create namespace asm-gateways

if [ "${ASM_MANAGED:-false}" = "true" ]; then
    ./asmcli install \
        --project_id "${PROJECT_ID}" \
        --cluster_name "${CLUSTER_NAME}" \
        --cluster_location "${CLUSTER_LOCATION}" \
        --enable_all \
        --managed \
        --ca mesh_ca \
        --output_dir out

    kubectl label namespace asm-gateways istio-injection- istio.io/rev="asm-managed" --overwrite
else
    ./asmcli install \
        --project_id "${PROJECT_ID}" \
        --cluster_name "${CLUSTER_NAME}" \
        --cluster_location "${CLUSTER_LOCATION}" \
        --enable_all \
        --ca mesh_ca \
        --output_dir out

    REVISION=$(kubectl get deploy -n istio-system -l app=istiod -o json |
        jq '.items[].metadata.labels."istio.io/rev"' -r)
    kubectl label namespace asm-gateways istio-injection- istio.io/rev="$REVISION" --overwrite
fi

# Not everything in theis folder is applicable, some files are only for certain versions of K8s
# so a blanket -f won't work:
# https://github.com/GoogleCloudPlatform/anthos-service-mesh-packages/tree/1.16.4-asm.2+config1/samples/gateways/istio-ingressgateway
kubectl apply -n asm-gateways \
    -f out/samples/gateways/istio-ingressgateway/deployment.yaml \
    -f out/samples/gateways/istio-ingressgateway/autoscalingv2/autoscaling-v2.yaml \
    -f out/samples/gateways/istio-ingressgateway/pdb-v1.yaml \
    -f out/samples/gateways/istio-ingressgateway/role.yaml \
    -f out/samples/gateways/istio-ingressgateway/service.yaml \
    -f out/samples/gateways/istio-ingressgateway/serviceaccount.yaml