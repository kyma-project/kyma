#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$CURRENT_DIR/../../

DOCKERFILE="${ROOT_DIR}/ci.Dockerfile"

FINAL_IMAGE="kyma-on-minikube"
KUBECTL_CLI_VERSION=1.10.0
KUBELESS_CLI_VERSION=1.0.0-alpha.7
MINIKUBE_VERSION=0.28.2
HELM_VERSION=2.8.2

pushd $ROOT_DIR

docker build -t ${FINAL_IMAGE} \
    -f ${DOCKERFILE} \
    --build-arg KUBECTL_CLI_VERSION=${KUBECTL_CLI_VERSION} \
    --build-arg KUBELESS_CLI_VERSION=${KUBELESS_CLI_VERSION} \
    --build-arg MINIKUBE_VERSION=${MINIKUBE_VERSION} \
    --build-arg HELM_VERSION=${HELM_VERSION} .

popd