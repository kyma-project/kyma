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

cleanupHelmTestPods kyma-system
cleanupCoreErr=$?

cleanupHelmTestPods istio-system
cleanupIstioErr=$?

cleanupHelmTestPods knative-serving
cleanupKnativeServingErr=$?

cleanupHelmTestPods kyma-integration
cleanupGatewayErr=$?

if [ ${cleanupGatewayErr} -ne 0 ] || [ ${cleanupIstioErr} -ne 0 ] || [ ${cleanupCoreErr} -ne 0 ] || [ ${cleanupKnativeServingErr} -ne 0 ]
then
    exit 1
fi

monitoringTestErr=0
loggingTestErr=0
eventBusTestErr=0

echo "----------------------------"
echo "- Testing Kyma..."
echo "----------------------------"

echo "- Testing Core components..."
# timeout set to 10 minutes
helm ${KUBE_CONTEXT_ARG} test core --timeout 600 --tls
coreTestErr=$?

# execute assetstore tests if 'assetstore' is installed
if helm ${KUBE_CONTEXT_ARG} list --tls | grep -q "assetstore"; then
echo "- Testing Asset Store"
helm ${KUBE_CONTEXT_ARG} test assetstore --timeout 600 --tls
assetstoreTestErr=$?

echo "=== CLUSTER DOCS TOPIKI CTRL MANAGER LOGS === "
kubectl logs assetstore-asset-store-controller-manager-0 -n kyma-system
echo "==="
fi

# execute monitoring tests if 'monitoring' is installed
if helm ${KUBE_CONTEXT_ARG} list --tls | grep -q "monitoring"; then
echo "- Montitoring module is installed. Running tests for same"
helm ${KUBE_CONTEXT_ARG} test monitoring --timeout 600 --tls
monitoringTestErr=$?
fi

# execute logging tests if 'logging' is installed
#if helm ${KUBE_CONTEXT_ARG} list --tls | grep -q "logging"; then
#echo "- Logging module is installed. Running tests for same"
#helm ${KUBE_CONTEXT_ARG} test logging --timeout 600
#loggingTestErr=$?
#fi

# run event-bus tests if Knative is installed
if kubectl -n knative-eventing get deployments.apps | grep -q "webhook"; then
    echo "- Testing Event-Bus..."
    helm ${KUBE_CONTEXT_ARG} test event-bus --timeout 600 --tls
    eventBusTestErr=$?
fi

checkAndCleanupTest kyma-system
testCheckCore=$?

echo "- Testing Istio components..."
helm ${KUBE_CONTEXT_ARG} test istio --tls
istioTestErr=$?

checkAndCleanupTest istio-system
testCheckIstio=$?

echo "- Testing Knative serving components..."
helm ${KUBE_CONTEXT_ARG} test knative-serving --tls
knativeServingTestErr=$?

checkAndCleanupTest knative-serving
knativeServingTestErr=$?

echo "- Testing Application Connector"
helm ${KUBE_CONTEXT_ARG} test application-connector --timeout 600 --tls
acTestErr=$?

checkAndCleanupTest kyma-integration
testCheckGateway=$?

printImagesWithLatestTag
latestTagsErr=$?

if [ ${latestTagsErr} -ne 0 ] || [ ${coreTestErr} -ne 0 ] || [ ${assetstoreTestErr} -ne 0 ]  || [ ${istioTestErr} -ne 0 ] || [ ${acTestErr} -ne 0 ] || [ ${loggingTestErr} -ne 0 ] || [ ${monitoringTestErr} -ne 0 ] || [ ${knativeServingTestErr} -ne 0 ] || [ ${eventBusTestErr} -ne 0 ]
then
    exit 1
else
    exit 0
fi
