#!/usr/bin/env bash

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Enabling SSL for Tiller"
bash "${CURRENT_DIR}/tiller-ssl.sh"

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/../resources/tiller.yaml
bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller
