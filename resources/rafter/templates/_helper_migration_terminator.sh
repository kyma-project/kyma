#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# removeResource remove k8s resource with given type name, resource name and namespace (last is optional)
#
# Arguments:
#   $1 - Resources type name
#   $2 - Resources name
#   $3 - Resources namespace
removeResource() {
  local -r resource_type="${1}"
  local -r resource_name="${2}"
  local -r timeout=15s

  if [ -n "${3-}" ] ; then
    local -r resource_namespace="${3}"

    kubectl delete "${resource_type}" "${resource_name}" -n "${resource_namespace}"
    kubectl wait --for=delete "${resource_type}/${resource_name}" -n "${resource_namespace}" --timeout="${timeout}" || true
  else
    kubectl delete "${resource_type}" "${resource_name}"
    kubectl wait --for=delete "${resource_type}/${resource_name}" --timeout="${timeout}" || true
  fi
}

# removeResources remove all k8s resources with given type name (namespaced and cluster wide)
#
# Arguments:
#   $1 - Resources type name
removeResources() {
  local -r resource_type="${1}"
  local -r timeout=120s

  echo "${resource_type}"

  kubectl delete "${resource_type}" --all --all-namespaces
  kubectl wait --for=delete "${resource_type}" --all --all-namespaces --timeout="${timeout}" || true
}

# removeHelmRelease remove Helm release with given name
#
# Arguments:
#   $1 - Release name
removeHelmRelease() {
  local -r release_name="${1}"
  local -r timeout=300 # 300 seconds

  helm delete "${release_name}" --purge --wait --timeout "${timeout}"
}

removeHeadlessCMS() {
  removeResources "docstopics.cms.kyma-project.io"
  removeResources "clusterdocstopics.cms.kyma-project.io"

  removeHelmRelease "cms"

  removeResource "crd" "docstopics.cms.kyma-project.io"
  removeResource "crd" "clusterdocstopics.cms.kyma-project.io"
}

removeAssetStore() {
  removeResources "assets.assetstore.kyma-project.io"
  removeResources "buckets.assetstore.kyma-project.io"
  removeResources "clusterassets.assetstore.kyma-project.io"
  removeResources "clusterbuckets.assetstore.kyma-project.io"

  removeHelmRelease "assetstore"

  removeResource "crd" "assets.assetstore.kyma-project.io"
  removeResource "crd" "buckets.assetstore.kyma-project.io"
  removeResource "crd" "clusterassets.assetstore.kyma-project.io"
  removeResource "crd" "clusterbuckets.assetstore.kyma-project.io"

  # remove custom ConfigMap created by assetstore-upload-service, which is not related with assetstore release
  removeResource "cm" "asset-upload-service" "kyma-system"
}

main() {
  removeResource "pvc" "${PVC_NAME}"

  removeHeadlessCMS
  removeAssetStore
}
main
