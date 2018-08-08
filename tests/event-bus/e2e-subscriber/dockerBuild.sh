#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p docker/image
cp Dockerfile docker/image/
cp -R e2e-subscriber.go docker/image/
cp -R ../vendor docker/image/
	
tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm docker/image
rm -rf docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
