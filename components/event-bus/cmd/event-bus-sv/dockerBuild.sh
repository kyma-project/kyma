#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p docker/image
cp Dockerfile docker/image/
cp -R main.go application docker/image/
cp -R ../../generated docker/image/
cp -R ../../internal docker/image/
cp -R ../../vendor docker/image/
cp -R ../../api docker/image/
tagName="${NAME}:${VERSION}"
docker build --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm docker/image
rm -rf docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."