#!/usr/bin/env bash

set -e
set -o pipefail

export NAMESPACE=serverless
resources=(
    namespace.yaml
    function-size-s.yaml
    function-size-m.yaml
    function-size-l.yaml
    function-size-xl.yaml
)

# TODO: XXX can be replaced by kubectl wait -f {filename}
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
    envsubst <"$PREREQ_PATH/serverless/$resource" | kubectl -n "$NAMESPACE" apply -f -
done

# wait for functions to be ready
kubectl wait --for=condition=Ready -n "$NAMESPACE" -l function=size-s pod
kubectl wait --for=condition=Ready -n "$NAMESPACE" -l function=size-m pod
kubectl wait --for=condition=Ready -n "$NAMESPACE" -l function=size-l pod
kubectl wait --for=condition=Ready -n "$NAMESPACE" -l function=size-xl pod

waitFor "hpa" "size-s"
waitFor "hpa" "size-m"
waitFor "hpa" "size-l"
waitFor "hpa" "size-xl"

waitFor "function" "size-s"
waitFor "function" "size-m"
waitFor "function" "size-l"
waitFor "function" "size-xl"

waitFor "service" "size-s"
waitFor "service" "size-m"
waitFor "service" "size-l"
waitFor "service" "size-xl"

waitFor "api" "size-s"
waitFor "api" "size-m"
waitFor "api" "size-l"
waitFor "api" "size-xl"
