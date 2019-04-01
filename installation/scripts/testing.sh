#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

source ${ROOT_PATH}/testing-common.sh

echo "-------------------------------"
echo "- Ensure test Pods are deleted "
echo "-------------------------------"

if [ -n "$KUBE_CONTEXT" ]; then
    echo "Using context: $KUBE_CONTEXT"
    KUBE_CONTEXT_ARG="--kube-context $KUBE_CONTEXT"
fi

# TODO deleting pods?

suiteName="testsuite-all-$(date '+%Y-%m-%d-%H-%M')"
echo "----------------------------"
echo "- Testing Kyma..."
echo "----------------------------"

cat <<EOF | kubectl apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: ${suiteName}
spec:
  maxRetries: 0
  concurrency: 3
  count: 1
EOF

startTime=$(date +%s)




while true
do
    currTime=$(date +%s)
    statusSucceeded=$(kubectl get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}" | grep "True")
    statusFailed=$(kubectl get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}" | grep "True")
    statusError=$(kubectl get cts  ${suiteName} -ojsonpath="{.status.conditions[?(@.type=='Error')]}" | grep "True")

    if [[ "${statusSucceeded}" == "true" ]]; then
        echo "Test succeeded"
        exit 0
    fi

    if [[ "${statusFailed}" == "true" ]]; then
        echo "Test failed"
        kubectl get cts testsuite-all -oyaml
        exit 1
    fi

    if [[ "${statusError}" == "true" ]]; then
        echo "Test errored"
        kubectl get cts testsuite-all -oyaml
        exit 1
    fi
    sec=$((currTime-startTime))
    min=$((sec/60))
    if (( min > 60 )); then
        echo "Timeout occurred"
        exit 1
    fi
    echo "ClusterTestSuite not finished. Waiting..."

    sleep 5

done


#TODO missing printing logs