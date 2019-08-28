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
# DEP ENSURE
##
dep ensure --v --vendor-only
check_result "dep ensure --v --vendor-only" $?

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
    ${buildEnv} go build -o "bin/${cmdName}" "${directory}"
    buildResult=$?
    check_result "go build ${directory}" "${buildResult}"
done <   <(find "./cmd" -mindepth 1 -type d -print0)

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
go test -count=1 ./...
check_result "go test" $?

goFilesToCheck=$(find . -type f -name "*.go" | grep -E -v "/vendor/|/automock/|/testdata/")

##
# GO LINT
##
echo "? golint"
go build -o golint-vendored ./vendor/golang.org/x/lint/golint
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
packagesToVet=("./cmd/..." "./pkg/...")

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet "${vPackage}")
    check_result "go vet ${vPackage}" "${#vetResult}" "${vetResult}"
done