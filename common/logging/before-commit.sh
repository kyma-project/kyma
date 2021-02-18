#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + ${USER}
echo "PATH: " + ${PATH}
echo "GOPATH:" + ${GOPATH}
echo -e "${NC}"

##
# GO MOD VENDOR
##
echo "? go mod vendor"
GO111MODULE=on go mod vendor
vendorResult=$?
if [[ ${vendorResult} != 0 ]]; then
	echo -e "${RED}✗ go mod vendor${NC}\n$vendorResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod vendor${NC}"
fi

##
# GO MOD VERIFY
##
echo "? go mod verify"
verifyResult=$(GO111MODULE=on go mod verify)
if [[ $? != 0 ]]; then
	echo -e "${RED}✗ go mod verify\n$verifyResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod verify${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test -count 100 -race -coverprofile=cover.out ./...
# Check if tests passed
if [[ $? != 0 ]]; then
	echo -e "${RED}✗ go test\n${NC}"
	rm cover.out
	exit 1
else 
	echo -e "Total coverage: $(go tool cover -func=cover.out | grep total | awk '{print $3}')"
	rm cover.out
	echo -e "${GREEN}√ go test${NC}"
fi


##
# GO IMPORTS & FMT
##
cd ..
go build -o ./logging/goimports-vendored ./vendor/golang.org/x/tools/cmd/goimports
buildGoImportResult=$?
if [[ ${buildGoImportResult} != 0 ]]; then
	echo -e "${RED}✗ go build goimports${NC}\n$buildGoImportResult${NC}"
	exit 1
fi


cd ./logging
goFilesToCheck=$(find . -type f -name "*.go" | egrep -v "/vendor")
goImportsResult=$(echo "${goFilesToCheck}" | xargs -L1 ./goimports-vendored -w -l)
rm goimports-vendored

if [[ $(echo ${#goImportsResult}) != 0 ]]; then
	echo -e "${RED}✗ goimports and fmt${NC}\n$goImportsResult${NC}"
	exit 1
else echo -e "${GREEN}√ goimports and fmt${NC}"
fi

##
# GO VET
##
packagesToVet=($(go list ./... | grep -v /vendor/))

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet ${vPackage})
	if [[ $(echo ${#vetResult}) != 0 ]]; then
		echo -e "${RED}✗ go vet ${vPackage} ${NC}\n$vetResult${NC}"
		exit 1
	else echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
	fi
done
