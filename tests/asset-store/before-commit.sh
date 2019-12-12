#!/usr/bin/env bash

readonly CI_FLAG=ci

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
# GO BUILD
##
buildEnv=""
if [ "$1" == "$CI_FLAG" ]; then
	# build binary statically
	buildEnv="env CGO_ENABLED=0"
fi

${buildEnv} go test ./main_test.go -c
goBuildResult=$?
rm main.test

if [ ${goBuildResult} != 0 ]; then
	echo -e "${RED}✗ go build${NC}\n$goBuildResult${NC}"
	exit 1
else echo -e "${GREEN}√ go build${NC}"
fi

##
# DEP
##
echo "? dep status"
depResult=$(dep status -v)
if [ $? != 0 ]
    then
        echo -e "${RED}✗ dep status\n$depResult${NC}"
        exit 1;
    else  echo -e "${GREEN}√ dep status${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test ./pkg/...
# Check if tests passed
if [ $? != 0 ];
    then
    	echo -e "${RED}✗ go test\n${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go test${NC}"
fi

filesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")
#
# GO IMPORTS
#
go build -o goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
goImportsResult=$(echo "${filesToCheck}" | xargs -L1 ./goimports-vendored -w -l)
rm goimports-vendored

if [ $(echo ${#goImportsResult}) != 0 ]
	then
    	echo -e "${RED}✗ goimports ${NC}\n$goImportsResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ goimports ${NC}"
fi

#
# GO FMT
#
goFmtResult=$(echo "${filesToCheck}" | xargs -L1 go fmt)
if [ $(echo ${#goFmtResult}) != 0 ]
	then
    	echo -e "${RED}✗ go fmt${NC}\n$goFmtResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

#
# GO VET
#
goVetResult=$(echo "${filesToCheck}" | xargs -L1 go vet)
if [ $(echo ${#goVetResult}) != 0 ]
	then
    	echo -e "${RED}✗ go vet${NC}\n$goVetResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go vet${NC}"
fi

##
# ERRCHECK
##
go build -o errcheck-vendored ./vendor/github.com/kisielk/errcheck
buildErrCheckResult=$?
if [[ ${buildErrCheckResult} != 0 ]]; then
    echo -e "${RED}✗ go build errcheck${NC}\n${buildErrCheckResult}${NC}"
    exit 1
fi

errCheckResult=$(./errcheck-vendored -blank -asserts -ignoregenerated ./...)
rm errcheck-vendored

if [[ $(echo ${#errCheckResult}) != 0 ]]; then
    echo -e "${RED}✗ [errcheck] unchecked error in:${NC}\n${errCheckResult}${NC}"
    exit 1
else echo -e "${GREEN}√ errcheck ${NC}"
fi