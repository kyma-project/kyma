#!/usr/bin/env bash

NAMESPACE=kyma-system

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly PROVISIONING_DIR="${CURRENT_DIR}/../provisioning"
readonly BUNDLES_DIR="${CURRENT_DIR}/../bundles"

bash ${PROVISIONING_DIR}/provision-bundles.sh ${BUNDLES_DIR}

HELM_BROKER_POD_NAME=$(kubectl get po -n ${NAMESPACE} -l app=core-helm-broker -o jsonpath="{.items[0].metadata.name}")

kubectl delete pod ${HELM_BROKER_POD_NAME} -n ${NAMESPACE}
