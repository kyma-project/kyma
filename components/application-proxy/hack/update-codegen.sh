#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}
APPLICATION_PROXY_ROOT_PKG="github.com/kyma-project/kyma/components/application-proxy/pkg"

./hack/generate-groups.sh all \
  ${APPLICATION_PROXY_ROOT_PKG}/client ${APPLICATION_PROXY_ROOT_PKG}/apis \
  remoteenvironment:v1alpha1 \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
