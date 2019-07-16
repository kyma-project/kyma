#!/usr/bin/env bash
set -e
set -o pipefail

NAME=$1
VERSION=$2
COMPONENT=$3

echo -e "Start building docker image...."
mkdir -p ./e2e-tester/docker/image
cp ./e2e-tester/Dockerfile ./e2e-tester/docker/image/
cp -R ./e2e-tester/e2e-tester.go ./e2e-tester/docker/image/
cp -R ./vendor ./e2e-tester/docker/image/
cp -R ./licenses ./e2e-tester/docker/image/
	
tagName="${NAME}:${VERSION}"
docker build --no-cache --build-arg version=${VERSION} -t ${tagName} --label version=${VERSION} --label component=${COMPONENT} --rm ./e2e-tester/docker/image
rm -rf ./e2e-tester/docker
echo -e "Docker image with the tag [ ${tagName} ] has been built successfully ..."
