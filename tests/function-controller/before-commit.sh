#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + $USER
echo "PATH: " + $PATH
echo "GOPATH:" + $GOPATH
echo -e "${NC}"

##
# DEP ENSURE
##
depResult=$(dep ensure --vendor-only 2>&1)
if [ $? != 0 ]; then
	echo -e "${RED}✗ dep ensure --vendor-only${NC}\n$depResult${NC}"
	exit 1
else
	echo -e "${GREEN}√ dep ensure --vendor-only${NC}"
fi

##
# DEP STATUS
##
depResult=$(dep status 2>&1)
if [ $? != 0 ]; then
	echo -e "${RED}✗ dep status\n$depResult${NC}"
	exit 1
else
	echo -e "${GREEN}√ dep status${NC}"
fi

#
# GO FMT
#
goFilesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")

goFmtResult=$(echo "${goFilesToCheck}" | xargs -L1 go fmt)
if [ $(echo ${#goFmtResult}) != 0 ]; then
	echo -e "${RED}✗ go fmt${NC}\n$goFmtResult${NC}"
	exit 1
else
	echo -e "${GREEN}√ go fmt${NC}"
fi

#
# GO IMPORTS
#
filesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")
go build -o goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
goImportsResult=$(echo "${filesToCheck}" | xargs -L1 ./goimports-vendored -w -l)
rm goimports-vendored

if [[ $(echo ${#goImportsResult}) != 0 ]]; then
	echo -e "${RED}✗ goimports ${NC}\n$goImportsResult${NC}"
	exit 1
else
	echo -e "${GREEN}√ goimports ${NC}"
fi

##
# GO VET
##
packagesToVet=("./...")

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet ${vPackage} 2>&1)
	if [ $(echo ${#vetResult}) != 0 ]; then
		echo -e "${RED}✗ go vet ${vPackage} ${NC}\n$vetResult${NC}"
		exit 1
	else
		echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
	fi
done
