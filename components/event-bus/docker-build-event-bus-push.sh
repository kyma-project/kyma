#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./cmd/event-bus-push/docker/image
cp ./cmd/event-bus-push/Dockerfile ./cmd/event-bus-push/docker/image/
cp -R ./cmd/event-bus-push/main.go ./cmd/event-bus-push/application ./cmd/event-bus-push/docker/image/
cp -R ./api ./cmd/event-bus-push/docker/image/
cp -R ./generated ./cmd/event-bus-push/docker/image/
cp -R ./internal ./cmd/event-bus-push/docker/image/
cp -R ./vendor ./cmd/event-bus-push/docker/image/
tagName="${NAME}:${VERSION}"
docker build --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/event-bus-push/docker/image
rm -rf ./cmd/event-bus-push/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
