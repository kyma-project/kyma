
#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd ${GOPATH}/src/k8s.io/code-generator
./generate-groups.sh all github.com/kyma-project/kyma/components/ui-api-layer/pkg/generated github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis uiapi.kyma-project.io:v1alpha1