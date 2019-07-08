#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./e2e-subscriber/docker/image
cp ./e2e-subscriber/Dockerfile ./e2e-subscriber/docker/image/
cp -R ./e2e-subscriber/e2e-subscriber.go ./e2e-subscriber/docker/image/
cp -R ./vendor ./e2e-subscriber/docker/image/
cp -R ./licenses ./e2e-subscriber/docker/image/
	
tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./e2e-subscriber/docker/image
rm -rf ./e2e-subscriber/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
