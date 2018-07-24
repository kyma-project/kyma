#!/bin/sh
COLOR='\033[0;36m'
NO_COLOR='\033[0m'

APP_NAME='gqlgen'
REPOSITORY="github.com/vektah/$APP_NAME"
TEMP_DIR="./temp_$APP_NAME"

PACKAGE='gqlschema'
SCHEMA_DIR="./internal/$PACKAGE"

echo "${COLOR}Building generator...${NO_COLOR}"
go build -o $TEMP_DIR/$APP_NAME ./vendor/$REPOSITORY

echo "${COLOR}Generating code from GraphQL schema...${NO_COLOR}"
$TEMP_DIR/$APP_NAME -schema $SCHEMA_DIR/schema.graphql -typemap $SCHEMA_DIR/types.json -out $SCHEMA_DIR/schema_gen.go -models $SCHEMA_DIR/models_gen.go -package $PACKAGE

echo "${COLOR}Cleaning up...${NO_COLOR}"
rm -rf $TEMP_DIR