#!/bin/sh

COLOR='\033[0;36m'
NO_COLOR='\033[0m'

APP_NAME="gqlgen"
PACKAGE_DIR="./vendor/github.com/99designs/$APP_NAME"
BASE_DIR="./internal/gqlschema"

echo "${COLOR}Building generator...${NO_COLOR}"
go build -o ${BASE_DIR}/${APP_NAME} ${PACKAGE_DIR}

echo "${COLOR}Generating code from GraphQL schema...${NO_COLOR}"
cd ${BASE_DIR}
./${APP_NAME} -v --config ./config.yml

echo "${COLOR}Cleaning up...${NO_COLOR}"
rm ./${APP_NAME}
