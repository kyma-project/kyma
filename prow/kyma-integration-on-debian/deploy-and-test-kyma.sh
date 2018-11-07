#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma on Minikube, and runs the integrations tests. 
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${CURRENT_DIR}/../../

sudo ${CURRENT_DIR}/install-deps-debian.sh
sudo ${ROOT_DIR}/installation/cmd/run.sh --vm-driver "none"
sudo ${ROOT_DIR}/installation/scripts/is-installed.sh
sudo ${ROOT_DIR}/installation/scripts/testing.sh