#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma on Minikube, and runs the integrations tests. 
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../../
SECONDS=0
sudo ${INSTALLATION_DIR}/cmd/run.sh --vm-driver "none"
sudo ${INSTALLATION_DIR}/scripts/is-installed.sh --timeout 30m
duration=$SECONDS
echo "$(($duration / 60)) minutes and $(($duration % 60)) seconds elapsed."
sudo ${INSTALLATION_DIR}/scripts/watch-pods.sh
sudo ${INSTALLATION_DIR}/scripts/testing.sh
