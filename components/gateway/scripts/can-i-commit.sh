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
# DEP STATUS
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
# MAKE BUILD
##
echo "? make build"
( make build )
if [ $? != 0 ]; # Check make build passed
    then
    	echo -e "${RED}✗ make build\n${NC}"
    	exit 1;
	else echo -e "${GREEN}√ make build${NC}"
fi

filesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")

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

echo -e "${GREEN}Congrats $(whoami)! You've made it! Now you can commit.${NC}"
