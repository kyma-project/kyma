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

cat <<EOF | kubectl apply -f -
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

# createClusterBuckets create rafter.ClusterBuckets from assetstore.ClusterBuckets
#
createClusterBuckets() {
  local -r clusterBuckets=$(getResources clusterbuckets.assetstore.kyma-project.io)

  if [[ -z ${clusterBuckets} ]]; then
    echo "There are not any ClusterBuckets :("
  else
    IFS=$'\n'
    for clusterBucket in ${clusterBuckets}
    do
      createClusterBucket "${clusterBucket}"
    done
  fi
}

# createClusterBucket create rafter.ClusterBucket by given from assetstore.ClusterBucket
#
# Arguments:
#   $1 - ClusterBucket
createClusterBucket() {
  local -r clusterBucket="${1}"

  local -r name="$(echo ${clusterBucket} | jq -r '.metadata.name')"
  local -r labels="$(echo ${clusterBucket} | jq -r '.metadata.labels')"
  local -r region="$(echo ${clusterBucket} | jq -r '.spec.region')"
  local -r policy="$(echo ${clusterBucket} | jq -r '.spec.policy')"

cat <<EOF | kubectl apply -f -
apiVersion: rafter.kyma-project.io/v1beta1
kind: ClusterBucket
metadata:
  name: ${name}
  labels: ${labels}
spec:
  region: ${region}
  policy: ${policy}
EOF
}

# createBuckets create rafter.Buckets from assetstore.Buckets
#
# Arguments:
#   $1 - Namespace name
createBuckets() {
  local -r namespace="${1}"
  local -r buckets=$(getResources buckets.assetstore.kyma-project.io ${namespace})

  if [[ -z ${buckets} ]]; then
    echo "There are not any Buckets :("
  else
    IFS=$'\n'
    for bucket in ${buckets}
    do
      createBucket "${bucket}" "${namespace}"
    done
  fi
}

# createBucket create rafter.Bucket by given from assetstore.Bucket
#
# Arguments:
#   $1 - Bucket
#   $2 - Namespace name
createBucket() {
  local -r bucket="${1}"
  local -r namespace="${2}"

  local -r name="$(echo ${bucket} | jq -r '.metadata.name')"
  local -r labels="$(echo ${bucket} | jq -r '.metadata.labels')"
  local -r region="$(echo ${bucket} | jq -r '.spec.region')"
  local -r policy="$(echo ${bucket} | jq -r '.spec.policy')"

cat <<EOF | kubectl apply -f -
apiVersion: rafter.kyma-project.io/v1beta1
kind: Bucket
metadata:
  name: ${name}
  namespace: ${namespace}
  labels: ${labels}
spec:
  region: ${region}
  policy: ${policy}
EOF
}

main() {
  local -r namespaces="$(kubectl get namespaces -o=jsonpath='{.items[*].metadata.name}')"

  if [[ -z ${namespaces} ]]; then
    echo "There are not any Namespaces :("
  else
    createClusterAssetGroups
    createClusterBucket

    for namespace in ${namespaces}
    do
      createAssetGroups "${namespace}"
      createBucket "${namespace}"
    done
  fi
}
main
