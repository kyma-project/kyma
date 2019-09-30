#!/usr/bin/env bash

readonly CI_FLAG=ci

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + $USER
echo "PATH: " + $PATH
echo -e "${NC}"

##
# GET DEPENDENCIES
##
go mod download
goModResult=$?
if [ ${goModResult} != 0 ]; then
	echo -e "${RED}✗ go mod download${NC}\n$goModResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod download${NC}"
fi

go mod vendor
goModVendorResult=$?
if [ ${goModVendorResult} != 0 ]; then
	echo -e "${RED}✗ go mod vendor${NC}\n$goModVendorResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod vendor${NC}"
fi

go mod verify
goModVerifyResult=$?
if [ ${goModVerifyResult} != 0 ]; then
	echo -e "${RED}✗ go mod verify${NC}\ngoModVerifyResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod verify${NC}"
fi

##
# GO BUILD
##
buildEnv=""
if [[ "$1" == "$CI_FLAG" ]]; then
	# build binary statically
	buildEnv="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64"
fi

${buildEnv} go build -o compass-runtime-agent ./cmd
goBuildResult=$?
rm compass-runtime-agent

if [ ${goBuildResult} != 0 ]; then
	echo -e "${RED}✗ go build${NC}\n$goBuildResult${NC}"
	exit 1
else echo -e "${GREEN}√ go build${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test -short -coverprofile=cover.out ./...
# Check if tests passed
if [ $? != 0 ]; then
	echo -e "${RED}✗ go test\n${NC}"
	rm cover.out
	exit 1
else 
	echo -e "Total coverage: $(go tool cover -func=cover.out | grep total | awk '{print $3}')"
	rm cover.out
	echo -e "${GREEN}√ go test${NC}"
fi

#
# GO FMT
#
filesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")

goFmtResult=$(echo "${filesToCheck}" | xargs -L1 go fmt)
if [[ $(echo ${#goFmtResult}) != 0 ]]
	then
    	echo -e "${RED}✗ go fmt${NC}\n$goFmtResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

#
# GO VET
#
goVetResult=$(echo "${filesToCheck}" | xargs -L1 go vet)
if [[ $(echo ${#goVetResult}) != 0 ]]
	then
    	echo -e "${RED}✗ go vet${NC}\n$goVetResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go vet${NC}"
fi
