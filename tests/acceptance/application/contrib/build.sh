#!/usr/bin/env bash

TAG=0.1
PROJECT=kyma

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e #terminate script immediately in case of errors

eval $(minikube docker-env --shell bash)

echo -e "${GREEN} Building test${NC}"
GOOS=linux GOARCH=amd64 go test -v -c -o re.test ./application/re_test.go

echo -e "${GREEN} Building gw${NC}"
GOOS=linux GOARCH=amd64 go build -o gateway.bin ./application/cmd/fake-gateway/main.go

echo -e "${GREEN} Building tester${NC}"
GOOS=linux GOARCH=amd64 go build -o client.bin ./application/cmd/gateway-client/main.go

IMAGE_NAME=acceptance-tests-re:${TAG}
docker build -t ${IMAGE_NAME} -f application/contrib/Dockerfile .
docker tag ${IMAGE_NAME} ${PROJECT}/${IMAGE_NAME}

rm gateway.bin
rm client.bin
rm re.test