#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [[ -z "${NAMESPACE}" ]]; then
  export NAMESPACE=istio-perf-test
fi

export CLUSTER_DOMAIN=$(kubectl get gateways.networking.istio.io kyma-gateway \
                        -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*.//g' )

resources=(
  namespace.yaml
  deployment.yaml
  apirule.yaml
)

for resource in "${resources[@]}"; do
    echo "deploying: $resource"
    envsubst <"${WORKING_DIR}/$resource" | kubectl -n "$NAMESPACE" apply -f -
done

sleep 3s