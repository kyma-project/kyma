#!/usr/bin/env bash

echo "The script install-tiller.sh is deprecated and will be removed with Kyma release 1.14, please use Kyma CLI instead"

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/../resources/tiller.yaml
bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller
