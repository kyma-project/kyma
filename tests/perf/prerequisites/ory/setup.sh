#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TIMEOUT=300

export NAMESPACE=ory-perf-test
export CLUSTER_DOMAIN=$(kubectl get gateways.networking.istio.io kyma-gateway \
                        -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*.//g' )

resources=(
  namespace.yaml
  deployment.yaml
  oauth2client.yaml
  apirules.yaml
)

for resource in "${resources[@]}"; do
    echo "deploying: $resource"
    envsubst <"${WORKING_DIR}/$resource" | kubectl -n "$NAMESPACE" apply -f -
done

sleep 3s

export CLIENT_ID="$(kubectl get secret -n $NAMESPACE perf-tests-secret -o jsonpath='{.data.client_id}' | base64 --decode)"
export CLIENT_SECRET="$(kubectl get s ecret -n $NAMESPACE perf-tests-secret -o jsonpath='{.data.client_secret}' | base64 --decode)"