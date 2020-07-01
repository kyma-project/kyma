#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma on Minikube, and runs the integrations tests. 
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../../
START_TIME=$SECONDS
sudo ${INSTALLATION_DIR}/cmd/run.sh --vm-driver "none"
sudo ${INSTALLATION_DIR}/scripts/is-installed.sh --timeout 30m
ELAPSED_TIME=$(($SECONDS - $START_TIME))
echo "$(($ELAPSED_TIME / 60)) minutes and $(($ELAPSED_TIME % 60)) seconds elapsed."
sudo ${INSTALLATION_DIR}/scripts/watch-pods.sh
sudo ${INSTALLATION_DIR}/scripts/testing.sh
