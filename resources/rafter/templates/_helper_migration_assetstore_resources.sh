#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly CLUSTER_DOCS_TOPICS_NAMES="kyma api-gateway api-gateway-v2 application-connector asset-store backup compass console event-bus headless-cms helm-broker logging monitoring security serverless service-catalog service-mesh tracing"
readonly BLACK_LIST_BUCKET_LABELS=""
readonly BLACK_LIST_ASSET_LABELS=""

# getResource get specification of k8s resources by given type name
#
# Arguments:
#   $1 - Resources type name
getResources() {
  local -r resource_type="${1}"
  echo "$(kubectl get ${resource_type} --all-namespaces -ojson | jq -c '.items[]')"
}

# serializeNamespace serialize namespace from given resource
#
# Arguments:
#   $1 - Resource
serializeNamespace() {
  local -r resource="${1}"
  local namespace="$(echo ${resource} | jq -r '.metadata.namespace')"
  if [ "${namespace}" == "null" ]; then
    echo ""
  fi
  echo "namespace: ${namespace}"
}

# serializeLabels serialize labels from given resource
#
# Arguments:
#   $1 - Resource
serializeLabels() {
  local -r resource="${1}"
  local labels="$(echo ${resource} | jq -r '.metadata.labels')"
  if [ "${labels}" == "null" ]; then
    echo ""
  fi

  local serialized_labels=""
  echo "${labels}" | jq -r '. | keys[]' | while read key ; do
    local value="$(echo ${labels} | jq ".${key}")"
    serialized_labels="
    ${serialized_labels}
    ${key}: ${value}
    "
  done

  echo "
  labels:
    ${serialized_labels}
  "
}

# serializeSources serialize sources from (Cluster)DocsTopic
#
# Arguments:
#   $1 - (Cluster)DocsTopic
serializeSources() {
  local -r docsTopic="${1}"
  local -r sources="$(echo ${docsTopic} | jq -c '.spec.sources[]')"
  local serialized_sources=""

  IFS=$'\n'
  for source in ${sources}
  do
    local serialized=$(serializeSource ${source})
serialized_sources="
${serialized_sources}
${serialized}
"
  done

  echo "${serialized_sources}"
}

# serializeSource serialize single source
#
# Arguments:
#   $1 - Source
serializeSource() {
  local -r source="${1}"

  local -r mode="$(echo ${source} | jq -r '.mode')"
  local -r name="$(echo ${source} | jq -r '.name')"
  local -r type="$(echo ${source} | jq -r '.type')"
  local -r url="$(echo ${source} | jq -r '.url')"
  local filter="$(echo ${source} | jq -r '.filter')"

  if [ "${filter}" == "null" ]; then
    filter=""
  else
    filter="filter: ${filter}"
  fi

  echo "
  - mode: ${mode}
    name: ${name}
    type: ${type}
    url: ${url}
    ${filter}
  "
}

# createAssetGroups create rafter.(Cluster)AssetGroups from cms.(Cluster)DocsTopics
#
# Arguments:
#   $1 - Namespaced or not
createAssetGroups() {
  local docsTopics=""
  local kind=""

  if [ -n "${1-}" ] ; then
    docsTopics=$(getResources docstopics.cms.kyma-project.io)
    kind="AssetGroup"
  else
    docsTopics=$(getResources clusterdocstopics.cms.kyma-project.io)
    kind="ClusterAssetGroup"
  fi

  if [[ -z ${docsTopics} ]]; then
    echo "There are not any ${kind}s :("
  fi

  IFS=$'\n'
  for docsTopic in ${docsTopics}
  do
    createAssetGroup "${docsTopic}" "${kind}"
  done
}

# createAssetGroup create rafter.(Cluster)AssetGroup by given cms.(Cluster)DocsTopic
#
# Arguments:
#   $1 - (Cluster)AssetGroup
#   $2 - Kind. ClusterAssetGroup or AssetGroup
createAssetGroup() {
  local -r docsTopic="${1}"
  local -r kind="${2}"

  local -r name="$(echo ${docsTopic} | jq -r '.metadata.name')"
  if [ "${kind}" == "ClusterAssetGroup" ]; then
    for docsCDT in $CLUSTER_DOCS_TOPICS_NAMES
    do
      if [ "${name}" == "${docsCDT}" ]; then
        return 0
      else
        continue
      fi    
    done
  fi

  local -r namespace="$(serializeNamespace ${docsTopic})"
  local -r labels="$(serializeLabels ${docsTopic})"
  echo "${labels}"
#   local -r description="$(echo ${docsTopic} | jq -r '.spec.description')"
#   local -r displayName="$(echo ${docsTopic} | jq -r '.spec.displayName')"
#   local -r sources="$(serializeSources ${docsTopic})"

# cat <<EOF | kubectl apply -f -
# apiVersion: rafter.kyma-project.io/v1beta1
# kind: ${kind}
# metadata:
#   name: ${name}
#   ${namespace}
#   ${labels}
# spec:
#   description: ${description}
#   displayName: ${displayName}
#   sources:
#   ${sources} 
# EOF
}

# createBuckets create rafter.(Cluster)Buckets from assetstore.(Cluster)Buckets
#
# Arguments:
#   $1 - Namespaced or not
createBuckets() {
  local buckets=""
  local kind=""

  if [ -n "${1-}" ] ; then
    buckets=$(getResources buckets.assetstore.kyma-project.io)
    kind="Bucket"
  else
    buckets=$(getResources clusterbuckets.assetstore.kyma-project.io)
    kind="ClusterBucket"
  fi

  if [[ -z ${buckets} ]]; then
    echo "There are not any ${kind}s :("
  fi

  IFS=$'\n'
  for bucket in ${buckets}
  do
    createBucket "${bucket}" "${kind}"
  done
}

# createBucket create rafter.(Cluster)Bucket by given assetstore.(Cluster)Bucket
#
# Arguments:
#   $1 - (Cluster)Bucket
#   $2 - Kind. ClusterBucket or Bucket
createBucket() {
  local -r bucket="${1}"
  local -r kind="${2}"

  local -r name="$(echo ${bucket} | jq -r '.metadata.name')"
  local -r namespace="$(serializeNamespace ${bucket})"
  local -r labels="$(serializeLabels ${bucket})"  
  local -r region="$(echo ${bucket} | jq -r '.spec.region')"
  local -r policy="$(echo ${bucket} | jq -r '.spec.policy')"

cat <<EOF | kubectl apply -f -
apiVersion: rafter.kyma-project.io/v1beta1
kind: ${kind}
metadata:
  name: ${name}
  ${namespace}
  ${labels}
spec:
  region: ${region}
  policy: ${policy}
EOF
}

main() {
  createAssetGroups
  createAssetGroups "namespaced"
  # createBuckets
  # createBuckets "namespaced"
  # createAssets
  # createAssets "namespaced"
}
main
