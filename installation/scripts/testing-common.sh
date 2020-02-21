ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${ROOT_PATH}/utils.sh

function context_arg() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "--context $KUBE_CONTEXT"
    fi
}

# retries are useful when api call can fail due to the infrastructure issue
function executeKubectlWithRetries() {
    local command="$1"
    local retry=0
    local result=""

    while [[ ${retry} -lt 10 ]]; do
        result=$(${command})
        if [[ $? -eq 0 ]]; then
            echo "${result}"
            return 0
        else
            sleep 5
        fi
        (( retry++ ))
    done
    echo "Maximum retries exceeded: ${result}"
    return 1
}

function cmdGetPodsForSuite() {
    local suiteName=$1
    cmd="kubectl $(context_arg) get pods -l testing.kyma-project.io/suite-name=${suiteName} \
            --all-namespaces \
            --no-headers=true \
            -o=custom-columns=name:metadata.name,ns:metadata.namespace"
    result=$(executeKubectlWithRetries "${cmd}")
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi
    echo "${result}"
}

function printLogsFromFailedTests() {
    local suiteName=$1
    local result=$(cmdGetPodsForSuite ${suiteName})
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi

    pod=""
    namespace=""
    idx=0
    for podOrNs in ${result}
    do
        n=$((idx%2))
         if [[ "$n" == 0 ]];then
            pod=${podOrNs}
            idx=$((${idx}+1))
            continue
        fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        log "Testing '${pod}' from namespace '${namespace}'" nc bold

        phase=$(executeKubectlWithRetries "kubectl $(context_arg) get pod ${pod} -n ${namespace} -o jsonpath={.status.phase}")
        if [[ $? -eq 1 ]]; then
            echo "${phase}"
            return 1
        fi

        case ${phase} in
        "Failed")
            log "'${pod}' has Failed status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Running")
            log "'${pod}' failed due to too long Running status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Pending")
            log "'${pod}' failed due to too long Pending status" red
            printf "Fetching events from '${pod}':\n"
            result=$(executeKubectlWithRetries "kubectl $(context_arg)  describe po ${pod} -n ${namespace} | awk 'x==1 {print} /Events:/ {x=1}'")
            echo "${result}"
            if [[ $? -eq 1 ]]; then
                return 1
            fi
        ;;
        "Unknown")
            log "'${pod}' failed with Unknown status" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        "Succeeded")
            # do nothing
        ;;
        *)
            log "Unknown status of '${pod}' - ${phase}" red
            printLogsFromPod ${namespace} ${pod}
        ;;
        esac
        log "End of testing '${pod}'\n" nc bold

    done
}

function getContainerFromPod() {
    local namespace="$1"
    local pod="$2"
    local containers2ignore="istio-init istio-proxy manager"

    result=$(executeKubectlWithRetries "kubectl get pods ${pod} -o jsonpath={.spec.containers[*].name} -n ${namespace}")
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi
    for cnt in ${result}; do
        if [[ ! ${containers2ignore[*]} =~ "${cnt}" ]]; then
            echo "${cnt}"
            return 0
        fi
    done
}

function printLogsFromPod() {
    local namespace=$1 pod=$2
    local tailLimit=2000 bytesLimit=500000
    log "Fetching logs from '${pod}' with options tailLimit=${tailLimit} and bytesLimit=${bytesLimit}" nc bold
    container=$(getContainerFromPod ${namespace} ${pod})
    if [ $? -eq 1 ]; then
        echo "${container}"
        return 1
    fi
    result=$(executeKubectlWithRetries "kubectl $(context_arg)  logs --tail=${tailLimit} --limit-bytes=${bytesLimit} -n ${namespace} ${pod} -c ${container}")
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local suiteName=$1
    runningPods=false

    pod=""
    namespace=""
    idx=0

    local result=$(cmdGetPodsForSuite ${suiteName})
    if [[ $? -eq 1 ]]; then
        echo "${result}"
        return 1
    fi

    for podOrNs in ${result}
    do
       n=$((idx%2))
       if [[ "$n" == 0 ]];then
         pod=${podOrNs}
         idx=$((${idx}+1))
         continue
       fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        phase=$(executeKubectlWithRetries "kubectl $(context_arg) get pod $pod -n ${namespace} -o jsonpath={.status.phase}")
        if [[ $? -eq 1 ]]; then
            echo "${phase}"
            return 1
        fi
        # A Pod's phase  Failed or Succeeded means pod has terminated.
        # see: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
        if [ "${phase}" !=  "Succeeded" ] && [ "${phase}" != "Failed" ]
        then
          log "Test pod '${pod}' has not terminated, pod phase: ${phase}" red
          runningPods=true
        fi
    done

    if [ ${runningPods} = true ];
    then
        return 1
    fi
}

