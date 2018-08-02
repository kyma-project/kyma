#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$CURRENT_DIR/.."
KYMA_PATH="${ROOT_DIR}/../.."

echo "
################################################################################
# Kyma Installer build
################################################################################
"

#sed regex engine is greedy - it will try to find longest possible match.
#it means that regardless of how many colons are there, it will match from beginning till the _last_ colon.
IMAGE_VERSION=$(cat ${KYMA_PATH}/installation/resources/installer.yaml  | grep "image:" | sed 's/^..*[:]//')
if [[ "${IMAGE_VERSION}" == "" ]]; then
    echo "Can't find image version!"
    exit 1
else
    echo "Building installer version: ${IMAGE_VERSION}"
fi

IMAGE_NAME=eu.gcr.io/kyma-project/installer:${IMAGE_VERSION}

pushd ${ROOT_DIR}

echo "Running update-codegen"
${ROOT_DIR}/hack/update-codegen.sh

echo "Running go build"
export GOOS=linux && go build -o installer ${ROOT_DIR}/cmd/operator/main.go

if [[ "${KYMA_PATH}" ]]; then
		echo "Building docker image"
    rm -rf kyma
		mkdir kyma
    cp -r ${KYMA_PATH}/resources kyma
    cp -r ${KYMA_PATH}/installation kyma

    eval $(minikube docker-env --shell bash)
    docker build -t ${IMAGE_NAME} -f deploy/installer/Dockerfile .
fi

popd
