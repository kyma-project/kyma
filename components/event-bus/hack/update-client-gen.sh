#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/kyma-project/kyma/components/event-bus/generated/push \
  github.com/kyma-project/kyma/components/event-bus/api/push \
  "eventing.kyma-project.io:v1alpha1" \
  --go-header-file hack/boilerplate/boilerplate.go.txt
