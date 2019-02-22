#!/bin/bash -e

# Description: This script performs a SelfSubjectAccessReview test by asking the k8s apiserver what permissions does each user have
# Tested users: admin@kyma.cx, developer@kyma.cx, user@kyma.cx
# Required ENVS: 
#  - EMAIL_FILE: path to a file with the email address of the user (used as username)
#  - PASSWORD_FILE: path to a file with the password for the user

function testPermissions() {
	USER="$1"
	OPERATION="$2"
	RESOURCE="$3"
	EXPECTED="$4"
	set +e
	TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" --as "${USER}")
	set -e
	if [[ "${TEST}" == "${EXPECTED}" ]]; then
		echo "----> PASSED"
		return 0
	fi
	echo "----> |FAIL| Expected result: ${EXPECTED} vs Actual result: ${TEST}"
	return 1
}

function runTests() {
	echo "--> developer@kyma.cx should be able to get Deployments"
	testPermissions "developer@kyma.cx" "get" "deploy" "yes"

	echo "--> developer@kyma.cx should be able to create Deployments"
	testPermissions "developer@kyma.cx" "create" "deploy" "yes"

	echo "--> developer@kyma.cx should be able to get CRD"
	testPermissions "developer@kyma.cx" "get" "crd" "yes"

	echo "--> developer@kyma.cx should be able to get specific CRD"
	testPermissions "developer@kyma.cx" "get" "crd/installations.installer.kyma-project.io" "yes"

	echo "--> developer@kyma.cx should NOT be able to list ClusterRole"
	testPermissions "developer@kyma.cx" "list" "clusterrole" "no"


	echo "--> admin@kyma.cx should be able to get ClusterRole"
	testPermissions "admin@kyma.cx" "get" "clusterrole" "yes"

	echo "--> admin@kyma.cx should be able to delete Deployments"
	testPermissions "admin@kyma.cx" "delete" "deploy" "yes"

	echo "--> admin@kyma.cx should be able to delete ClusterRole"
	testPermissions "admin@kyma.cx" "delete" "clusterrole" "yes"

	echo "--> admin@kyma.cx should be able to delete specific CRD"
	testPermissions "admin@kyma.cx" "delete" "crd/installations.installer.kyma-project.io" "yes"

	echo "--> user@kyma.cx should NOT be able to get ClusterRole"
	testPermissions "user@kyma.cx" "get" "clusterrole" "no"

	echo "--> user@kyma.cx should NOT be able to list Deployments"
	testPermissions "user@kyma.cx" "list" "deploy" "no"

	echo "--> user@kyma.cx should NOT be able to create Namespace"
	testPermissions "user@kyma.cx" "create" "ns" "no"
}

function init() {
	readonly REGISTRATION_REQUEST=$(curl -s -X GET -H 'Content-Type: application/x-www-form-urlencoded' "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM")
	readonly REQUEST_ID=$(echo ${REGISTRATION_REQUEST} | cut -d '"' -f 2 | cut -d '?' -f 2)
	readonly EMAIL=$(cat ${EMAIL_FILE})
	readonly PASSWORD=$(cat ${PASSWORD_FILE})
	curl -X POST -F "login=${EMAIL}" -F "password=${PASSWORD}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}"
	readonly RESPONSE=$(curl -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}")
	export AUTH_TOKEN=$(echo "${RESPONSE}" | grep -o -P '(?<=id_token=).*(?=&amp;state)')
	curl -s -H "Authorization: Bearer ${AUTH_TOKEN}" ${CONFIGURATIONS_GENERATOR_SERVICE_HOST}:${CONFIGURATIONS_GENERATOR_SERVICE_PORT_HTTP}/kube-config -o "${PWD}/kubeconfig"
	export KUBECONFIG="${PWD}/kubeconfig"
	kubectl version
}

discoverUnsetVar=false

for var in EMAIL_FILE PASSWORD_FILE ; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

init
runTests
echo "ALL TESTS PASSED"
