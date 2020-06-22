#!/usr/bin/env bash

set -eux -o pipefail
echo "starting deletion of knative-eventing core custom resources"
resources=(
  "apiserversources.sources.eventing.knative.dev"
  "containersources.sources.eventing.knative.dev"
  "cronjobsources.sources.eventing.knative.dev"
  "sinkbindings.sources.eventing.knative.dev"

  "eventtypes.eventing.knative.dev"
  "brokers.eventing.knative.dev"
  "triggers.eventing.knative.dev"

  "channels.messaging.knative.dev"
  "parallels.messaging.knative.dev"
  "sequences.messaging.knative.dev"
  "subscriptions.messaging.knative.dev"
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

echo "done"