#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image..."

mkdir -p ./cmd/event-publish-service/docker/image/
mkdir -p ./cmd/event-publish-service/docker/image/api
mkdir -p ./cmd/event-publish-service/docker/image/pkg
mkdir -p ./cmd/event-publish-service/docker/image/vendor
mkdir -p ./cmd/event-publish-service/docker/image/internal

cp -R ./vendor            ./cmd/event-publish-service/docker/image/
cp -R ./api/publish/      ./cmd/event-publish-service/docker/image/api/publish/
cp -R ./api/push/      ./cmd/event-publish-service/docker/image/api/push/
cp -R ./internal/trace/   ./cmd/event-publish-service/docker/image/internal/trace/
cp -R ./internal/knative/ ./cmd/event-publish-service/docker/image/internal/knative/
cp -R ./internal/ea/ ./cmd/event-publish-service/docker/image/internal/ea/
cp -R ./licenses ./cmd/event-publish-service/docker/image/

cp -R ./cmd/event-publish-service/main.go     ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/application ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/handlers    ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/httpserver  ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/publisher   ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/validators  ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/metrics     ./cmd/event-publish-service/docker/image/
cp -R ./cmd/event-publish-service/util        ./cmd/event-publish-service/docker/image/
cp    ./cmd/event-publish-service/Dockerfile  ./cmd/event-publish-service/docker/image/

tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/event-publish-service/docker/image
rm -rf ./cmd/event-publish-service/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully"
