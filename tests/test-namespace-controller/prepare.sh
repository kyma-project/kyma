#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
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

##
# GO FMT
##
fmtResult=$(go fmt ./...)
if [ $(echo ${#fmtResult}) != 0 ]
	then
    	echo -e "${RED}✗ go fmt${NC}\n${fmtResult}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
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