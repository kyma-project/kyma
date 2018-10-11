#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$CURRENT_DIR/.."
KYMA_PATH="${ROOT_DIR}/../.."
IMAGE_NAME=eu.gcr.io/kyma-project/develop/installer:dev

echo "
################################################################################
# Installer build
################################################################################
"

pushd ${ROOT_DIR}

echo "Running update-codegen"
${ROOT_DIR}/hack/update-codegen.sh

echo "Running go build"
export GOOS=linux && go build -o installer ${ROOT_DIR}/cmd/operator/main.go

echo "Building docker image"
eval $(minikube docker-env --shell bash)
docker build -t ${IMAGE_NAME} -f deploy/installer/Dockerfile .

popd
