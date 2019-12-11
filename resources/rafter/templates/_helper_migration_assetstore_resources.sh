#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly CLUSTER_DOCS_TOPICS_NAMES="kyma api-gateway api-gateway-v2 application-connector asset-store backup compass console event-bus headless-cms helm-broker logging monitoring security serverless service-catalog service-mesh tracing"
readonly BLACK_LIST_LABELS="cms.kyma-project.io/access cms.kyma-project.io/docs-topic"
readonly OLD_MINIO_ENDPOINT="https://minio.kyma.local"
readonly NEW_MINIO_ENDPOINT="http://rafter-minio.kyma-system.svc.cluster.local:9000"
readonly DONT_CREATE_RESOURCE="dont create"

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
    return 0
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
    return 0
  fi

  local serialized_labels=""
  local -r keys="$(echo ${labels} | jq -r '. | keys[]')"

  IFS=$' '
  for blackLabel in $BLACK_LIST_LABELS
  do
    IFS=$'\n'
    for key in $keys
    do
      if [ "${blackLabel}" == "${key}" ]; then
        echo "${DONT_CREATE_RESOURCE}"
        return 0
      fi
    done
  done

  labels=$(echo ${labels} | jq -r 'to_entries|map("\(.key) = \(.value|tostring)" )| .[]')
  IFS=$'\n'
  for label in $labels
  do
    IFS=$" = " read key value <<< ${label}

    

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
    local serialized=$(serializeAssetGroupSource ${source})
serialized_sources="
${serialized_sources}
${serialized}
"
  done

  echo "
  sources:
  ${serialized_sources}
  "
}

