#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image..."

mkdir -p ./cmd/nats-streaming-init/docker/image/

cp -R ./cmd/nats-streaming-init/scripts/prepare-config.sh ./cmd/nats-streaming-init/docker/image/

tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./cmd/nats-streaming-init/docker/image
rm -rf ./cmd/nats-streaming-init/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully"
