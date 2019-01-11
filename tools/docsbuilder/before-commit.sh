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
if [ "$1" == "$CI_FLAG" ]; then
	# build binary statically
	buildEnv="env CGO_ENABLED=0"
fi

${buildEnv} go build -o main
goBuildResult=$?
rm main

if [ ${goBuildResult} != 0 ]; then
	echo -e "${RED}✗ go build${NC}\n$goBuildResult${NC}"
	exit 1
else echo -e "${GREEN}√ go build${NC}"
fi

##
# DEP STATUS
##
echo "? dep status"
dep status -v
check_result "dep status" $?

##
# GO TEST
##
echo "? go test"
go test ./...
check_result "go test" $?

goFilesToCheck=$(find . -type f -name "*.go" | grep -E -v "/vendor/|/automock/|/testdata/")

##
#  GO LINT
##
echo "? golint"
go build -o golint-vendored ./vendor/github.com/golang/lint/golint
check_result "go build lint" $?

golintResult=$(echo "${goFilesToCheck}" | xargs -L1 ./golint-vendored)
rm golint-vendored

check_result "golint" "${#golintResult}" "${golintResult}"

##
# GO IMPORTS & FMT
##
echo "? goimports and fmt"
go build -o goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
check_result "go build goimports" $?

goImportsResult=$(echo "${goFilesToCheck}" | xargs -L1 ./goimports-vendored -w -l)
rm goimports-vendored

check_result "goimports and fmt" "${#goImportsResult}" "${goImportsResult}"

##
# GO VET
##
echo "? go vet"
packagesToVet=("./internal/...")

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet "${vPackage}")
    check_result "go vet ${vPackage}" "${#vetResult}" "${vetResult}"
done