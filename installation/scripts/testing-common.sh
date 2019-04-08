ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${ROOT_PATH}/utils.sh

function context_arg() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "--context $KUBE_CONTEXT"
    fi
}

function printLogsFromFailedHelmTests() {
    local namespace=$1

    for POD in $(kubectl $(context_arg)  get pods -n ${namespace} -l helm-chart-test=true --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        log "Testing '${POD}'" nc bold

        phase=$(kubectl $(context_arg)  get pod ${POD} -n ${namespace} -o jsonpath="{ .status.phase }")

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
            kubectl $(context_arg)  describe po ${POD} -n ${namespace} | awk 'x==1 {print} /Events:/ {x=1}'
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

function getContainerFromPod() {
    local namespace="$1"
    local pod="$2"
    local containers2ignore="istio-init istio-proxy manager"
    containersInPod=$(kubectl get pods ${pod} -o jsonpath='{.spec.containers[*].name}' -n ${namespace})
    for container in $containersInPod; do
        if [[ ! ${containers2ignore[*]} =~ "${container}" ]]; then
            echo "${container}"
        fi
    done
}

function printLogsFromPod() {
    local namespace=$1 pod=$2
    local tailLimit=2000 bytesLimit=500000
    log "Fetching logs from '${pod}' with options tailLimit=${tailLimit} and bytesLimit=${bytesLimit}" nc bold
    testPod=$(getContainerFromPod ${namespace} ${pod})
    echo "testPod = $testPod"
    result=$(kubectl $(context_arg)  logs --tail=${tailLimit} --limit-bytes=${bytesLimit} -n ${namespace} ${testPod})
    if [ "${#result}" -eq 0 ]; then
        log "FAILED" red
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local namespace=$1

    runningPods=false

    for POD in $(kubectl $(context_arg)  get pods -n ${namespace} -l helm-chart-test=true --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        phase=$(kubectl $(context_arg)  get pod "$POD" -n ${namespace} -o jsonpath="{ .status.phase }")
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
    for POD in $(kubectl $(context_arg)  get pods -n ${namespace} --show-all -o jsonpath='{.items[*].metadata.name}')
    do
        annotation=$(kubectl $(context_arg)  get pod "$POD" -n ${namespace} -o jsonpath="{ .metadata.annotations.helm\.sh/hook }")
        if [ "${annotation}" == "test-success" ] || [ "${annotation}" == "test-failure" ]
        then
            helmLabel=$(kubectl $(context_arg)  get pod "${POD}" -n ${namespace} -o jsonpath="{ .metadata.labels.helm-chart-test }" )
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
    kubectl $(context_arg)  delete pod -n ${namespace} -l helm-chart-test=true
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

    local images=$(kubectl $(context_arg)  get pods --all-namespaces -o jsonpath="{..image}" |\
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