# serializeAssetGroupSource serialize single source for (CLuster)AssetGroup
#
# Arguments:
#   $1 - Source
serializeAssetGroupSource() {
  local -r source="${1}"

  local -r mode="$(echo ${source} | jq -r '.mode')"
  local -r name="$(echo ${source} | jq -r '.name')"
  local -r type="$(echo ${source} | jq -r '.type')"
  local url="$(echo ${source} | jq -r '.url')"
  url=$(echo "${url/$OLD_MINIO_ENDPOINT/$NEW_MINIO_ENDPOINT}")

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

# serializeAssetSource serialize single source for (CLuster)Asset
#
# Arguments:
#   $1 - Source
serializeAssetSource() {
  local -r source="${1}"

  local -r mode="$(echo ${source} | jq -r '.mode')"
  local url="$(echo ${source} | jq -r '.url')"
  url=$(echo "${url/$OLD_MINIO_ENDPOINT/$NEW_MINIO_ENDPOINT}")

  local filter="$(echo ${source} | jq -r '.filter')"
  if [ "${filter}" == "null" ]; then
    filter=""
  else
    filter="filter: ${filter}"
  fi

  local metadataWebhookService="$(serializeWebhooks ${source} metadataWebhookService)"
  if [ "${metadataWebhookService}" == "null" ]; then
    metadataWebhookService=""
  fi

  local mutationWebhookService="$(serializeWebhooks ${source} mutationWebhookService)"
  if [ "${mutationWebhookService}" == "null" ]; then
    mutationWebhookService=""
  fi

  local validationWebhookService="$(serializeWebhooks ${source} validationWebhookService)"
  if [ "${validationWebhookService}" == "null" ]; then
    validationWebhookService=""
  fi

  echo "
    mode: ${mode}
    url: ${url}
    ${filter}
    ${metadataWebhookService}
    ${mutationWebhookService}
    ${validationWebhookService}
  "
}

# serializeWebhooks serialize webhooks from fingle source
#
# Arguments:
#   $1 - Source
#   $2 - Webhooks name
serializeWebhooks() {
  local -r source="${1}"
  local -r webhooks_name="${2}"

  local webhooks="$(echo ${source} | jq -c ".${webhooks_name}")"
  if [ "${webhooks}" == "null" ]; then
    echo "null"
    return 0
  fi
  webhooks="$(echo ${source} | jq -c ".${webhooks_name}[]")"

  local serialized_webhooks=""
  IFS=$'\n'
  for webhook in ${webhooks}
  do
    local serialized=$(serializeWebhook ${webhook})
serialized_webhooks="
${serialized_webhooks}
${serialized}
"
  done

  echo "
    ${webhooks_name}:
    ${serialized_webhooks}
  "
}

# serializeWebhook serialize single webhook
#
# Arguments:
#   $1 - Webhook
serializeWebhook() {
  local -r webhook="${1}"

  local -r name="$(echo ${webhook} | jq -r '.name')"
  local -r namespace="$(echo ${webhook} | jq -r '.namespace')"

  local filter="$(echo ${webhook} | jq -r '.filter')"
  if [ "${filter}" == "null" ]; then
    filter=""
  else
    filter="filter: ${filter}"
  fi

  local endpoint="$(echo ${webhook} | jq -r '.endpoint')"
  if [ "${endpoint}" == "null" ]; then
    endpoint=""
  else
    endpoint="endpoint: ${endpoint}"
  fi

  echo "
    - name: ${name}
      namespace: ${namespace}
      ${filter}
      ${endpoint}
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
    return 0
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
#   $1 - (Cluster)DocsTopic
#   $2 - Kind. ClusterAssetGroup or AssetGroup
createAssetGroup() {
  local -r docsTopic="${1}"
  local -r kind="${2}"

  local -r labels="$(serializeLabels ${docsTopic})"
  if [ "${labels}" == "${DONT_CREATE_RESOURCE}" ]; then
    return 0
  fi

  local -r name="$(echo ${docsTopic} | jq -r '.metadata.name')"
  if [ "${kind}" == "ClusterAssetGroup" ]; then
    IFS=$' '
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
  local -r description="$(echo ${docsTopic} | jq -r '.spec.description')"
  local -r displayName="$(echo ${docsTopic} | jq -r '.spec.displayName')"
  local -r sources="$(serializeSources ${docsTopic})"

# cat <<EOF | kubectl apply -f -
cat <<EOF
apiVersion: rafter.kyma-project.io/v1beta1
kind: ${kind}
metadata:
  name: ${name}
  ${namespace}
  ${labels}
spec:
  description: ${description}
  displayName: ${displayName}
  ${sources} 
EOF
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
    return 0
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

  local -r labels="$(serializeLabels ${bucket})"
  if [ "${labels}" == "${DONT_CREATE_RESOURCE}" ]; then
    return 0
  fi

  local -r name="$(echo ${bucket} | jq -r '.metadata.name')"
  local -r namespace="$(serializeNamespace ${bucket})"
  local -r region="$(echo ${bucket} | jq -r '.spec.region')"
  local -r policy="$(echo ${bucket} | jq -r '.spec.policy')"

# cat <<EOF | kubectl apply -f -
cat <<EOF
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

# createAssets create rafter.(Cluster)Assets from assetstore.(Cluster)Assets
#
# Arguments:
#   $1 - Namespaced or not
createAssets() {
  local assets=""
  local kind=""

  if [ -n "${1-}" ] ; then
    assets=$(getResources assets.assetstore.kyma-project.io)
    kind="Asset"
  else
    assets=$(getResources clusterassets.assetstore.kyma-project.io)
    kind="ClusterAsset"
  fi

  if [[ -z ${assets} ]]; then
    echo "There are not any ${kind}s :("
    return 0
  fi

  IFS=$'\n'
  for asset in ${assets}
  do
    createAsset "${asset}" "${kind}"
  done
}

# createAsset create rafter.(Cluster)Asset by given assetstore.(Cluster)Asset
#
# Arguments:
#   $1 - (Cluster)Assets
#   $2 - Kind. ClusterAsset or Asset
createAsset() {
  local -r asset="${1}"
  local -r kind="${2}"

  local -r labels="$(serializeLabels ${asset})"
  if [ "${labels}" == "${DONT_CREATE_RESOURCE}" ]; then
    return 0
  fi

  local -r name="$(echo ${asset} | jq -r '.metadata.name')"
  local -r namespace="$(serializeNamespace ${asset})"
  local -r bucketRef="$(echo ${asset} | jq -r '.spec.bucketRef.name')"
  local source="$(echo ${asset} | jq -c '.spec.source')"
  source=$(serializeAssetSource ${source})

# cat <<EOF | kubectl apply -f -
cat <<EOF
apiVersion: rafter.kyma-project.io/v1beta1
kind: ${kind}
metadata:
  name: ${name}
  ${namespace}
  ${labels}
spec:
  bucketRef:
    name: ${bucketRef}
  source:
    ${source}
EOF
}

main() {
  createAssetGroups
  createAssetGroups "namespaced"
  createBuckets
  createBuckets "namespaced"
  createAssets
  createAssets "namespaced"
}
main
