#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

main() {
  local -r isAssetStoreInstalled="$(kubectl get cm -n kube-system -l NAME=assetstore,OWNER=TILLER)"
  if [ -z "${isAssetStoreInstalled}" ]; then
    exit 0
  fi

  local -r public_bucket="$(kubectl get cm asset-upload-service -n kyma-system -o jsonpath="{.data['public']}")"
  local -r private_bucket="$(kubectl get cm asset-upload-service -n kyma-system -o jsonpath="{.data['private']}")"

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${CONFIGMAP_NAME}
  namespace: ${CONFIGMAP_NAMESPACE}
data:
  public: ${public_bucket}
  private: ${private_bucket}
EOF
}
main
