#!/bin/bash -e

# Description: This script performs a SelfSubjectAccessReview test by asking the k8s apiserver what permissions does each user have
# Tested users: admin@kyma.cx, developer@kyma.cx, user@kyma.cx
# Required ENVS: 
#  - EMAIL_FILE: path to a file with the email address of the user (used as username)
#  - PASSWORD_FILE: path to a file with the password for the user
#  - NAMESPACE: namespace in which we perform the tests

function testPermissions() {
	USER="$1"
	OPERATION="$2"
	RESOURCE="$3"
	TEST_NS="$4"
	EXPECTED="$5"
	set +e
	TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" --as "${USER} -n ${TEST_NS}")
	set -e
	if [[ "${TEST}" == "${EXPECTED}" ]]; then
		echo "----> PASSED"
		return 0
	fi
	echo "----> |FAIL| Expected result: ${EXPECTED} vs Actual result: ${TEST}"
	return 1
}

function runTests() {
	echo "--> developer@kyma.cx should be able to get Deployments in ${NAMESPACE}"
	testPermissions "developer@kyma.cx" "get" "deploy" "${NAMESPACE}" "yes"

	echo "--> developer@kyma.cx should be able to create Deployments in ${NAMESPACE}"
	testPermissions "developer@kyma.cx" "create" "deploy" "${NAMESPACE}" "yes"

	echo "--> developer@kyma.cx should be able to get CRD in ${NAMESPACE}"
	testPermissions "developer@kyma.cx" "get" "crd" "${NAMESPACE}" "yes"

	echo "--> developer@kyma.cx should be able to get specific CRD in ${NAMESPACE}"
	testPermissions "developer@kyma.cx" "get" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> developer@kyma.cx should NOT be able to delete ClusterRole in ${NAMESPACE}"
	testPermissions "developer@kyma.cx" "delete" "clusterrole" "${NAMESPACE}" "no"

	echo "--> developer@kyma.cx should NOT be able to list Deployments in production"
	testPermissions "developer@kyma.cx" "list" "clusterrole" "production" "no"

	echo "--> developer@kyma.cx should NOT be able to create Services in production"
	testPermissions "developer@kyma.cx" "create" "service" "production" "no"

	echo "--> admin@kyma.cx should be able to get ClusterRole"
	testPermissions "admin@kyma.cx" "get" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> admin@kyma.cx should be able to delete Deployments"
	testPermissions "admin@kyma.cx" "delete" "deploy" "${NAMESPACE}" "yes"

	echo "--> admin@kyma.cx should be able to delete ClusterRole"
	testPermissions "admin@kyma.cx" "delete" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> admin@kyma.cx should be able to delete specific CRD"
	testPermissions "admin@kyma.cx" "delete" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> user@kyma.cx should NOT be able to get ClusterRole"
	testPermissions "user@kyma.cx" "get" "clusterrole" "${NAMESPACE}" "no"

	echo "--> user@kyma.cx should NOT be able to list Deployments"
	testPermissions "user@kyma.cx" "list" "deploy" "${NAMESPACE}" "no"

	echo "--> user@kyma.cx should NOT be able to create Namespace"
	testPermissions "user@kyma.cx" "create" "ns" "${NAMESPACE}" "no"
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

	echo "---> Create testing RoleBinding"
	kubectl create -f ./kyma-developer-binding.yaml -n "${NAMESPACE}"
}

function cleanup() {
	EXIT_STATUS=$?

    if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
        echo "AN ERROR OCCURED! Take a look at preceding log entries."
    fi
    echo "---> Deleting bindings for tests"
    kubectl delete -f ./kyma-developer-binding.yaml -n "${NAMESPACE}"
    MSG=""
    if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
    echo "Job is finished ${MSG}"
    set -e
    exit "${EXIT_STATUS}"
}

discoverUnsetVar=false

for var in EMAIL_FILE PASSWORD_FILE NAMESPACE; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

trap cleanup EXIT
ERROR_LOGGING_GUARD="true"

init
runTests

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"
