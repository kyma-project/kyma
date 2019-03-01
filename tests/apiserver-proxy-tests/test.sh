#!/usr/bin/env bash

set -o errexit

getConfigFile() {
	readonly REGISTRATION_REQUEST=$(curl -s -X GET -H 'Content-Type: application/x-www-form-urlencoded' "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM")
	readonly REQUEST_ID=$(echo "${REGISTRATION_REQUEST}" | cut -d '"' -f 2 | cut -d '?' -f 2)
	readonly EMAIL=$(cat "${EMAIL_FILE}")
	readonly PASSWORD=$(cat "${PASSWORD_FILE}")
	curl -X POST -F "login=${EMAIL}" -F "password=${PASSWORD}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}"
	readonly RESPONSE=$(curl -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}")
	readonly AUTH_TOKEN=$(echo "${RESPONSE}" | grep -o -P '(?<=id_token=).*(?=&amp;state)')
	curl -s -H "Authorization: Bearer ${AUTH_TOKEN}" "${CONFIGURATIONS_GENERATOR_SERVICE_HOST}:${CONFIGURATIONS_GENERATOR_SERVICE_PORT_HTTP}/kube-config" -o "${PWD}/kubeconfig"
}

test(){
  UUID=$(cat /proc/sys/kernel/random/uuid)
  echo ${UUID} > "${PWD}/uuid"
  
  set +e
  NEW_UUID=$(kubectl exec test-proxy cat ${PWD}/uuid)
  set -e

  if [[ ${UUID} != ${NEW_UUID} ]]; then
    echo "TEST FAILED"
    exit 1
  fi
  echo "TEST SUCCEEDED"
}

getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"
test
