#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# getResources get all resource by given type
#
# Arguments:
#   $1 - Resources type name
#   $2 - Namespace name
getResources() {
  local -r resource_type="${1}"

  if [ -n "${2-}" ] ; then
    local -r resource_namespace="${2}"
    echo "$(kubectl get ${resource_type} -n ${resource_namespace})"
  else
    echo "$(kubectl get ${resource_type})"
  fi
}

# createClusterAssetGroups create ClusterAssetGroups from ClusterDocsTopics
#
createClusterAssetGroups() {
  local -r clusterDocsTopics=($(getResources clusterdocstopics.cms.kyma-project.io))

  if [ ${#clusterDocsTopics[@]} -eq 0 ]; then
    echo "There are not any ClusterDocsTopics :("
  else
    for clusterDocsTopic in "${clusterDocsTopics[@]}"
    do
      createClusterAssetGroup ""
    done
  fi
}

# createClusterAssetGroup create ClusterAssetGroup
#
# Arguments:
#   $1 - Namespace name
createClusterAssetGroup() {
  local -r namespace="${1}"
}

# createAssetGroups create AssetGroups from DocsTopics
#
# Arguments:
#   $1 - Namespace name
createAssetGroups() {
  local -r namespace="${1}"
}

# createAssetGroup create AssetGroup
#
# Arguments:
#   $1 - Namespace name
createAssetGroup() {
  local -r namespace="${1}"
}


main() {
  local -r namespaces=($(kubectl get namespaces -o=jsonpath='{.items[*].metadata.name}')))

  if [ ${#namespaces[@]} -eq 0 ]; then
    echo "There are not any Namespaces :("
  else
    createClusterAssetGroups


    for namespace in "${namespaces[@]}"
    do
      createAssetGroups "${namespace}"
    done
  fi
}
main
