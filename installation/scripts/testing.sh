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
  concurrency: 1
  count: 1
EOF

startTime=$(date +%s)

while true
do
    currTime=$(date +%s)
    statusSucceeded=$(kubectl get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(kubectl get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(kubectl get cts  ${suiteName} -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

    if [[ "${statusSucceeded}" == *"True"* ]]; then
        echo "Test succeeded"
        exit 0
    fi

    if [[ "${statusFailed}" == *"True"* ]]; then
        echo "Test failed. Details: "
        kubectl get cts ${suiteName} -oyaml
        exit 1
    fi

    if [[ "${statusError}" == *"True"* ]]; then
        echo "Test errored. Details"
        kubectl get cts ${suiteName} -oyaml
        exit 1
    fi
    sec=$((currTime-startTime))
    min=$((sec/60))
    if (( min > 60 )); then
        echo "Timeout occurred. Current state of the test suite:"
        kubectl get cts ${suiteName} -oyaml
        exit 1
    fi
    echo "ClusterTestSuite not finished. Waiting..."

    sleep 5

done


#TODO missing printing logs