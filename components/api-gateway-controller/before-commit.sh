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

${buildEnv} go build -o bin/manager main.go

goBuildResult=$?
if [ ${goBuildResult} != 0 ]; then
	echo -e "${RED}✗ go build${NC}\n$goBuildResult${NC}"
	exit 1
else echo -e "${GREEN}√ go build${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test -short ./...
# Check if tests passed
if [ $? != 0 ]; then
	echo -e "${RED}✗ go test\n${NC}"
	exit 1
else echo -e "${GREEN}√ go test${NC}"
fi

##
# GO FMT
##

go fmt ./...

##
# GO VET
##
packagesToVet=($(go list ./... | grep -v /vendor/ | grep -v api-controller/pkg/clients ))

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet ${vPackage})
	if [ $(echo ${#vetResult}) != 0 ]; then
		echo -e "${RED}✗ go vet ${vPackage} ${NC}\n$vetResult${NC}"
		exit 1
	else echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
	fi
done

##
# INFO.JSON
##
author=$(git show -s --pretty=%an)
branch=$(git rev-parse --abbrev-ref HEAD)
commit=$(git rev-parse --verify HEAD)
commitMsg=$(git show -s --pretty=%s)
commitDate=$(git log -1 --format=%cd)
printf "{\n\t\"author\": \"""$author""\",\n\t\"commit\": \""$commit"\",\n\t\"branch\": \""$branch"\",\n\t\"commitDate\":\"""$commitDate""\",\n\t\"commitMessage\":\"""$commitMsg""\",\n\t\"deployDate\": \"""$(date)""\"\n}" >info.json
echo -e "${GREEN}√ info.json ${NC}" $(cat info.json)
