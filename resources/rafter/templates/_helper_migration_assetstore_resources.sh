#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# getResource get specification of k8s resources by given type name
#
# Arguments:
#   $1 - Resources type name
#   $2 - Namespace name
getResources() {
  local -r resource_type="${1}"

  if [ -n "${2-}" ] ; then
    local -r resource_namespace="${2}"
    echo "$(kubectl get ${resource_type} -n ${resource_namespace} -ojson | jq -c '.items[]')"
  else
    echo "$(kubectl get ${resource_type} -ojson | jq -c '.items[]')"
  fi
}

# serializeSources serialize sources field from ClusterDocsTopic to ClusterAssetGroup
#
# Arguments:
#   $1 - Sources
serializeSources() {
  local -r sources="${1}"

  if [[ -z ${sources} ]]; then
    echo "There are not any Sources :("
  else
    IFS=$'\n'
    for source in ${sources}
    do
      
    done
  fi
}

# createClusterAssetGroups create ClusterAssetGroups from ClusterDocsTopics
#
createClusterAssetGroups() {
  local -r clusterDocsTopics=$(getResources clusterdocstopics.cms.kyma-project.io)

  if [[ -z ${clusterDocsTopics} ]]; then
    echo "There are not any ClusterDocsTopics :("
  else
    IFS=$'\n'
    for clusterDocsTopic in ${clusterDocsTopics}
    do
      createClusterAssetGroup "${clusterDocsTopic}"
    done
  fi
}

# createClusterAssetGroup create ClusterAssetGroup by given ClusterDocsTopic
#
# Arguments:
#   $1 - ClusterDocsTopic
createClusterAssetGroup() {
  local -r clusterDocsTopic="${1}"

  local -r name="$(echo ${clusterDocsTopic} | jq -r '.metadata.name')"
  local -r labels="$(echo ${clusterDocsTopic} | jq -r '.metadata.labels')"
  local -r description="$(echo ${clusterDocsTopic} | jq -r '.spec.description')"
  local -r displayName="$(echo ${clusterDocsTopic} | jq -r '.spec.displayName')"
  local -r sources="$(echo ${clusterDocsTopic} | jq -c '.spec.sources[]')"
  echo "${sources}"
  local -r serialized_sources=$(serializeSources ${sources})

# cat <<EOF | kubectl apply -f -
# apiVersion: rafter.kyma-project.io/v1beta1
# kind: ClusterAssetGroup
# metadata:
#   name: ${name}
#   labels: ${labels}
# spec:
#   description: ${description}
#   displayName: ${displayName}
#   sources: ${sources}
# EOF
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
  local -r namespaces="$(kubectl get namespaces -o=jsonpath='{.items[*].metadata.name}')"

  if [[ -z ${namespaces} ]]; then
    echo "There are not any Namespaces :("
  else
    createClusterAssetGroups

    # for namespace in "${namespaces}"
    # do
    #   createAssetGroups "${namespace}"
    # done
  fi
}
main
