#!/bin/bash -e

# Description: This script performs a SelfSubjectAccessReview test by asking the k8s apiserver what permissions does each user have
# Tested users: admin@kyma.cx, developer@kyma.cx, user@kyma.cx
# Required ENVS:
#  - ADMIN_EMAIL: email address of the admin user (used as username)
#  - ADMIN_PASSWORD: password for the admin user
#  - DEVELOPER_EMAIL: email address of the developer user (used as username)
#  - DEVELOPER_PASSWORD: password for the developer user
#  - VIEW_EMAIL: email address of the view user (used as username)
#  - VIEW_PASSWORD: password for the view user
#  - NAMESPACE: namespace in which we perform the tests
#  - IS_LOCAL_ENV: boolean indicating tests running on minikube/kind

RETRY_TIME=3 #Seconds
MAX_RETRIES=5

# Helper used to count retry attempts
function __retry() {
    local current="${1}"
    if [[ "${current}" -eq "${MAX_RETRIES}" ]]; then return 1; fi
    echo "---> Retrying in ${RETRY_TIME} seconds..."
    sleep "${RETRY_TIME}"
}

# Helper used to end retrying after max failed attempts
function __failRetry() {
    echo "---> Fatal! Retires failed for: $1"
    exit 1
}

function __createTestBindings() {
	echo "---> $1"
	kubectl create -f ./kyma-test-bindings.yaml -n "${NAMESPACE}"
}

function __deleteTestBindings() {
	echo "---> $1"
	kubectl delete -f ./kyma-test-bindings.yaml -n "${NAMESPACE}"
}

# Retries on errors. Note it is not "clever" and retries even on obvious non-retryable errors.
function createTestBindingsRetry() {
	local MSG="Create test RoleBinding(s)"
	for i in $(seq 1 "${MAX_RETRIES}"); do __createTestBindings "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

# Retries on errors. Note it is not "clever" and retries even on obvious non-retryable errors.
function deleteTestBindingsRetry() {
	local MSG="Delete test RoleBinding(s)"
	for i in $(seq 1 "${MAX_RETRIES}"); do __deleteTestBindings "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function testPermissions() {
	OPERATION="$1"
	RESOURCE="$2"
	TEST_NS="$3"
	EXPECTED="$4"

	set +e
	TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" -n "${TEST_NS}")
	set -e
	if [[ ${TEST} == ${EXPECTED}* ]]; then
		echo "----> PASSED"
		return 0
	fi

	echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"

	# If previous attempt failed (network error?), repeat just one time
	echo "Re-trying one more time..."
	sleep ${RETRY_TIME}

	set +e
	TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" -n "${TEST_NS}")
	set -e
	if [[ ${TEST} == ${EXPECTED}* ]]; then
		echo "----> PASSED"
		return 0
	fi

	echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"
	return 1
}

function __registrationRequest() {
	echo "---> $1"
	curl -k -f -X GET -H 'Content-Type: application/x-www-form-urlencoded' "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM" > registration_request
}

function __loginRequest() {
	local REQUEST_ID="$1"
	echo "---> $2"
	curl -X POST -F "login=${EMAIL}" -F "password=${PASSWORD}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}"
}

# Modifies external APPROVAL_RESPONSE variable!
function __approvalRequest() {
	local REQUEST_ID="$1"
	echo "---> $2"
	APPROVAL_RESPONSE=$(curl -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}")
}

function __configFileRequest() {
	local AUTH_TOKEN="$1"
	echo "---> $2"
	curl -k -f -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}/kube-config" -o "${PWD}/kubeconfig"
}

#Creates a file "registration_request"
function registrationRequestRetry() {
	local MSG="Make registration request"
	for i in $(seq 1 "${MAX_RETRIES}"); do __registrationRequest "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function loginRequestRetry() {
	local REQUEST_ID="$1"
	local MSG="Make login request"
	for i in $(seq 1 "${MAX_RETRIES}"); do __loginRequest "${REQUEST_ID}" "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function approvalRequestRetry() {
	local REQUEST_ID="$1"
	local MSG="Make approval request"
	for i in $(seq 1 "${MAX_RETRIES}"); do __approvalRequest "${REQUEST_ID}" "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function configFileRequestRetry() {
	local AUTH_TOKEN="$1"
	local MSG="Make config file request"
	for i in $(seq 1 "${MAX_RETRIES}"); do __configFileRequest "${AUTH_TOKEN}" "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function getConfigFile() {
	registrationRequestRetry

	local REQUEST_ID
	REQUEST_ID=$(grep '/auth/local?req' < registration_request | cut -d '"' -f 2 | cut -d '?' -f 2)
	rm -f registration_request

	loginRequestRetry "${REQUEST_ID}"

	#APPROVAL_RESPONSE is altered by approvalRequestRetry function!
	APPROVAL_RESPONSE=""
	approvalRequestRetry "${REQUEST_ID}"

	local AUTH_TOKEN
	AUTH_TOKEN=$(echo "${APPROVAL_RESPONSE}" | grep -o -P '(?<=id_token=).*(?=&amp;state)')
	configFileRequestRetry "${AUTH_TOKEN}"

	if [[ ! -s "${PWD}/kubeconfig" ]]; then
		echo "---> KUBECONFIG not created, or is empty!"
		exit 1
	fi
}

