#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(dirname ${BASH_SOURCE})/..

echo "Installing latest mockery..."
go get github.com/vektra/mockery/.../
echo "Installing latest failery..."
go get github.com/kyma-project/kyma/tools/failery/.../
echo "Generating mock implementation for interfaces..."
cd ${PROJECT_ROOT}
go generate ./...
