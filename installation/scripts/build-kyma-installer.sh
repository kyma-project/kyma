#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
IMAGE_NAME=eu.gcr.io/kyma-project/develop/kyma-installer:63f27f76

eval $(minikube docker-env --shell bash)
docker build -t ${IMAGE_NAME} -f ${CURRENT_DIR}/../../kyma-installer/kyma.Dockerfile .
