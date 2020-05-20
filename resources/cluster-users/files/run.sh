#!/bin/bash -e
NAMESPACE_ADMIN_EMAIL=
VIEW_EMAIL=
DEVELOPER_EMAIL=
ADMIN_EMAIL=admin@kyma.cx
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source "${DIR}/library.sh"

# resources/cluster-users/values.yaml clusterRoles.verbs.view
readonly VIEW_OPERATIONS=( "get" "list" )
# resources/cluster-users/values.yaml clusterRoles.verbs.edit - clusterRoles.verbs.view
readonly EDIT_OPERATIONS=( "create" "delete" "deletecollection" "patch" "update" "watch" )

# users used in tests
readonly USERS=( "${NAMESPACE_ADMIN_EMAIL}" "${VIEW_EMAIL}" "${DEVELOPER_EMAIL}" "${ADMIN_EMAIL}" )

readonly K8S_RESOURCES=(
  "deployment"
  "pod"
  "secret"
  "confgimap"
  "serviceaccount"
  "role"
  "rolebinding"
)
readonly K8S_RESOURCES_CLUSTER_SCOPE=(
  "namespace"
  "nodes"
  "clusterrole"
  "clusterrolebinding"
  "crd"
)

readonly ORY_RESOURCES=(
  "oauth2clients.hydra.ory.sh"
  "rule.oathkeeper.ory.sh"
)

readonly KYMA_RESOURCES=(
  "installations.installer.kyma-project.io"
  "apirules.gateway.kyma-project.io"
  "applicationmappings.applicationconnector.kyma-project.io"
  "addonsconfigurations.addons.kyma-project.io"
  "servicebindingusages.servicecatalog.kyma-project.io"
  "assetgroup.rafter.kyma-project.io"
  "asset.rafter.kyma-project.io"
  "bucket.rafter.kyma-project.io"
)
readonly KYMA_RESOURCES_CLUSTER_SCOPE=(
  "applications.applicationconnector.kyma-project.io"
  "usagekinds.servicecatalog.kyma-project.io"
  "clusterassetgroup.rafter.kyma-project.io"
  "clusterasset.rafter.kyma-project.io"
  "clusterbucket.rafter.kyma-project.io"
)

readonly SERVICE_CATALOG_RESOURCES=(
  "serviceinstances.servicecatalog.k8s.io"
  "servicebindings.servicecatalog.k8s.io"
)

readonly KNATIVE_SERVING_KYMA_RESOURCES=(
  "services.serving.knative.dev"
  "routes.serving.knative.dev"
  "revisions.serving.knative.dev"
  "configurations.serving.knative.dev"
  "podautoscalers.autoscaling.internal.knative.dev"
  "images.caching.internal.knative.dev"
)

trap cleanup EXIT
ERROR_LOGGING_GUARD="true"

CreateBindings

for USER in "${USERS[@]}"; do
	# Get password for USER
	LogInAsUser $USER $ADMIN_PASSWORD
	export KUBECONFIG=${PWD}/kubeconfig-${USER}
	# Run all tests for USER
	testComponent $USER $NAMESPACE "yes" "yes" "${K8S_RESOURCES[@]}"
	testComponent $USER $NAMESPACE "yes" "yes" "${ORY_RESOURCES[@]}"
	testComponent $USER $NAMESPACE "yes" "yes" "${KYMA_RESOURCES[@]}"
done

# Create bindings
# for USER in [X Y Z]
	# Login as user X
	# Test as user X
# Cleanup

echo "ALL TESTS PASSED"
ERROR_LOGGING_GUARD="false"
