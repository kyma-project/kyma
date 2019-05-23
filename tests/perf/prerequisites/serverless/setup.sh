#!/usr/bin/env bash

set -e
set -o pipefail

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export NAMESPACE=serverless
export FUNC_DELAY=200 # delay in ms
resources=(
    namespace.yaml
    resource_quota.yaml
    function-size-s.yaml
    function-size-m.yaml
    function-size-l.yaml
    # function-size-xl.yaml
)

# wait for resource with the given name until it exists or there is a timeout
function waitFor() {
    echo "waiting for $1/$2"
    start=$(date +%s)
    timeout=60 # seconds
    while true; do
        # run until command finishes with exitcode=0
        if kubectl get "$1" -n "$NAMESPACE" "$2" >/dev/null 2>&1; then
            break
        fi
        current_time=$(date +%s)
        timeout_time=$((start + timeout))
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
    envsubst <"${SCRIPTS_DIR}/$resource" | kubectl -n "$NAMESPACE" apply -f -
done

# wait for resources to be ready
kubectl wait --timeout=30s --for=condition=Available -n "$NAMESPACE" deployment/size-s
kubectl wait --timeout=30s --for=condition=Available -n "$NAMESPACE" deployment/size-m
kubectl wait --timeout=30s --for=condition=Available -n "$NAMESPACE" deployment/size-l
# kubectl wait --timeout=30s --for=condition=Available -n "$NAMESPACE" deployment/size-xl

waitFor "hpa" "size-s"
waitFor "hpa" "size-m"
waitFor "hpa" "size-l"
# waitFor "hpa" "size-xl"

waitFor "function" "size-s"
waitFor "function" "size-m"
waitFor "function" "size-l"
# waitFor "function" "size-xl"

waitFor "service" "size-s"
waitFor "service" "size-m"
waitFor "service" "size-l"
# waitFor "service" "size-xl"

waitFor "api" "size-s"
waitFor "api" "size-m"
waitFor "api" "size-l"
# waitFor "api" "size-xl"
