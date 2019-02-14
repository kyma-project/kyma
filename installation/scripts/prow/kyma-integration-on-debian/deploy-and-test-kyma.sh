#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma on Minikube, and runs the integrations tests. 
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../../

sudo ${CURRENT_DIR}/install-deps-debian.sh
sudo ${INSTALLATION_DIR}/cmd/run.sh --vm-driver "none"
sudo ${INSTALLATION_DIR}/scripts/is-installed.sh --timeout 30m
sudo ${INSTALLATION_DIR}/scripts/watch-pods.sh
sudo ${INSTALLATION_DIR}/scripts/testing.sh