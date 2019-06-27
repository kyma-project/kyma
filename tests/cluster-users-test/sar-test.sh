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
	echo "----> |FAIL| Expected: ${EXPECTED} got: ${TEST}"
	return 1
}

function getConfigFile() {
	readonly REGISTRATION_REQUEST=$(curl -s -k -f -X GET -H 'Content-Type: application/x-www-form-urlencoded' "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM")
	readonly REQUEST_ID=$(echo "${REGISTRATION_REQUEST}" | cut -d '"' -f 2 | cut -d '?' -f 2)
	curl -X POST -F "login=${EMAIL}" -F "password=${PASSWORD}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}"
	readonly RESPONSE=$(curl -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}")
	readonly AUTH_TOKEN=$(echo "${RESPONSE}" | grep -o -P '(?<=id_token=).*(?=&amp;state)')
	curl -s -k -f -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}/kube-config" -o "${PWD}/kubeconfig"
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

	echo "--> ${DEVELOPER_EMAIL} should be able to get specific CRD in ${NAMESPACE}"
	testPermissions "get" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to delete ClusterRole in ${NAMESPACE}"
	testPermissions "delete" "clusterrole" "${NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should be able to list Deployments in production"
	testPermissions "list" "deployment" "production" "yes"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create Services in production"
	testPermissions "create" "service" "production" "no"

	EMAIL=${ADMIN_EMAIL} PASSWORD=${ADMIN_PASSWORD} getConfigFile
	echo "--> ${ADMIN_EMAIL} should be able to get ClusterRole"
	testPermissions "get" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete Deployments"
	testPermissions "delete" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete ClusterRole"
	testPermissions "delete" "clusterrole" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete specific CRD"
	testPermissions "delete" "crd/installations.installer.kyma-project.io" "${NAMESPACE}" "yes"

	EMAIL=${VIEW_EMAIL} PASSWORD=${VIEW_PASSWORD} getConfigFile
	echo "--> ${VIEW_EMAIL} should NOT be able to get ClusterRole"
	testPermissions "get" "clusterrole" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to list Deployments"
	testPermissions "list" "deployment" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to create Namespace"
	testPermissions "create" "ns" "${NAMESPACE}" "no"
}

function cleanup() {
	EXIT_STATUS=$?
	unset KUBECONFIG
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

echo "---> Create testing RoleBinding"
kubectl apply -f ./kyma-developer-binding.yaml -n "${NAMESPACE}"
runTests

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"
