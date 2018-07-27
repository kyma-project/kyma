#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

echo $SCRIPT_ROOT

CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

echo $CODEGEN_PKG

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh  "deepcopy,client,informer,lister" \
  github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx github.com/kyma-project/kyma/components/api-controller/pkg/apis \
  gateway.kyma.cx:v1alpha2 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

${CODEGEN_PKG}/generate-groups.sh  "deepcopy,client,informer,lister" \
  github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io github.com/kyma-project/kyma/components/api-controller/pkg/apis \
  authentication.istio.io:v1alpha1 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

${CODEGEN_PKG}/generate-groups.sh  "deepcopy,client,informer,lister" \
  github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io github.com/kyma-project/kyma/components/api-controller/pkg/apis \
  networking.istio.io:v1alpha3 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../../../.." \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt