#!/usr/bin/env bash

set -e

if [[ -z "${POD_NAME}" ]]; then
    echo "POD_NAME not provided"
    echo "TEST FAILED"
    exit 1
fi

getConfigFile() {
    AUTH_TOKEN=$(/root/app)
	curl -s -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}:${IAM_KUBECONFIG_SVC_PORT}/kube-config" -o "${PWD}/kubeconfig"
}

test(){
    local retry=$1

    set +e
    OUT=$(kubectl exec "${POD_NAME}" -- /bin/bash -c 'sleep 10 && cat /etc/os-release' 2>&1)
    STATUS=$?
    set -e

    if [[ "$status" -ne 0 ]] || [[ -z "$OUT" ]]; then
        echo "Kubectl exec error!:"
        echo "Status: ${STATUS}"
        echo "Output: ${OUT} should NOT be an empty string!"
        echo "Execution ${retry}: Failure"
        exit 1
    fi

    echo "Execution ${retry}: Success"
    echo "${OUT}"
}

echo "---> Get kubeconfig from ${IAM_KUBECONFIG_SVC_FQDN}"
getConfigFile

export KUBECONFIG="${PWD}/kubeconfig"

for (( i = 1; i < (($MAX_TEST_RETRIES + 1)); i++ )); do
    echo "===> Try: ${i}/${MAX_TEST_RETRIES}"
    test ${i}
done

echo "---> All executions completed successfully"