#!/bin/bash -e

CURRENT_DIR=$(pwd)
GEN_DIR=$(dirname $0)
REPO_DIR="$CURRENT_DIR/$GEN_DIR/.."
echo $REPO_DIR

PROJECT_MODULE="github.com/kyma-project/kyma/components/compass-runtime-agent"
IMAGE_NAME="kubernetes-codegen:latest"

CUSTOM_RESOURCE_GROUP_DIR="compass"
CUSTOM_RESOURCE_VERSION="v1alpha1"

echo "Building codegen Docker image..."
docker build -f "${GEN_DIR}/Dockerfile" \
             -t "${IMAGE_NAME}" \
             "${REPO_DIR}"

PROJECT_IN_GOPATH="/go/src/${PROJECT_MODULE}"

copy_cmd="cp -r ${PROJECT_IN_GOPATH}/vendor ."

cmd="./generate-groups.sh "deepcopy,client,informer,lister" \
    "$PROJECT_MODULE/pkg/client" \
    "$PROJECT_MODULE/pkg/apis" \
    $CUSTOM_RESOURCE_GROUP_DIR:$CUSTOM_RESOURCE_VERSION \
    --go-header-file ${PROJECT_IN_GOPATH}/hack/boilerplate.go.txt"

echo "Generating client codes..."
docker run -it --rm \
           -v "${REPO_DIR}:${PROJECT_IN_GOPATH}" \
           -e PROJECT_MODULE=${PROJECT_MODULE} \
           "${IMAGE_NAME}" $cmd

