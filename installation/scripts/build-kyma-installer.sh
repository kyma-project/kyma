#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
IMAGE_NAME="$(${CURRENT_DIR}/extract-kyma-installer-image.sh)"

eval $(minikube docker-env --shell bash)
docker build -t ${IMAGE_NAME} -f ${CURRENT_DIR}/../../kyma-installer/kyma.Dockerfile .
