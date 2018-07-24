#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e #terminate script immediately in case of errors

##
# DEP ENSURE
##
dep ensure -v --vendor-only
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
	echo -e "${RED}✗ dep ensure -v --vendor-only${NC}\n$ensureResult${NC}"
	exit 1
else echo -e "${GREEN}√ dep ensure -v --vendor-only${NC}"
fi

fmtResult=$(go fmt ./...)
if [ $(echo ${#fmtResult}) != 0 ]
	then
    	echo -e "${RED}✗ go fmt${NC}\n${fmtResult}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

# goBuild compiles the Go projects
#
# params:
# `CGO_ENABLED=0` disable the use of cgo
# `-w` flag turns off DWARF debugging information
# `-s` turns off generation of the Go symbol table
function goBuild() {
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $1 $2
}

echo -e "${GREEN} Building fake-gateway${NC}"
goBuild gateway.bin ./remote-environment/cmd/fake-gateway/main.go

echo -e "${GREEN} Building gateway-client${NC}"
goBuild client.bin ./remote-environment/cmd/gateway-client/main.go

echo -e "${GREEN} Building env-tester${NC}"
goBuild env-tester.bin ./servicecatalog/cmd/env-tester/main.go
