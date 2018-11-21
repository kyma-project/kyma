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

    log "Fetching logs from '${pod}'" nc bold
    result=$(kubectl logs -n ${namespace} ${pod})
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

    log "\nCleaning up helm test pods" nc bold
    kubectl delete pod -n ${namespace} -l helm-chart-test=true
    log "End of cleaning test pods.\n" nc bold
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

    if [ ${checkTestPodTerminatedErr} -ne 0 ] || [ ${checkTestPodLabelErr} -ne 0 ]
    then
        return 1
    fi
}

function printImagesWithLatestTag(){

    # We ignore the alpine image as this is required by istio-sidecar
    local images=$(kubectl get pods --all-namespaces -o jsonpath="{..image}" |\
    tr -s '[[:space:]]' '\n' |\
    grep ":latest" | grep -v "alpine:latest")

    log "Images with tag latest are not allowed. Checking..." nc bold
    if [ ${#images} -ne 0 ]; then
        log "${images}" red
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
    return 0
}



echo "----------------------------"
echo "- Testing Kyma ..."
echo "----------------------------"

echo "- Testing Core components..."
# timeout set to 10 minutes
helm test core --timeout 600
coreTestErr=$?

echo "- Testing Logging components..."
helm test logging
loggingTestErr=$?

checkAndCleanupTest kyma-system
testCheckKymaCore=$?

echo "- Testing Istio components..."
helm test istio
istioTestErr=$?

checkAndCleanupTest istio-system
testCheckIstio=$?

echo "- Testing Remote Environments"
helm test ec-default
ecTestErr=$?
helm test hmc-default
hmcTestErr=$?

checkAndCleanupTest kyma-integration
testCheckGateway=$?

printImagesWithLatestTag
latestTagsErr=$?

if  [ ${coreTestErr} -ne 0 ] || [ ${testCheckKymaCore} -ne 0 ] || [ ${istioTestErr} -ne 0 ] ||
    [ ${testCheckIstio} -ne 0 ] || [ ${ecTestErr} -ne 0 ] || [ ${hmcTestErr} -ne 0 ] ||
    [ ${testCheckGateway} -ne 0 ] || [ ${latestTagsErr} -ne 0 ] || [ ${loggingTestErr} -ne 0 ]

then
    exit 1
else
    exit 0
fi