#!/usr/bin/env bash

set -o errexit

AUTH_TOKEN=$(/root/app)

getConfigFile() {
	curl -s -H "Authorization: Bearer ${AUTH_TOKEN}" "${CONFIGURATIONS_GENERATOR_SERVICE_HOST}:${CONFIGURATIONS_GENERATOR_SERVICE_PORT_HTTP}/kube-config" -o "${PWD}/kubeconfig"
}

test(){

  if [[ -z "${POD_NAME}" ]]; then
    echo "POD_NAME not provided"
    echo "TEST FAILED"
    exit 1
  fi

  UUID=$(cat /proc/sys/kernel/random/uuid)
  echo ${UUID} > "${PWD}/uuid"
  
  set +e
  NEW_UUID=$(kubectl exec ${POD_NAME} cat ${PWD}/uuid)
  set -e

  if [[ ${UUID} != ${NEW_UUID} ]]; then
    echo "NEW_UUID should be ${UUID}"
    echo "NEW_UUID is ${NEW_UUID}"
    echo "TEST FAILED"
    exit 1
  fi
  echo "TEST SUCCEEDED"
}

getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"
test
