#!/usr/bin/env bash

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TIMEOUT=300

export NAMESPACE=ory-perf-test
export CLUSTER_DOMAIN=$(kubectl get gateways.networking.istio.io kyma-gateway \
                        -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*.//g' )
export CLIENT_ID=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-z' | fold -w 8 | head -n 1 | base64)
export CLIENT_SECRET=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-z' | fold -w 32 | head -n 1 | base64)

resources=(
  namespace.yaml
  deployment.yaml
  oauth2client.yaml
  apirules.yaml
)

# wait for resource with the given name until it exists or there is a timeout
function waitFor() {
    echo "waiting for $1/$2"
    start=$(date +%s)
    while true; do
        # run until command finishes with exitcode=0
        if kubectl get "$1" -n "$NAMESPACE" "$2" >/dev/null 2>&1; then
            break
        fi
        current_time=$(date +%s)
        timeout_time=$((start + TIMEOUT))
        # or timeout occurrs
        if ((current_time > timeout_time)); then
            echo "error: timeout waiting for $1/$2"
            exit 1
        else
            echo -n "."
        fi
        sleep 1
    done
}

for resource in "${resources[@]}"; do
    echo "deploying: $resource"
    envsubst <"${WORKING_DIR}/$resource" | kubectl -n "$NAMESPACE" apply -f -
done

