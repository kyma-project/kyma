#!/usr/bin/env bash

set -eu -o pipefail

readonly KSVC_RESOURCE_TYPE="services.serving.knative.dev"
readonly KSVC_NAMESPACE="kyma-integration"

function delete_orphaned_ksvcs() {
  local -r ksvcs=$(kubectl get "${KSVC_RESOURCE_TYPE}" -n "${KSVC_NAMESPACE}" -ojson | jq -c ".items[]")
  if [[ -z ${ksvcs} ]]; then
    echo "There are not any Knative Services in namespace ${KSVC_NAMESPACE}. Skipping..."
    return 0
  fi

  IFS=$'\n'
  for ksvc in ${ksvcs}
  do
    does_belong_to_httpsource=$(echo "${ksvc}" | jq  '.metadata.ownerReferences | .[] | select(.kind=="HTTPSource")' )
    if [[ ! -z "${does_belong_to_httpsource}" ]]; then
        ksvc_name=$(echo "${ksvc}" | jq -r '.metadata.name' )
        echo "Found a ksvc with name ${ksvc_name} which belongs to HTTPSource ... deleting ..."
        kubectl delete -n "${KSVC_NAMESPACE}" "${KSVC_RESOURCE_TYPE}"  "${ksvc_name}"
    fi
  done
}

delete_orphaned_ksvcs
