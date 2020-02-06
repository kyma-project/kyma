#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma on Minikube, and runs the integrations tests. 
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../../

sudo ${INSTALLATION_DIR}/cmd/run.sh --vm-driver "none"

echo "Fix Minikube issue - https://github.com/kubernetes/minikube/issues/6523"
sudo docker pull gcr.io/k8s-minikube/storage-provisioner@sha256:088daa9fcbccf04c3f415d77d5a6360d2803922190b675cb7fc88a9d2d91985a
sudo docker tag gcr.io/k8s-minikube/storage-provisioner@sha256:088daa9fcbccf04c3f415d77d5a6360d2803922190b675cb7fc88a9d2d91985a gcr.io/k8s-minikube/storage-provisioner:v1.8.1
sudo kubectl -n kube-system delete pod storage-provisioner

sudo ${INSTALLATION_DIR}/scripts/is-installed.sh --timeout 30m
sudo ${INSTALLATION_DIR}/scripts/watch-pods.sh
sudo ${INSTALLATION_DIR}/scripts/testing.sh
