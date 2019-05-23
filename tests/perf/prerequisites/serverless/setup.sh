#!/usr/bin/env bash

set -e
set -o pipefail

WORKING_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
TIMEOUT=300 # seconds

export NAMESPACE=serverless
export FUNC_DELAY=200 # delay in ms

resources=(
    namespace.yaml
    resource_quota.yaml
    function-size-s.yaml
    function-size-m.yaml
    # function-size-l.yaml
    # function-size-xl.yaml
)
functions=(
    size-s
    size-m
    # size-l
    # size-xl
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

for func in "${functions[@]}"; do
    kubectl wait --timeout="$TIMEOUT"s --for=condition=Available -n "$NAMESPACE" "deployment/$func"
    waitFor "hpa" "$func"
    waitFor "function" "$func"
    waitFor "service" "$func"
    waitFor "api" "$func"
done
