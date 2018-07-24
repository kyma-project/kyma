#!/usr/bin/env bash

NAMESPACE=kyma-system

PROVISIONING_DIR="$(dirname $0)/../provisioning"

bash ${PROVISIONING_DIR}/provision-bundles.sh bundles

HELM_BROKER_POD_NAME=$(kubectl get po -n ${NAMESPACE} -l app=core-helm-broker -o jsonpath="{.items[0].metadata.name}")

kubectl delete pod ${HELM_BROKER_POD_NAME} -n ${NAMESPACE}