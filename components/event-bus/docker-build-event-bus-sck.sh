#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./cmd/event-bus-subscription-controller-knative/docker/image
cp ./cmd/event-bus-subscription-controller-knative/Dockerfile ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./cmd/event-bus-subscription-controller-knative/cmd ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./generated ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./internal ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./vendor ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./api ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./pkg ./cmd/event-bus-subscription-controller-knative/docker/image/
cp -R ./licenses ./cmd/event-bus-subscription-controller-knative/docker/image/
tagName="${NAME}:${VERSION}"
docker build --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/event-bus-subscription-controller-knative/docker/image
rm -rf ./cmd/event-bus-subscription-controller-knative/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
