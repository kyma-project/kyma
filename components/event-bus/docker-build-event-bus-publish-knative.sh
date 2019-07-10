#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image..."

mkdir -p ./cmd/event-bus-publish-knative/docker/image/
mkdir -p ./cmd/event-bus-publish-knative/docker/image/api
mkdir -p ./cmd/event-bus-publish-knative/docker/image/pkg
mkdir -p ./cmd/event-bus-publish-knative/docker/image/vendor
mkdir -p ./cmd/event-bus-publish-knative/docker/image/internal

cp -R ./vendor            ./cmd/event-bus-publish-knative/docker/image/
cp -R ./api/publish/      ./cmd/event-bus-publish-knative/docker/image/api/publish/
cp -R ./api/push/      ./cmd/event-bus-publish-knative/docker/image/api/push/
cp -R ./pkg      ./cmd/event-bus-publish-knative/docker/image/
cp -R ./internal/trace/   ./cmd/event-bus-publish-knative/docker/image/internal/trace/
cp -R ./internal/knative/ ./cmd/event-bus-publish-knative/docker/image/internal/knative/
cp -R ./internal/ea/ ./cmd/event-bus-publish-knative/docker/image/internal/ea/
cp -R ./licenses ./cmd/event-bus-publish-knative/docker/image/

cp -R ./cmd/event-bus-publish-knative/main.go     ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/application ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/handlers    ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/httpserver  ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/publisher   ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/validators  ./cmd/event-bus-publish-knative/docker/image/
cp -R ./cmd/event-bus-publish-knative/metrics     ./cmd/event-bus-publish-knative/docker/image/
cp    ./cmd/event-bus-publish-knative/Dockerfile  ./cmd/event-bus-publish-knative/docker/image/

tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/event-bus-publish-knative/docker/image
rm -rf ./cmd/event-bus-publish-knative/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully"
