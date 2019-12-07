#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

main() {
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
