#!/usr/bin/env bash

readonly CI_FLAG=ci
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly INVERTED='\033[7m'
readonly NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: ${USER}"
echo "PATH: ${PATH}"
echo "GOPATH: ${GOPATH}"
echo -e "${NC}"

function check_result() {
    local step=$1
    local result=$2
    local output=$3

    if [[ ${result} != 0 ]]; then
        echo -e "${RED}✗ ${step}${NC}\\n${output}"
        exit 1
    else
        echo -e "${GREEN}√ ${step}${NC}"
    fi
}

##
# GO BUILD
##

echo "? go build"
buildEnv=""
if [[ "$1" == "$CI_FLAG" ]]; then
	# build binary statically for linux architecture
	buildEnv="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64"
fi

while IFS= read -r -d '' directory
do
    cmdName=$(basename "${directory}")
    ${buildEnv} GO111MODULE=on go build -o "bin/${cmdName}" "${directory}"
    buildResult=$?
    check_result "go build ${directory}" "${buildResult}"
done <   <(find "./cmd" -mindepth 1 -type d -print0)

##
# GO TEST
##
echo "? go test"
GO111MODULE=on go test -count=1 -coverprofile=cover.out ./...
echo -e "Total coverage: $(go tool cover -func=cover.out | grep total | awk '{print $3}')"
rm cover.out
check_result "go test" $?

goFilesToCheck=$(find . -type f -name "*.go" | grep -E -v "/vendor/|/automock/|/testdata/")
