#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${ROOT_PATH}/utils.sh

function printLogsFromFailedHelmTests() {
    local namespace=$1

    for POD in $(kubectl get pods -n ${namespace} -l helm-chart-test=true --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        log "Testing '${POD}'" nc bold

        phase=$(kubectl get pod ${POD} -n ${namespace} -o jsonpath="{ .status.phase }")

        case ${phase} in
        "Failed")
            log "'${POD}' has Failed status" red
            printLogsFromPod ${namespace} ${POD}
        ;;
        "Running")
            log "'${POD}' failed due to too long Running status" red
            printLogsFromPod ${namespace} ${POD}
        ;;
        "Pending")
            log "'${POD}' failed due to too long Pending status" red
            printf "Fetching events from '${POD}':\n"
            kubectl describe po ${POD} -n ${namespace} | awk 'x==1 {print} /Events:/ {x=1}'
        ;;
        "Unknown")
            log "'${POD}' failed with Unknown status" red
            printLogsFromPod ${namespace} ${POD}
        ;;
        "Succeeded")
            echo "Test of '${POD}' was successful"
            echo "Logs are not displayed after success"
        ;;
        *)
            log "Unknown status of '${POD}' - ${phase}" red
            printLogsFromPod ${namespace} ${POD}
        ;;
        esac
        log "End of testing '${POD}'\n" nc bold
    done
}

function printLogsFromPod() {
    local namespace=$1 pod=$2
    local tailLimit=2000 bytesLimit=500000
    log "Fetching logs from '${pod}' with options tailLimit=${tailLimit} and bytesLimit=${bytesLimit}" nc bold
    result=$(kubectl logs --tail=${tailLimit} --limit-bytes=${bytesLimit} -n ${namespace} ${pod})
    if [ "${#result}" -eq 0 ]; then
        log "FAILED" red
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local namespace=$1

    runningPods=false

    for POD in $(kubectl get pods -n ${namespace} -l helm-chart-test=true --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        phase=$(kubectl get pod "$POD" -n ${namespace} -o jsonpath="{ .status.phase }")
        # A Pod's phase  Failed or Succeeded means pod has terminated.
        # see: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
        if [ "${phase}" !=  "Succeeded" ] && [ "${phase}" != "Failed" ]
        then
          log "Test pod '${POD}' has not terminated, pod phase: ${phase}" red
          runningPods=true
        fi
    done

    if [ ${runningPods} = true ];
    then
        return 1
    fi
}

function checkTestPodLabel() {
    local namespace=$1

    err=false

    log "Test pods should be marked with label 'helm-chart-test=true'. Checking..." nc bold
    for POD in $(kubectl get pods -n ${namespace} --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        annotation=$(kubectl get pod "$POD" -n ${namespace} -o jsonpath="{ .metadata.annotations.helm\.sh/hook }")
        if [ "${annotation}" == "test-success" ] || [ "${annotation}" == "test-failure" ]
        then
            helmLabel=$(kubectl get pod "${POD}" -n ${namespace} -o jsonpath="{ .metadata.labels.helm-chart-test }" )
            if [ "${helmLabel}" != "true" ];
            then
                err=true
                log "Pod ${POD} is wrongly labeled" red
            fi
        fi
    done

    if [ ${err} = true ];
    then
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
}

function cleanupHelmTestPods() {
    local namespace=$1

    log "\nCleaning up helm test pods in namespace ${namespace}" nc bold
    kubectl delete pod -n ${namespace} -l helm-chart-test=true
    deleteErr=$?
    if [ ${deleteErr} -ne 0 ]
    then
      log "FAILED cleaning test pods.\n" red
      return 1
    fi
    log "Success cleaning test pods.\n" nc bold
}

function waitForTestPodsTermination() {
    local retry=0
    local namespace=$1

    log "All test pods should be terminated. Checking..." nc bold
    while [ ${retry} -lt 3 ]; do
        checkTestPodTerminated ${namespace}
        checkTestPodTerminatedErr=$?
        if [ ${checkTestPodTerminatedErr} -ne 0 ]; then
            echo "Waiting for test pods to terminate..."
            sleep 1
        else
            log "OK" green bold
            return 0
        fi
        retry=$[retry + 1]
    done
    log "FAILED" red
    return 1
}

function checkAndCleanupTest() {
    local namespace=$1

    waitForTestPodsTermination ${namespace}
    checkTestPodTerminatedErr=$?

    printLogsFromFailedHelmTests ${namespace}

    checkTestPodLabel ${namespace}
    checkTestPodLabelErr=$?

    cleanupHelmTestPods ${namespace}
    cleanupErr=$?

    if [ ${checkTestPodTerminatedErr} -ne 0 ] || [ ${checkTestPodLabelErr} -ne 0 ] | [ ${cleanupErr} -ne 0 ]
    then
        return 1
    fi
}

function printImagesWithLatestTag() {

    local images=$(kubectl get pods --all-namespaces -o jsonpath="{..image}" |\
    tr -s '[[:space:]]' '\n' |\
    grep ":latest")

    log "Images with tag latest are not allowed. Checking..." nc bold
    if [ ${#images} -ne 0 ]; then
        log "${images}" red
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
    return 0
}

echo "-------------------------------"
echo "- Ensure test Pods are deleted "
echo "-------------------------------"

cleanupHelmTestPods kyma-system
cleanupCoreErr=$?

cleanupHelmTestPods istio-system
cleanupIstioErr=$?

cleanupHelmTestPods kyma-integration
cleanupGatewayErr=$?

if [ ${cleanupGatewayErr} -ne 0 ] || [ ${cleanupIstioErr} -ne 0 ]  || [ ${cleanupCoreErr} -ne 0 ]
then
    exit 1
fi

monitoringTestErr=0
loggingTestErr=0

echo "----------------------------"
echo "- Testing Kyma..."
echo "----------------------------"

echo "- Testing Core components..."
# timeout set to 10 minutes
helm test core --timeout 600
coreTestErr=$?

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

checkAndCleanupTest kyma-system
testCheckCore=$?

echo "- Testing Istio components..."
helm test istio
istioTestErr=$?

checkAndCleanupTest istio-system
testCheckIstio=$?

echo "- Testing Application Connector"
helm test application-connector
acTestErr=$?

checkAndCleanupTest kyma-integration
testCheckGateway=$?

printImagesWithLatestTag
latestTagsErr=$?

if [ ${latestTagsErr} -ne 0 ] || [ ${coreTestErr} -ne 0 ]  || [ ${istioTestErr} -ne 0 ] || [ ${acTestErr} -ne 0 ] || [ ${loggingTestErr} -ne 0 ] || [ ${monitoringTestErr} -ne 0 ]
then
    exit 1
else
    exit 0
fi
