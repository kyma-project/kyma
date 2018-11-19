#!/usr/bin/env bash

set -o errexit # exit immediately if a command exits with a non-zero status.
set -o nounset # exit when script tries to use undeclared variables

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
WORKING_NAMESPACE="kyma-system"

echo "Run stability test..."
kubectl create configmap pod-watch-config -n ${WORKING_NAMESPACE} --from-literal="ARGS=-maxWaitingPeriod=10m -ignorePodsPattern=core-azure-broker-docs-*|kyma-release-adder-*" >/dev/null
kubectl apply -f ${CURRENT_DIR}/../resources/watch-pods.yaml >/dev/null
${CURRENT_DIR}/../scripts/is-ready.sh ${WORKING_NAMESPACE} app watch-pods

kubectl logs pod/watch-pods --timestamps --follow -n ${WORKING_NAMESPACE}
STATUS="$(kubectl get pods watch-pods -o jsonpath='{.status.containerStatuses[*].lastState.terminated.exitCode}' -n ${WORKING_NAMESPACE})"
echo "Resulting exit status of stability test is: $STATUS"

kubectl delete -f ${CURRENT_DIR}/../resources/watch-pods.yaml >/dev/null
kubectl delete configmap pod-watch-config -n ${WORKING_NAMESPACE} >/dev/null

exit ${STATUS}
