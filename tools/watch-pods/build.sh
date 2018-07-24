#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

function onError() {
    echo -e "${RED}Script exit with error(s) ${NC}"
}

trap onError ERR

dep ensure -v -vendor-only

echo -e "${GREEN}Build app${NC}"
GCO_ENABLED=0 GOOS=linux go build -o bin/app

echo -e "${GREEN}Run tests ${NC}"
go test ./internal/...

echo -e "${GREEN}Run go vet ${NC}"
go vet ./...

echo -e "${GREEN}Download golint ${NC}"
go get -u github.com/golang/lint/golint

echo -e "${GREEN}Run golint ${NC}"
golint -set_exit_status ./internal/...

echo -e "${GREEN}Run gofmt ${NC}"
gofmt -l -w main.go
gofmt -l -w ./internal/

echo -e "${GREEN}Build image${NC}"
docker build -t watch-pods:latest .

