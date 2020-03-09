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
#  - NAMESPACE_ADMIN_EMAIL: email address of the namespace admin user (used as username)
#  - NAMESPACE_ADMIN_PASSWORD: password for the namespace admin user
#  - NAMESPACE: namespace in which we perform the tests
#  - SYSTEM_NAMESPACE: namespace to which namespace admin should not have access
#  - CUSTOM_NAMESPACE: namespace which will be created by namespace admin

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

function __deleteTestNamespace() {
	echo "---> $1"
	kubectl delete namespace "${CUSTOM_NAMESPACE}"
}

function __createRoleBindingForNamespaceDeveloper() {
	local result=1
	set +e
	kubectl create rolebinding 'namespace-developer' --clusterrole='kyma-developer' --user="${DEVELOPER_EMAIL}" -n "${CUSTOM_NAMESPACE}"
	result=$?
	set -e

	if [[ ${result} -eq 0 ]]; then
		echo "----> PASSED"
		return 0
	fi

	echo "----> |FAIL|"
	return 1
}

function __createNamespaceForNamespaceAdmin() {
	local result=1
	set +e
	kubectl create namespace "${CUSTOM_NAMESPACE}"
	result=$?
	set -e

	if [[ ${result} -eq 0 ]]; then
		echo "----> PASSED"
		return 0
	fi

	echo "----> |FAIL|"
	return 1
}

function __testPermissions() {
	local OPERATION="$1"
	local RESOURCE="$2"
	local TEST_NS="$3"
	local EXPECTED="$4"
	local TEST="not-set-yet"

	if [[ "${TEST_NS}" != "--all-namespaces" ]]; then
		TEST_NS="-n${TEST_NS}"
	fi

	sleep 0.1

	set +e
	TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" "${TEST_NS}")
	set -e
	if [[ "${TEST}" == "${EXPECTED}" ]]; then
		echo "----> PASSED"
		return 0
	fi

	echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"
	return 1
}

function __testPermissionsClusterScoped() {
	local OPERATION="$1"
	local RESOURCE="$2"
	local EXPECTED="$3"
  __testPermissions "${OPERATION}" "${RESOURCE}" --all-namespaces "${EXPECTED}"
}

function __testDescribe() {
	local RESOURCE="$1"
	local TEST_NS="$2"
	local EXPECTED="$3"

	if [[ "${TEST_NS}" != "--all-namespaces" ]]; then
		TEST_NS="-n${TEST_NS}"
	fi

	sleep 0.1

	local result=1
	set +e
	kubectl describe "${RESOURCE}" "${TEST_NS}" > /dev/null
	result=$?
	set -e

	local IS_OK="false"

	if [[ "${EXPECTED}" == "yes" ]] && [[ ${result} -eq 0 ]]; then
		IS_OK="true"
	fi
	if [[ "${EXPECTED}" == "no" ]] && [[ ${result} -ne 0 ]]; then
		IS_OK="true"
	fi

	if [[ "${IS_OK}" == "true" ]]; then
			echo "----> PASSED"
			return 0
	fi

	echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"
	return 1
}

function __testDescribeClusterScoped() {
	local RESOURCE="$1"
	local EXPECTED="$2"

	__testDescribe "${RESOURCE}" --all-namespaces "${EXPECTED}"
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

# Retries on errors. Note it is not "clever" and retries even on obvious non-retryable errors.
function deleteTestNamespaceRetry() {
	local MSG="Delete test Namespace created by Namespace Admin"
	for i in $(seq 1 "${MAX_RETRIES}"); do __deleteTestNamespace "${MSG}" && break || __retry "${i}" || __failRetry "${MSG}" ; done
}

function createRoleBindingForNamespaceDeveloper() {
	__createRoleBindingForNamespaceDeveloper || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __createRoleBindingForNamespaceDeveloper || return 1)
}

function createNamespaceForNamespaceAdmin() {
	__createNamespaceForNamespaceAdmin || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __createNamespaceForNamespaceAdmin || return 1)
}

function testPermissions() {
	__testPermissions "$@" || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __testPermissions "$@" || return 1)
}

function testPermissionsClusterScoped() {
	__testPermissionsClusterScoped "$@" || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __testPermissionsClusterScoped "$@" || return 1)
}

function testDescribe() {
	__testDescribe "$@" || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __testDescribe "$@" || return 1)
}

