#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

main() {
  local -r isAssetStoreInstalled="$(kubectl get cm -n kube-system -l NAME=assetstore,OWNER=TILLER)"
  if [ -z "${isAssetStoreInstalled}" ]; then
    exit 0
  fi

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${CONFIGMAP_NAME}
  namespace: ${CONFIGMAP_NAMESPACE}
data:
  private: ${ASSET_STORE_PRIVATE_BUCKET}
  public: ${ASSET_STORE_PUBLIC_BUCKET}
EOF
}
main
