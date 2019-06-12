#!/usr/bin/env bash

echo "The install-tiller.sh script is deprecated and will be removed. Use Kyma CLI instead."

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/../resources/tiller.yaml
bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller
