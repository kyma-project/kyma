#!/bin/bash

function LogInAsUser() {
	local USER="$1"
	local PASS="$2"
	local AUTH_TOKEN

	# Handle registration request
	curl -k -f -X GET -H 'Content-Type: application/x-www-form-urlencoded' "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM" > registration_request
	REQUEST_ID=$(grep '/auth/local?req' < registration_request | cut -d '"' -f 2 | cut -d '?' -f 2)
	rm -f registration_request

	# Handle login request
	curl -X POST -F "login=${USER}" -F "password=${PASS}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}"

	# Handle approval response
	APPROVAL_RESPONSE=$(curl -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}")
	AUTH_TOKEN=$(echo "${APPROVAL_RESPONSE}" | grep -o -P '(?<=id_token=).*(?=&amp;state)')

	# Get kubeconfig
	curl -k -f -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}/kube-config" -o "${PWD}/kubeconfig"
	if [[ ! -s "${PWD}/kubeconfig" ]]; then
		echo "---> KUBECONFIG not created, or is empty!"
		exit 1
	fi
}

function TestForUser() {
	echo "foobar"
}