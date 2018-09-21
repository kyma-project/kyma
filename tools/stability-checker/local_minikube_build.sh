#!/usr/bin/env bash

eval $(minikube docker-env --shell=bash)

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
# DEP ENSURE
##
dep ensure -v --vendor-only
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
	echo -e "${RED}✗ dep ensure -v --vendor-only${NC}\n$ensureResult${NC}"
	exit 1
else echo -e "${GREEN}√ dep ensure -v --vendor-only${NC}"
fi

##
# GO BUILD
##
binaries=("logs-printer" "stability-checker")
buildEnv=""

# build binary statically
buildEnv="env CGO_ENABLED=0 GOOS=linux"


for binary in "${binaries[@]}"; do
	${buildEnv} go build -o ${binary} ./cmd/${binary}
	goBuildResult=$?
	if [ ${goBuildResult} != 0 ]; then
		echo -e "${RED}✗ go build ${binary} ${NC}\n $goBuildResult${NC}"
		exit 1
	else echo -e "${GREEN}√ go build ${binary} ${NC}"
	fi
done

cp stability-checker deploy/stability-checker/stability-checker
cp logs-printer deploy/stability-checker/logs-printer

docker build -t local/stability-checker:local deploy/stability-checker