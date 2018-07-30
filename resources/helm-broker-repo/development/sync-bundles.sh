#!/usr/bin/env bash

NAMESPACE=kyma-system

BASE_DIR="$(dirname $0)/.."

PROVISIONING_DIR="${BASE_DIR}/provisioning"

bash ${PROVISIONING_DIR}/provision-bundles.sh ${BASE_DIR}/bundles

HELM_BROKER_POD_NAME=$(kubectl get po -n ${NAMESPACE} -l app=core-helm-broker -o jsonpath="{.items[0].metadata.name}")

kubectl delete pod ${HELM_BROKER_POD_NAME} -n ${NAMESPACE}
