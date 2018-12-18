#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./cmd/event-bus-publish/docker/image
cp ./cmd/event-bus-publish/Dockerfile ./cmd/event-bus-publish/docker/image/
cp -R ./cmd/event-bus-publish/main.go ./cmd/event-bus-publish/application ./cmd/event-bus-publish/controllers ./cmd/event-bus-publish/handlers ./cmd/event-bus-publish/docker/image/
cp -R ./api ./cmd/event-bus-publish/docker/image/
cp -R ./internal ./cmd/event-bus-publish/docker/image/
cp -R ./vendor ./cmd/event-bus-publish/docker/image/
tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/event-bus-publish/docker/image
rm -rf ./cmd/event-bus-publish/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
