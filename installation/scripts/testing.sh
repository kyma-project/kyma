#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

source ${ROOT_PATH}/testing-common.sh

echo "-------------------------------"
echo "- Ensure test Pods are deleted "
echo "-------------------------------"

cleanupHelmTestPods kyma-system
cleanupCoreErr=$?

cleanupHelmTestPods istio-system
cleanupIstioErr=$?

cleanupHelmTestPods knative-serving
cleanupKnativeErr=$?

cleanupHelmTestPods kyma-integration
cleanupGatewayErr=$?

if [ ${cleanupGatewayErr} -ne 0 ] || [ ${cleanupIstioErr} -ne 0 ] || [ ${cleanupCoreErr} -ne 0 ] || [ ${cleanupKnativeErr} -ne 0 ]
then
    exit 1
fi

monitoringTestErr=0
loggingTestErr=0
eventBusTestErr=0

echo "----------------------------"
echo "- Testing Kyma..."
echo "----------------------------"

# echo "- Testing Core components..."
# # timeout set to 10 minutes
# helm test core --timeout 600
# coreTestErr=$?

# execute assetstore tests if 'assetstore' is installed
if helm list | grep -q "assetstore"; then
echo "- Testing Asset Store"
helm test assetstore --timeout 600
assetstoreTestErr=$?
fi

# execute monitoring tests if 'monitoring' is installed
if helm list | grep -q "monitoring"; then
echo "- Montitoring module is installed. Running tests for same"
helm test monitoring --timeout 600
monitoringTestErr=$?
fi

# execute logging tests if 'logging' is installed
if helm list | grep -q "logging"; then
echo "- Logging module is installed. Running tests for same"
helm test logging --timeout 600
loggingTestErr=$?
fi

# run event-bus tests if Knative is not installed
if ! kubectl get namespaces | grep -q "knative-eventing"; then
    echo "- Testing Event-Bus..."
    helm test event-bus --timeout 600
    eventBusTestErr=$?
fi

checkAndCleanupTest kyma-system
testCheckCore=$?

echo "- Testing Istio components..."
helm test istio
istioTestErr=$?

checkAndCleanupTest istio-system
testCheckIstio=$?

echo "- Testing Knative components..."
helm test knative
knativeTestErr=$?

checkAndCleanupTest knative-serving
knativeTestErr=$?

echo "- Testing Application Connector"
helm test application-connector --timeout 600
acTestErr=$?

checkAndCleanupTest kyma-integration
testCheckGateway=$?

printImagesWithLatestTag
latestTagsErr=$?

if [ ${latestTagsErr} -ne 0 ] || [ ${coreTestErr} -ne 0 ] || [ ${assetstoreTestErr} -ne 0 ]  || [ ${istioTestErr} -ne 0 ] || [ ${acTestErr} -ne 0 ] || [ ${loggingTestErr} -ne 0 ] || [ ${monitoringTestErr} -ne 0 ] || [ ${knativeTestErr} -ne 0 ] || [ ${eventBusTestErr} -ne 0 ]
then
    exit 1
else
    exit 0
fi