function testDescribeClusterScoped() {
	__testDescribeClusterScoped "$@" || (echo "Re-trying one more time..." && sleep ${RETRY_TIME} && __testDescribeClusterScoped "$@" || return 1)
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

function testRafter() {
	local -r USER_EMAIL="${1}"
	local -r TEST_NAMESPACE="${2}"

	echo "--> ${USER_EMAIL} should be able to get AssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "get" "assetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create AssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "create" "assetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete AssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "assetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch AssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "assetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should be able to get ClusterAssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "get" "clusterassetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create ClusterAssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "create" "clusterassetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete ClusterAssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "clusterassetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch ClusterAssetGroup CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "clusterassetgroup.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should be able to get Asset CR in ${TEST_NAMESPACE}"
	testPermissions "get" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create Asset CR in ${TEST_NAMESPACE}"
	testPermissions "create" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete Asset CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch Asset CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should be able to get ClusterAsset CR in ${TEST_NAMESPACE}"
	testPermissions "get" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create ClusterAsset CR in ${TEST_NAMESPACE}"
	testPermissions "create" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete ClusterAsset CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch ClusterAsset CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "asset.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should be able to get Bucket CR in ${TEST_NAMESPACE}"
	testPermissions "get" "bucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create Bucket CR in ${TEST_NAMESPACE}"
	testPermissions "create" "bucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete Bucket CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "bucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch Bucket CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "bucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should be able to get ClusterBucket CR in ${TEST_NAMESPACE}"
	testPermissions "get" "clusterbucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "yes"

	echo "--> ${USER_EMAIL} should NOT be able to create ClusterBucket CR in ${TEST_NAMESPACE}"
	testPermissions "create" "clusterbucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to delete ClusterBucket CR in ${TEST_NAMESPACE}"
	testPermissions "delete" "clusterbucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"

	echo "--> ${USER_EMAIL} should NOT be able to patch ClusterBucket CR in ${TEST_NAMESPACE}"
	testPermissions "patch" "clusterbucket.rafter.kyma-project.io" "${TEST_NAMESPACE}" "no"
}

function runTests() {
	EMAIL=${ADMIN_EMAIL} PASSWORD=${ADMIN_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${ADMIN_EMAIL} should be able to get ClusterRole"
	testPermissionsClusterScoped "get" "clusterrole" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete ClusterRole"
	testPermissionsClusterScoped "delete" "clusterrole" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete Deployments"
	testPermissions "delete" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to get ory Access Rule"
	testPermissions "get" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete ory Access Rule"
	testPermissions "delete" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to create ory Access Rule"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to list applicationmappings.applicationconnector.kyma-project.io"
	testPermissions "list" "applicationmappings.applicationconnector.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to create applicationmapping.applicationconnector.kyma-project.io"
	testPermissions "create" "applicationmapping.applicationconnector.kyma-project.io" "${NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to get applications.applicationconnector.kyma-project.io in the cluster"
	testPermissionsClusterScoped "get" "applications.applicationconnector.kyma-project.io" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to list applications.applicationconnector.kyma-project.io in the cluster"
	testPermissionsClusterScoped "list" "applications.applicationconnector.kyma-project.io" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to watch applications.applicationconnector.kyma-project.io in the cluster"
	testPermissionsClusterScoped "watch" "applications.applicationconnector.kyma-project.io" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to delete specific CRD in the cluster"
	testPermissionsClusterScoped "delete" "crd/installations.installer.kyma-project.io" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to patch Installation CR in ${NAMESPACE}"
	testPermissions "patch" "installation" "${NAMESPACE}" "yes"

	testRafter "${ADMIN_EMAIL}" "${NAMESPACE}"

	echo "--> ${ADMIN_EMAIL} should be able to delete any namespace in the cluster"
	testPermissionsClusterScoped "delete" "namespace" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to get Nodes in the cluster"
	testPermissionsClusterScoped "get" "nodes" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to describe Pods in ${CUSTOM_NAMESPACE}"
	testDescribe "pods" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to describe Pods in ${SYSTEM_NAMESPACE}"
	testDescribe "pods" "${SYSTEM_NAMESPACE}" "yes"

	echo "--> ${ADMIN_EMAIL} should be able to describe Nodes in the cluster"
	testDescribeClusterScoped "nodes" "yes"

	EMAIL=${VIEW_EMAIL} PASSWORD=${VIEW_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${VIEW_EMAIL} should be able to get ClusterRole"
	testPermissionsClusterScoped "get" "clusterrole" "yes"

	echo "--> ${VIEW_EMAIL} should be able to list Deployments"
	testPermissions "list" "deployment" "${NAMESPACE}" "yes"

	echo "--> ${VIEW_EMAIL} should NOT be able to create Namespace"
	testPermissionsClusterScoped "create" "ns" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to patch pod"
	testPermissions "patch" "pod" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to create secret"
	testPermissions "create" "secret" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to delete ory Access Rule"
	testPermissions "delete" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "no"

	echo "--> ${VIEW_EMAIL} should NOT be able to create ory Access Rule"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${NAMESPACE}" "no"

	testRafter "${VIEW_EMAIL}" "${NAMESPACE}"

	EMAIL=${NAMESPACE_ADMIN_EMAIL} PASSWORD=${NAMESPACE_ADMIN_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create new namespace"
	createNamespaceForNamespaceAdmin
	export SHOULD_CLEANUP_NAMESPACE="true"
	sleep 10 #To ensure rolebinding is in effect

	# namespace admin should not be able to get or create any resource in system namespaces
	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to list Deployments in system namespace"
	testPermissions "list" "deployment" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to get ory Access Rule in system namespace"
	testPermissions "get" "rule.oathkeeper.ory.sh" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to create secret in system namespace"
	testPermissions "create" "secret" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to delete system namespace"
	testPermissions "delete" "namespace" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to create clusterrolebindings"
	testPermissionsClusterScoped "create" "clusterrolebinding" "no"

	# namespace admin should be able to get/list/create/delete k8s and kyma resources in the namespace they created
  echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list Deployments in the namespace they created"
	testPermissions "list" "deployments" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create Deployment in the namespace they created"
	testPermissions "create" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update Deployment in the namespace they created"
	testPermissions "update" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get Deployment in the namespace they created"
	testPermissions "get" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete Deployment in the namespace they created"
	testPermissions "delete" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list Pods in the namespace they created"
	testPermissions "list" "pods" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create Pod in the namespace they created"
	testPermissions "create" "pod" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update Pod in the namespace they created"
	testPermissions "update" "pod" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get Pod in the namespace they created"
	testPermissions "get" "pod" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete Pod in the namespace they created"
	testPermissions "delete" "pod" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list ConfigMaps in the namespace they created"
	testPermissions "list" "configmaps" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create ConfigMap in the namespace they created"
	testPermissions "create" "configmap" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update ConfigMap in the namespace they created"
	testPermissions "update" "configmap" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get ConfigMap in the namespace they created"
	testPermissions "get" "configmap" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete ConfigMap in the namespace they created"
	testPermissions "delete" "configmap" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list ORY Access Rules in the namespace they created"
	testPermissions "list" "rules.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create ORY Access Rule in the namespace they created"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update ORY Access Rule in the namespace they created"
	testPermissions "update" "rule.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get ORY Access Rule in the namespace they created"
	testPermissions "get" "rule.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete ORY Access Rule in the namespace they created"
	testPermissions "delete" "rule.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list Secrets in the namespace they created"
	testPermissions "list" "secrets" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create Secret in the namespace they created"
	testPermissions "create" "secret" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update Secret in the namespace they created"
	testPermissions "update" "secret" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get Secret in the namespace they created"
	testPermissions "get" "secret" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete Secret in the namespace they created"
	testPermissions "delete" "secret" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list Roles in the namespace they created"
	testPermissions "list" "roles.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create Role in the namespace they created"
	testPermissions "create" "role.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get Role in the namespace they created"
	testPermissions "get" "role.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update Role in the namespace they created"
	testPermissions "update" "role.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete Role in the namespace they created"
	testPermissions "delete" "role.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list RoleBindings in the namespace they created"
	testPermissions "list" "rolebindings.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create RoleBinding in the namespace they created"
	testPermissions "create" "rolebinding.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get RoleBinding in the namespace they created"
	testPermissions "get" "rolebinding.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update RoleBinding in the namespace they created"
	testPermissions "update" "rolebinding.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete RoleBinding in the namespace they created"
	testPermissions "delete" "rolebinding.rbac.authorization.k8s.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create rolebindings to kyma-developer clusterrole in the namespace they created"
	createRoleBindingForNamespaceDeveloper

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to delete arbitrary namespaces in the cluster"
	testPermissionsClusterScoped "delete" "namespace" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to delete system namespace"
	testPermissions "delete" "namespace" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete namespace they created"
	testPermissions "delete" "namespace" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "get" "addonsconfigurations.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "list" "addonsconfigurations.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to watch addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "watch" "addonsconfigurations.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create addonsconfiguration.addons.kyma-project.io in the namespace they created"
	testPermissions "create" "addonsconfiguration.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "update" "addonsconfigurations.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "delete" "addonsconfigurations.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	testRafter "${NAMESPACE_ADMIN_EMAIL}" "${CUSTOM_NAMESPACE}"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "get" "addonsconfigurations/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "list" "addonsconfigurations/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to watch addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "watch" "addonsconfigurations/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create addonsconfiguration.addons.kyma-project.io in the namespace they created"
	testPermissions "create" "addonsconfiguration/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "update" "addonsconfigurations/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "delete" "addonsconfigurations/status.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to get addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "get" "addonsconfigurations/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to list addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "list" "addonsconfigurations/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to watch addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "watch" "addonsconfigurations/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to create addonsconfiguration.addons.kyma-project.io in the namespace they created"
	testPermissions "create" "addonsconfiguration/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to update addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "update" "addonsconfigurations/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to delete addonsconfigurations.addons.kyma-project.io in the namespace they created"
	testPermissions "delete" "addonsconfigurations/finalizers.addons.kyma-project.io" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should be able to describe Pods in ${CUSTOM_NAMESPACE}"
	testDescribe "pods" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to describe Pods in ${SYSTEM_NAMESPACE}"
	testDescribe "pods" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${NAMESPACE_ADMIN_EMAIL} should NOT be able to describe Nodes in the cluster"
	testDescribeClusterScoped "nodes" "no"

	# developer who was granted kyma-developer role should be able to operate in the scope of its namespace
	EMAIL=${DEVELOPER_EMAIL} PASSWORD=${DEVELOPER_PASSWORD} getConfigFile
	export KUBECONFIG="${PWD}/kubeconfig"

	echo "--> ${DEVELOPER_EMAIL} should be able to get Deployments in ${CUSTOM_NAMESPACE}"
	testPermissions "get" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to create Deployments in ${CUSTOM_NAMESPACE}"
	testPermissions "create" "deployment" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to list CRD in the cluster"
	testPermissionsClusterScoped "list" "crd" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to get CRD in the cluster"
	testPermissionsClusterScoped "get" "crd" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to get specific CRD in the cluster"
	testPermissionsClusterScoped "get" "crd/installations.installer.kyma-project.io" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to delete secret in ${CUSTOM_NAMESPACE}"
	testPermissions "delete" "secret" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to patch configmap in ${CUSTOM_NAMESPACE}"
	testPermissions "patch" "configmap" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should be able to create Access Rules in ${CUSTOM_NAMESPACE}"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to delete ClusterRole in the cluster"
	testPermissionsClusterScoped "delete" "clusterrole" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to delete Role in ${CUSTOM_NAMESPACE}"
	testPermissions "delete" "role" "${CUSTOM_NAMESPACE}" "no"

	# developer who was granted kyma-developer role should not be able to operate in system namespaces
	echo "--> ${DEVELOPER_EMAIL} should NOT be able to list Deployments in system namespace"
	testPermissions "list" "deployment" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to get Pods in system namespace"
	testPermissions "get" "pod" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create ory Access Rule in system namespace"
	testPermissions "create" "rule.oathkeeper.ory.sh" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create secret in system namespace"
	testPermissions "create" "secret" "${SYSTEM_NAMESPACE}" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create clusterrolebindings in the cluster"
	testPermissionsClusterScoped "create" "clusterrolebinding" "no"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to create rolebindings in system namespace"
	testPermissions "create" "rolebinding" "${SYSTEM_NAMESPACE}" "no"

	testRafter "${DEVELOPER_EMAIL}" "${SYSTEM_NAMESPACE}"

	echo "--> ${DEVELOPER_EMAIL} should be able to describe Pods in ${CUSTOM_NAMESPACE}"
	testDescribe "pods" "${CUSTOM_NAMESPACE}" "yes"

	echo "--> ${DEVELOPER_EMAIL} should NOT be able to describe Pods in ${SYSTEM_NAMESPACE}"
	testDescribe "pods" "${SYSTEM_NAMESPACE}" "no"
}

function cleanup() {
	EXIT_STATUS=$?
	unset KUBECONFIG
	if [ "${ERROR_LOGGING_GUARD}" = "true" ]; then
		echo "AN ERROR OCCURED! Take a look at preceding log entries."
	fi

	if [ "${SHOULD_CLEANUP_NAMESPACE}" = "true" ]; then
		deleteTestNamespaceRetry
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
