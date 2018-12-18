#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}
REB_ROOT_PKG="github.com/kyma-project/kyma/components/application-broker/pkg"

./hack/generate-groups.sh all \
  ${REB_ROOT_PKG}/client ${REB_ROOT_PKG}/apis \
  applicationconnector:v1alpha1 \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt