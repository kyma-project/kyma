#!/usr/bin/env bash

set -eux -o pipefail
resources=(
  "natsschannels.messaging.knative.dev"
)

for resource in "${resources[@]}"; do
  echo "starting deletion of ${resource} custom resources"
  if kubectl get crd "${resource}"; then
    kubectl \
      delete \
      --ignore-not-found \
      "${resource}" \
      --all-namespaces \
      --all
  else
    echo "CRD ${resource} unknown to apiserver"
  fi
  echo "finished deletion of ${resource} custom resources"
done