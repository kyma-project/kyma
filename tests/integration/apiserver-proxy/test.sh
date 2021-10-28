#!/usr/bin/env bash

set -e

reportErrorStatus() {
    if [ $? -ne 0 ]; then
        echo
        echo "===> TEST FAILED"
        echo " \=> $(date)"
    fi
}

trap reportErrorStatus EXIT

getConfigFile() {
    local CURL_CONN_TIMEOUT=5
    local CURL_CALL_RETRY_TIME=1
    local MAX_CURL_RETRIES=10
    local LAST_CALL_RES=0

    echo
    echo "===> Fetching Auth token"
    AUTH_TOKEN=$(/test/app)
    echo "===> Success: Auth token fetched!"
    echo " \=> $(date)"
    echo
    echo "===> Downloading kubeconfig file"

    set +e
    for (( i = 1; i < (($MAX_CURL_RETRIES + 1)); i++ )); do
        echo "===> Attempt: ${i}/${MAX_CURL_RETRIES}"
        curl --fail --connect-timeout ${CURL_CONN_TIMEOUT} -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}:${IAM_KUBECONFIG_SVC_PORT}/kube-config" -o "${PWD}/kubeconfig"
        LAST_CALL_RES=$?
        [[ $LAST_CALL_RES -eq 0 ]] && break
        echo
        echo "===> ERROR: Attempt ${i}/${MAX_CURL_RETRIES} failed."
        echo " \=> $(date)"
        echo
        sleep 1
    done
    if [ $LAST_CALL_RES -ne 0 ]; then
      echo "===> ERROR: Couldn't download kubeconfig file - all attempts failed."
      exit 1
    fi
    echo "===> Success: Attempt ${i}/${MAX_CURL_RETRIES}, kubeconfig file downloaded"
    echo " \=> $(date)"
    set -e
}


doTest() {
    local retry=$1

    set +e
    OUT=$(kubectl exec "${POD_NAME}" -- /bin/bash -c 'sleep 10 && cat /etc/os-release' 2>&1)
    STATUS=$?
    set -e

    if [[ "$status" -ne 0 ]] || [[ -z "$OUT" ]]; then
        echo "===> Kubectl exec error!:"
        echo "===> Status: ${STATUS}"
        echo "===> Output: ${OUT} should NOT be an empty string!"
        echo "===> Execution ${retry}: Failure"
        exit 1
    fi

    echo "${OUT}"
    echo "===> Success: Execution ${retry}"
		echo " \=> $(date)"
}

echo "===> Starting test"
echo " \=> $(date)"

if [[ -z "${POD_NAME}" ]]; then
    echo "POD_NAME not provided"
    exit 1
fi

echo
echo "===> Get kubeconfig from address: ${IAM_KUBECONFIG_SVC_FQDN}"

getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"

echo
echo "===> Testing apiserver-proxy functionality"
echo " \=> $(date)"

for (( i = 1; i < (($MAX_TEST_RETRIES + 1)); i++ )); do
    echo
    echo "===> Execution: ${i}/${MAX_TEST_RETRIES}"
    doTest ${i}
done

echo
echo "===> Success: All executions completed"
echo " \=> $(date)"