function runTests() {
	EMAIL=${DEVELOPER_EMAIL} PASSWORD=${DEVELOPER_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${DEVELOPER_EMAIL} should be able to get Deployments in ${NAMESPACE}"
	testPermissions "get" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to create Deployments in ${NAMESPACE}"
	testPermissions "create" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to get CRD in ${NAMESPACE}"
	testPermissions "get" "crd" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to delete secret in ${NAMESPACE}"
	testPermissions "delete" "secret" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to patch configmap in ${NAMESPACE}"
	testPermissions "patch" "configmap" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to get specific CRD in ${NAMESPACE}"
	testPermissions "get" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to create Access Rules in ${NAMESPACE}"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to delete ClusterRole in ${NAMESPACE}"
	testPermissions "delete" "clusterrole" "${NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to list Deployments in production"
	testPermissions "list" "deployment" "production" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create Services in production"
	testPermissions "create" "service" "production" "no"

	echo "--> ${DEVELOPER_EMAIL} should  be able to get Installation CR in ${NAMESPACE}"
	testPermissions "get" "installation" "${NAMESPACE}" "yes"

 	if [ "${IS_LOCAL_ENV}" = "false" ]; then
            echo "--> ${DEVELOPER_EMAIL} should be able to create certificate"
            testPermissions "create" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"

            echo "--> ${DEVELOPER_EMAIL} should be able to get certificate"
            testPermissions "get" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"

            echo "--> ${DEVELOPER_EMAIL} should be able to patch certificate"
            testPermissions "patch" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"

            echo "--> ${DEVELOPER_EMAIL} should be able to delete certificate"
            testPermissions "delete" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"
	fi

	EMAIL=${ADMIN_EMAIL} PASSWORD=${ADMIN_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${ADMIN_EMAIL} should be able to get ClusterRole"
	testPermissions "get" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete Deployments"
	testPermissions "delete" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete ClusterRole"
	testPermissions "delete" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to get ory Access Rule"
	testPermissions "get" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete ory Access Rule"
	testPermissions "delete" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to create ory Access Rule"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete specific CRD"
	testPermissions "delete" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should  be able to patch Installation CR in ${NAMESPACE}"
	testPermissions "patch" "installation" "${NAMESPACE}" "yes"

	if [ "${IS_LOCAL_ENV}" = "false" ]; then
		echo "--> ${ADMIN_EMAIL} should be able to create certificate"
		testPermissions "create" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"
		
		echo "--> ${ADMIN_EMAIL} should be able to get certificate"
		testPermissions "get" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"
		
		echo "--> ${ADMIN_EMAIL} should be able to patch certificate"
		testPermissions "patch" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"
		
		echo "--> ${ADMIN_EMAIL} should be able to delete certificate"
		testPermissions "delete" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"
	fi

	EMAIL=${VIEW_EMAIL} PASSWORD=${VIEW_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${VIEW_EMAIL} should be able to get ClusterRole"
	testPermissions "get" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> ${VIEW_EMAIL} should be able to list Deployments"
	testPermissions "list" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${VIEW_EMAIL} should NOT be able to create Namespace"
	testPermissions "create" "ns" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to patch pod"
	testPermissions "patch" "pod" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to create secret"
	testPermissions "create" "secret" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to delete ory Access Rule"
	testPermissions "delete" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to create ory Access Rule"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "no"

	if [ "${IS_LOCAL_ENV}" = "false" ]; then
		echo "--> ${VIEW_EMAIL} should be able to get certificate"
		testPermissions "get" "certificate.certmanager.k8s.io" "${NAMESPACE}" "yes"

		echo "--> ${VIEW_EMAIL} should NOT be able to create certificate"
		testPermissions "create" "certificate.certmanager.k8s.io" "${NAMESPACE}" "no"
		
		echo "--> ${VIEW_EMAIL} should NOT be able to delete certificate"
		testPermissions "delete" "certificate.certmanager.k8s.io" "${NAMESPACE}" "no"
		
		echo "--> ${VIEW_EMAIL} should NOT be able to patch certificate"
		testPermissions "patch" "certificate.certmanager.k8s.io" "${NAMESPACE}" "no"
	fi
}


function cleanup() {
	EXIT_STATUS=$?
	unset KUBECONFIG
	if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
		echo "AN ERROR OCCURED! Take a look at preceding log entries."
	fi

	deleteTestBindingsRetry

	local MSG=""
	if [[ ${EXIT_STATUS} -ne 0 ]]; then MSG="(exit status: ${EXIT_STATUS})"; fi
	echo "Job is finished ${MSG}"
	set -e
	exit "${EXIT_STATUS}"
}

discoverUnsetVar=false

for var in ADMIN_EMAIL ADMIN_PASSWORD DEVELOPER_EMAIL DEVELOPER_PASSWORD VIEW_EMAIL VIEW_PASSWORD NAMESPACE; do
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

createTestBindingsRetry
runTests

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"
