#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}
ROOT_PKG="github.com/kyma-project/kyma/components/ui-api-layer/pkg"
API_TYPE_VERSION="ui:v1alpha1"

go install ./vendor/k8s.io/code-generator/cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}

./vendor/k8s.io/code-generator/generate-groups.sh all \
  ${ROOT_PKG}/client ${ROOT_PKG}/apis \
  ${API_TYPE_VERSION} \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt
