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

# serializeSources serialize single source from ClusterDocsTopic to ClusterAssetGroup
#
# Arguments:
#   $1 - Sources
serializeSource() {
  local -r source="${1}"
  local -r mode="$(echo ${source} | jq -r '.mode')"
  local -r name="$(echo ${source} | jq -r '.name')"
  local -r type="$(echo ${source} | jq -r '.type')"
  local -r url="$(echo ${source} | jq -r '.url')"
  local -r filter="$(echo ${source} | jq -r '.filter')"

echo "
  - mode: ${mode}
    name: ${name}
    type: ${type}
    url: ${url}
    filter: ${filter}
"
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

  for kymaDocsCDT in "api-gateway" "api-gateway-v2" "application-connector" "asset-store" "backup" "compass" "console" "event-bus" "headless-cms" "helm-broker" "logging" "monitoring" "security" "serverless" "service-catalog" "service-mesh" "tracing" "kyma"
  do
    if [ "${clusterDocsTopic}" == "${kymaDocsCDT}" ]; then
      return 0
    else
      continue
    fi    
  done

  local -r labels="$(echo ${clusterDocsTopic} | jq -r '.metadata.labels')"
  local -r description="$(echo ${clusterDocsTopic} | jq -r '.spec.description')"
  local -r displayName="$(echo ${clusterDocsTopic} | jq -r '.spec.displayName')"
  local -r sources="$(echo ${clusterDocsTopic} | jq -c '.spec.sources[]')"
  local serialized_sources=""

  IFS=$'\n'
  for source in ${sources}
  do
serialized_sources="
${serialized_sources}
$(serializeSource ${source})"
  done

cat <<EOF | kubectl apply -f -
apiVersion: rafter.kyma-project.io/v1beta1
kind: ClusterAssetGroup
metadata:
  name: ${name}
  labels: ${labels}
spec:
  description: ${description}
  displayName: ${displayName}
  sources: 
  ${serialized_sources}
EOF
}

# createAssetGroups create AssetGroups by given DocsTopic
#
# Arguments:
#   $1 - Namespace name
createAssetGroups() {
  local -r namespace="${1}"
  local -r docsTopics=$(getResources docstopics.cms.kyma-project.io ${namespace})

  if [[ -z ${docsTopics} ]]; then
    echo "In ${namespace} namespace are not any DocsTopics :("
  else
    IFS=$'\n'
    for docsTopic in ${docsTopics}
    do
      createAssetGroup "${docsTopic}" "${namespace}"
    done
  fi
}

# createAssetGroup create AssetGroup from DocsTopic
#
# Arguments:
#   $1 - DocsTopic
#   $2 - Namespace name
createAssetGroup() {
  local -r docsTopic="${1}"
  local -r namespace="${2}"

  local -r name="$(echo ${docsTopic} | jq -r '.metadata.name')"
  local -r labels="$(echo ${docsTopic} | jq -r '.metadata.labels')"
  local -r description="$(echo ${docsTopic} | jq -r '.spec.description')"
  local -r displayName="$(echo ${docsTopic} | jq -r '.spec.displayName')"
  local -r sources="$(echo ${docsTopic} | jq -c '.spec.sources[]')"
  local serialized_sources=""

  IFS=$'\n'
  for source in ${sources}
  do
serialized_sources="
${serialized_sources}
$(serializeSource ${source})"
  done

cat <<EOF
apiVersion: rafter.kyma-project.io/v1beta1
kind: AssetGroup
metadata:
  name: ${name}
  namespace: ${namespace}
  labels: ${labels}
spec:
  description: ${description}
  displayName: ${displayName}
  sources: 
  ${serialized_sources}
EOF
}

main() {
  local -r namespaces="$(kubectl get namespaces -o=jsonpath='{.items[*].metadata.name}')"

  if [[ -z ${namespaces} ]]; then
    echo "There are not any Namespaces :("
  else
    createClusterAssetGroups

    for namespace in ${namespaces}
    do
      createAssetGroups "${namespace}"
    done
  fi
}
main
