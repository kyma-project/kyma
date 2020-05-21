set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source "${DIR}/library.sh"

# resources/cluster-users/values.yaml clusterRoles.verbs.view
readonly VIEW_OPERATIONS=( "get" "list" )
# resources/cluster-users/values.yaml clusterRoles.verbs.edit - clusterRoles.verbs.view
readonly EDIT_OPERATIONS=( "create" "delete" "deletecollection" "patch" "update" "watch" )

readonly COMMON_RESOURCES=(
  "deployment"
  "pod"
  "secret"
  "configmap"
  "serviceaccount"
 # Ory resources
  "oauth2clients.hydra.ory.sh"
  "rule.oathkeeper.ory.sh"
 # Kyma resources
  "installations.installer.kyma-project.io"
  "apirules.gateway.kyma-project.io"
  "applicationmappings.applicationconnector.kyma-project.io"
  "servicebindingusages.servicecatalog.kyma-project.io"
)

readonly NAMESPACE_ADMIN_RESOURCES=(
  "role"
  "rolebinding"
 # Service catalog resources
  "serviceinstances.servicecatalog.k8s.io"
  "servicebindings.servicecatalog.k8s.io"
)

readonly NAMESPACE_ADMIN_CLUSTER_SCOPE=(
  "namespace"
#  "nodes"
  "clusterrole"
  "clusterrolebinding"
  "crd"
 # Kyma resources
  "applications.applicationconnector.kyma-project.io"
  "usagekinds.servicecatalog.kyma-project.io"
)

readonly ADMIN_RESOURCES=(
 # Knative serving resources
  "services.serving.knative.dev"
  "routes.serving.knative.dev"
  "revisions.serving.knative.dev"
  "configurations.serving.knative.dev"
  "podautoscalers.autoscaling.internal.knative.dev"
  "images.caching.internal.knative.dev"
 # Rafter resources
  "assetgroup.rafter.kyma-project.io"
  "asset.rafter.kyma-project.io"
  "bucket.rafter.kyma-project.io"
 # Addons
  "addonsconfigurations.addons.kyma-project.io"
)

readonly ADMIN_CLUSTER_SCOPE=(
 # Rafter resources
  "clusterassetgroup.rafter.kyma-project.io"
  "clusterasset.rafter.kyma-project.io"
  "clusterbucket.rafter.kyma-project.io"
)

trap cleanup EXIT
ERROR_LOGGING_GUARD="true"

CreateBindings

echo "---> Run tests for ${VIEW_EMAIL}"
LogInAsUser $VIEW_EMAIL $VIEW_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${VIEW_EMAIL}
testComponent $USER $NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $SYSTEM_NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponentClusterScoped $USER "yes" "no" "${NAMESPACE_ADMIN_CLUSTER_SCOPE[@]}" "${ADMIN_CLUSTER_SCOPE[@]}"

echo "---> Run tests for ${DEVELOPER_EMAIL}"
LogInAsUser $DEVELOPER_EMAIL $DEVELOPER_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${DEVELOPER_EMAIL}
testComponent $USER $NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $SYSTEM_NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponentClusterScoped $USER "yes" "no" "${NAMESPACE_ADMIN_CLUSTER_SCOPE[@]}" "${ADMIN_CLUSTER_SCOPE[@]}"

echo "---> Run tests for ${NAMESPACE_ADMIN_EMAIL}"
LogInAsUser $NAMESPACE_ADMIN_EMAIL $NAMESPACE_ADMIN_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${NAMESPACE_ADMIN_EMAIL}
testComponent $USER $NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}"
testComponent $USER $NAMESPACE "yes" "no" "${ADMIN_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "no" "${ADMIN_RESOURCES[@]}"
testComponent $USER $SYSTEM_NAMESPACE "yes" "no" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponentClusterScoped $USER "yes" "yes" "${NAMESPACE_ADMIN_CLUSTER_SCOPE[@]}"
testComponentClusterScoped $USER "yes" "no" "${ADMIN_CLUSTER_SCOPE[@]}"

echo "---> Run tests for ${ADMIN_EMAIL}"
LogInAsUser $ADMIN_EMAIL $ADMIN_PASSWORD
export KUBECONFIG=${PWD}/kubeconfig-${ADMIN_EMAIL}
testComponent $USER $NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $CUSTOM_NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponent $USER $SYSTEM_NAMESPACE "yes" "yes" "${COMMON_RESOURCES[@]}" "${NAMESPACE_ADMIN_RESOURCES[@]}" "${ADMIN_RESOURCES[@]}"
testComponentClusterScoped $USER "yes" "yes" "${NAMESPACE_ADMIN_CLUSTER_SCOPE[@]}" "${ADMIN_CLUSTER_SCOPE[@]}"

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"