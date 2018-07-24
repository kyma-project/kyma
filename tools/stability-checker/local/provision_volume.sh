#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "$0")"; pwd)"

kubectl create -f ${CURRENT_DIR}/provisioning.yaml

bash ${CURRENT_DIR}/helpers/isready.sh kyma-system app  stability-test-provisioner

kubectl cp ${CURRENT_DIR}/input stability-test-provisioner:/home/input -n kyma-system

kubectl delete pod -n kyma-system stability-test-provisioner