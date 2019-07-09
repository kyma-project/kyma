#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

fmtResult=$(go fmt ./...)
if [ $(echo ${#fmtResult}) != 0 ]
	then
    	echo -e "${RED}✗ go fmt${NC}\n${fmtResult}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

depResult=$(dep status -v)
if [ $? != 0 ]
    then
        echo -e "${RED}✗ dep status\n$depResult${NC}"
        exit 1;
    else  echo -e "${GREEN}√ dep status${NC}"
fi
