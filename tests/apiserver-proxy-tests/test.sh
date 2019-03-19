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
  KUBECTL_OUT=$(kubectl exec ${POD_NAME} cat ${PWD}/uuid)
  KUBECTL_ERR=$((kubectl exec ${POD_NAME} cat ${PWD}/uuid) 2>&1 1>/dev/null)
  set -e

  if [[ -n "${KUBECTL_ERR}" ]]; then
  echo "kubectl exec output:"
  echo "${KUBECTL_OUT}"
  echo "kubectl exec error:"
  echo "${KUBECTL_ERR}"

  echo "TEST FAILED"
  exit 1
  fi

  if [[ "${UUID}" != "${KUBECTL_OUT}" ]]; then
    echo "KUBECTL_OUT should be ${UUID}"
    echo "KUBECTL_OUT is ${KUBECTL_OUT}"
    echo "TEST FAILED"
    exit 1
  fi
  echo "TEST SUCCEEDED"
}

getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"
test
