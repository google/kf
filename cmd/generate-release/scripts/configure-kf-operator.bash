#!/usr/bin/env bash

set -eux

project_id=$1
zone=$2
cluster_name=$3

if [ -z "${project_id}" ] || [ -z "${zone}" ] || [ -z "${cluster_name}" ]; then
  echo "usage: $0 [PROJECT_ID] [ZONE] [CLUSTER_NAME]"
  exit 1
fi

# Convert zone into a region
region=$(echo "${zone}" | cut -d'-' -f1,2)

registry=${region}-docker.pkg.dev/${project_id}/${cluster_name}

serviceAccount="${cluster_name}-sa"

# Patch cloudrun-cr
kubectl patch \
  kfsystem kfsystem \
  --type='json' \
  -p="[{'op': 'replace', 'path': '/spec/kf', 'value': {'enabled': true, 'config': {'spaceContainerRegistry': '${registry}', 'secrets':{'workloadidentity':{'googleserviceaccount':'${serviceAccount}', 'googleprojectid':'${project_id}'}}}}}]"

# Waits until all pods are running in the given namespace.
# Parameters: $1 - namespace.
function wait_until_pods_running() {
  echo -n "Waiting until all pods in namespace $1 are up"
  local failed_pod=""
  local previous_ready=-1
  for i in {1..150}; do  # timeout after 5 minutes
    # List all pods. Ignore Terminating pods as those have either been replaced through
    # a deployment or terminated on purpose (through chaosduck for example).
    local pods="$(kubectl get pods --no-headers -n $1 2>/dev/null | grep -v Terminating)"
    # All pods must be running (ignore ImagePull error to allow the pod to retry)
    local not_running_pods=$(echo "${pods}" | grep -v Running | grep -v Completed | grep -v ErrImagePull | grep -v ImagePullBackOff)
    if [[ -n "${pods}" ]] && [[ -z "${not_running_pods}" ]]; then
      # All Pods are running or completed. Verify the containers on each Pod.
      local all_ready=1
      while read pod ; do
        local status=(`echo -n ${pod} | cut -f2 -d' ' | tr '/' ' '`)
        # Set this Pod as the failed_pod. If nothing is wrong with it, then after the checks, set
        # failed_pod to the empty string.
        failed_pod=$(echo -n "${pod}" | cut -f1 -d' ')
        # All containers must be ready
        [[ -z ${status[0]} ]] && all_ready=0 && previous_ready=-1 && break
        [[ -z ${status[1]} ]] && all_ready=0 && previous_ready=-1 && break
        [[ ${status[0]} -lt 1 ]] && all_ready=0 && previous_ready=-1 && break
        [[ ${status[1]} -lt 1 ]] && all_ready=0 && previous_ready=-1 && break
        [[ ${status[0]} -ne ${status[1]} ]] && all_ready=0 && previous_ready=-1 && break
        # All the tests passed, this is not a failed pod.
        failed_pod=""
      done <<< "$(echo "${pods}" | grep -v Completed)"
      if (( all_ready )); then
        echo -e "\nAll pods are up:\n${pods}"
        # previous_ready not set. This previous_ready to current cycle number.
        echo -e "\nCurrent cycle: $i, previous_ready: $previous_ready"
        [[ "$previous_ready" -eq -1 ]] && previous_ready="$i" && sleep 5 && continue
        # i - previous_ready < 10. Cluster hasn't been ready for 10 cycles. Continue checking>
        [[ "$previous_ready" -gt "$i"-10 ]] && sleep 5 && continue
        return 0
      fi
    elif [[ -n "${not_running_pods}" ]]; then
      # At least one Pod is not running, just save the first one's name as the failed_pod.
      failed_pod="$(echo "${not_running_pods}" | head -n 1 | cut -f1 -d' ')"
      previous_ready=-1
    fi
    echo -n "."
    sleep 5
  done
  echo -e "\n\nERROR: timeout waiting for pods to come up\n${pods}"
  if [[ -n "${failed_pod}" ]]; then
    echo -e "\n\nFailed Pod (data in YAML format) - ${failed_pod}\n"
    kubectl -n $1 get pods "${failed_pod}" -oyaml
    echo -e "\n\nPod Logs\n"
    kubectl -n $1 logs "${failed_pod}" --all-containers
  fi
  return 1
}

wait_until_pods_running kf
