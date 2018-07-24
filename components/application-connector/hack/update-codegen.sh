#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}
METADATA_ROOT_PKG="github.com/kyma-project/kyma/components/application-connector/pkg"

./hack/generate-groups.sh all \
  ${METADATA_ROOT_PKG}/client ${METADATA_ROOT_PKG}/apis \
  istio:v1alpha2 \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
