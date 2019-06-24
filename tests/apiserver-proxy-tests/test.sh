#!/usr/bin/env bash

set -e

if [[ -z "${POD_NAME}" ]]; then
    echo "POD_NAME not provided"
    echo "TEST FAILED"
    exit 1
fi

AUTH_TOKEN=$(/root/app)

getConfigFile() {
	curl -s -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}/kube-config" -o "${PWD}/kubeconfig"
}

test(){
    local retry=$1
    local maxRetries=$2
    if [[ "$retry" -ge "$maxRetries" ]]; then
    	echo "TEST FAILED"
    	exit 1
    fi
    echo "Try $retry/$maxRetries"

    UUID=$(cat /proc/sys/kernel/random/uuid)
    echo "${UUID}" > "${PWD}/uuid"

    local out
    local status
    set +e
    out=$(kubectl exec "${POD_NAME}" cat ${PWD}/uuid 2>&1)
    status=$?
    set -e

    if [[ "$status" -ne 0 ]] || [[ "${UUID}" != $(echo "${out}" | tail -n 1) ]]; then
        echo "kubectl exec error ($status):"
        echo "${out}"
        echo "---"
        echo "UUID: $(cat ${PWD}/uuid)"
        test "$((retry+1))" "${maxRetries}"
    else
        echo "TEST SUCCEEDED"
    fi
}

getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"
test 0 2
