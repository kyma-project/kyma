#!/bin/sh

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./internal/gqlschema
GO111MODULE=off go run ../../hack/gqlgen.go --verbose --config ./config.yml
