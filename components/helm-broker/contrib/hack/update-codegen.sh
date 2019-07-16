#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly CURRENT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${CURRENT_DIR}/../../; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}
REB_ROOT_PKG="github.com/kyma-project/kyma/components/helm-broker/pkg"

${CURRENT_DIR}/generate-groups.sh all \
  ${REB_ROOT_PKG}/client ${REB_ROOT_PKG}/apis \
  addons:v1alpha1 \
  --go-header-file ${CURRENT_DIR}/boilerplate.go.txt