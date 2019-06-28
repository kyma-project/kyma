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

goFilesToCheck=$(find . -type f -name "*.go" | egrep -v "/vendor")
##
# GO IMPORTS & FMT
##
go build -o goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
buildGoImportResult=$?
if [ ${buildGoImportResult} != 0 ]; then
	echo -e "${RED}✗ go build goimports${NC}\n$buildGoImportResult${NC}"
	exit 1
fi

goImportsResult=$(echo "${goFilesToCheck}" | xargs -L1 ./goimports-vendored -w -l)
rm goimports-vendored

if [ $(echo ${#goImportsResult}) != 0 ]; then
	echo -e "${RED}✗ goimports and fmt${NC}\n$goImportsResult${NC}"
	exit 1
else echo -e "${GREEN}√ goimports and fmt${NC}"
fi

##
# GO VET
##
packagesToVet=$(find . -type d -maxdepth 1 | egrep -v "/vendor|/pkg" | egrep "./")
packagesArray=(${packagesToVet//\\n/})
for vPackage in "${packagesArray[@]}"; do
	vetResult=$(go vet "${vPackage}/...")
	if [ $(echo ${#vetResult}) != 0 ]; then
		echo -e "${RED}✗ go vet ${vPackage}/... ${NC}\n$vetResult${NC}"
		exit 1
	else echo -e "${GREEN}√ go vet ${vPackage}/... ${NC}"
	fi
done

# goBuild compiles the Go projects
#
# params:
# `CGO_ENABLED=0` disable the use of cgo
# `-w` flag turns off DWARF debugging information
# `-s` turns off generation of the Go symbol table
function goBuild() {
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $1 $2
}
