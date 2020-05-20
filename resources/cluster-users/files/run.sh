set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source "${DIR}/library.sh"

# resources/cluster-users/values.yaml clusterRoles.verbs.view
readonly VIEW_OPERATIONS=( "get" "list" )
# resources/cluster-users/values.yaml clusterRoles.verbs.edit - clusterRoles.verbs.view
readonly EDIT_OPERATIONS=( "create" "delete" "deletecollection" "patch" "update" "watch" )

readonly K8S_RESOURCES=( "deployment" "pod" "secret" "cm" "serviceaccount" "role" )
readonly ORY_RESOURCES=( "oauth2clients.hydra.ory.sh" "rules.oathkeeper.ory.sh" )
readonly KYMA_RESOURCES=( "installations.installer.kyma-project.io" "apirules.gateway.kyma-project.io" "applications.applicationconnector.kyma-project.io" )

trap cleanup EXIT
ERROR_LOGGING_GUARD="true"

CreateBindings

echo "---> Run tests for ${ADMIN_EMAIL}"
LogInAsUser $ADMIN_EMAIL $ADMIN_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${ADMIN_EMAIL}
testComponent $USER $NAMESPACE "yes" "yes" "${K8S_RESOURCES[@]}"
testComponent $USER $NAMESPACE "yes" "yes" "${ORY_RESOURCES[@]}"
testComponent $USER $NAMESPACE "yes" "yes" "${KYMA_RESOURCES[@]}"

echo "---> Run tests for ${DEVELOPER_EMAIL}"
LogInAsUser $DEVELOPER_EMAIL $DEVELOPER_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${DEVELOPER_EMAIL}
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${K8S_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${ORY_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${KYMA_RESOURCES[@]}"

echo "---> Run tests for ${VIEW_EMAIL}"
LogInAsUser $VIEW_EMAIL $VIEW_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${VIEW_EMAIL}
testComponent $USER $NAMESPACE "yes" "no" "${K8S_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "no" "${K8S_RESOURCES[@]}"
testComponent $USER $NAMESPACE "yes" "no" "${ORY_RESOURCES[@]}"
testComponent $USER $NAMESPACE "yes" "no" "${KYMA_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "no" "${KYMA_RESOURCES[@]}"

echo "---> Run tests for ${NAMESPACE_ADMIN_EMAIL}"
LogInAsUser $NAMESPACE_ADMIN_EMAIL $NAMESPACE_ADMIN_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${NAMESPACE_ADMIN_EMAIL}
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${K8S_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${ORY_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${KYMA_RESOURCES[@]}"

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"