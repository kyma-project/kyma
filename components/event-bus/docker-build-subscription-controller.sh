#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./cmd/subscription-controller/docker/image
cp ./cmd/subscription-controller/Dockerfile ./cmd/subscription-controller/docker/image/
cp -R ./cmd/subscription-controller/cmd ./cmd/subscription-controller/docker/image/
cp -R ./generated ./cmd/subscription-controller/docker/image/
cp -R ./internal ./cmd/subscription-controller/docker/image/
cp -R ./vendor ./cmd/subscription-controller/docker/image/
cp -R ./api ./cmd/subscription-controller/docker/image/
cp -R ./licenses ./cmd/subscription-controller/docker/image/
tagName="${NAME}:${VERSION}"
docker build --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/subscription-controller/docker/image
rm -rf ./cmd/subscription-controller/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