function waitForTestPodsTermination() {
    local retry=0
    local suiteName=$1

    log "All test pods should be terminated. Checking..." nc bold
    while [ ${retry} -lt 3 ]; do
        checkTestPodTerminated ${suiteName}
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

function deleteCTS() {
  local retry=0
  local suiteName=$1

  log "Deleting ClusterTestSuite ${suiteName}" nc bold
  while [[ ${retry} -lt 20 ]]; do
        msg=$(kubectl delete cts ${suiteName} 2>&1)
        status=$?
        if [[ ${status} -ne 0 ]]; then
            echo "Unable to delete ClusterTestSuite: ${msg}"
            echo "waiting 5s..."
            sleep 5
        else
            log "OK" green bold
            return 0
        fi
        retry=$[retry + 1]
  done
  log "FAILED" red
  return 1
}

function waitForTerminationAndPrintLogs() {
    local suiteName=$1

    waitForTestPodsTermination ${suiteName}
    checkTestPodTerminatedErr=$?

    printLogsFromFailedTests ${suiteName}
    if [ ${checkTestPodTerminatedErr} -ne 0 ]
    then
        return 1
    fi
}

function waitForTestSuiteResult() {
    local suiteName=$1

    kc="kubectl $(context_arg)"

    startTime=$(date +%s)
    testExitCode=0
    previousPrintTime=-1

while true
do
    currTime=$(date +%s)
    statusSucceeded=$(${kc} get cts "${suiteName}"  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(${kc} get cts "${suiteName}"  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(${kc} get cts  "${suiteName}" -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

    if [[ "${statusSucceeded}" == *"True"* ]]; then
       echo "Test suite '${suiteName}' succeeded."
       break
    fi

    if [[ "${statusFailed}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' failed."
        testExitCode=1
        break
    fi

    if [[ "${statusError}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' errored."
        testExitCode=1
        break
    fi

    sec=$((currTime-startTime))
    min=$((sec/60))
    if (( min > 60 )); then
        echo "Timeout for test suite '${suiteName}' occurred."
        testExitCode=1
        break
    fi
    if (( previousPrintTime != min )); then
        echo "ClusterTestSuite not finished. Waiting..."
        previousPrintTime=${min}
    fi
    sleep 3
done

    echo "Test summary"
    kubectl get cts  ${suiteName} -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

    if [[ ${testExitCode} -eq 1 ]]; then
        waitForTerminationAndPrintLogs ${suiteName}
    fi

    echo "ClusterTestSuite details:"
    kubectl get cts ${suiteName} -oyaml

    return ${testExitCode}
}

function printImagesWithLatestTag() {
    retry=10
    while true; do
        local images=$(kubectl $(context_arg)  get pods --all-namespaces -o jsonpath="{..image}" |\
        tr -s '[[:space:]]' '\n' |\
        grep ":latest")
        if [[ $? -eq 0 ]]; then
            break
        fi
        (( retry-- ))
        if [[ ${retry} -eq 0 ]]; then
        return 1
        fi
        sleep 5
    done

    log "Images with tag latest are not allowed. Checking..." nc bold
    if [ ${#images} -ne 0 ]; then
        log "${images}" red
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
    return 0
}

TESTING_ADDONS_CFG_NAME="testing-addons"

function injectTestingAddons() {
    retry=10
    while true; do
        cat <<EOF | kubectl apply -f -
apiVersion: addons.kyma-project.io/v1alpha1
kind: ClusterAddonsConfiguration
metadata:
  labels:
    addons.kyma-project.io/managed: "true"
  name: ${TESTING_ADDONS_CFG_NAME}
spec:
  repositories:
  - url: "https://github.com/kyma-project/addons/releases/download/0.9.0/index-testing.yaml"
EOF
        if [[ $? -eq 0 ]]; then
            break
        fi
        (( retry-- ))
        if [[ ${retry} -eq 0 ]]; then
            return 1
        fi
        sleep 5
    done

    local retry=0
    while [[ ${retry} -lt 10 ]]; do
        msg=$(kubectl get clusteraddonsconfiguration ${TESTING_ADDONS_CFG_NAME} -o=jsonpath='{.status.phase}')
        if [[ "${msg}" = "Ready" ]]; then
            log "Testing addons injected" green
            return 0
        fi
        if [[ "${msg}" = "Failed" ]]; then
            log "Testing addons configuration failed" red
            removeTestingAddons
            return 1
        fi
        echo "Waiting for ready testing addons ${retry}/10.. status: ${msg}"
        retry=$[retry + 1]
        sleep 3
    done
    log "Testing addons couldn't be injected" red
    return 1
}

function removeTestingAddons() {
    result=$(executeKubectlWithRetries "kubectl delete clusteraddonsconfiguration ${TESTING_ADDONS_CFG_NAME}")
    echo "${result}"
    if [[ $? -eq 1 ]]; then
        return 1
    fi
    log "Testing addons removed" green
}
