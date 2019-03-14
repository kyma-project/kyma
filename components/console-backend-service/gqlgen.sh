#!/bin/sh

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./internal/gqlschema
go run ../../hack/gqlgen.go -v --config ./config.yml
