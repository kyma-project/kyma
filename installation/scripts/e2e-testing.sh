#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${ROOT_PATH}/utils.sh"
source ${ROOT_PATH}/testing-common.sh
#copied from  testing-common.sh: in testing-common.sh we use Octopus, here helm test. TODO later: rewrite e2e-testing to Octopus.

function context_arg() {
    if [ -n "$KUBE_CONTEXT" ]; then
        echo "--context $KUBE_CONTEXT"
    fi
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

function printLogsFromPod() {
    local namespace=$1 pod=$2
    local tailLimit=2000 bytesLimit=500000
    log "Fetching logs from '${pod}' with options tailLimit=${tailLimit} and bytesLimit=${bytesLimit}" nc bold
    container=$(getContainerFromPod ${namespace} ${pod})
    result=$(kubectl $(context_arg) logs --tail=${tailLimit} --limit-bytes=${bytesLimit} -n ${namespace} -c ${container} ${pod})
    if [ "${#result}" -eq 0 ]; then
        log "FAILED" red
        return 1
    fi
    echo "${result}"
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

#end of copy


if [ -z "${DOMAIN}" ] ; then
  echo "ERROR: DOMAIN is not set"
  exit 1
fi

cleanupHelmE2ERelease () {
    local release=$1
    log 'Running cleanup'
    helm del --purge "${release}" --tls
    local deleteErr=$?
    if [ ${deleteErr} -ne 0 ]
    then
      log "FAILED cleaning release.\n" red
      return 1
    fi
    while helm list --deleting --tls 2>/dev/null | grep "${release}" ; do
        sleep 1
        echo .
    done
}


echo "-------------------------------"
echo "- Ensure test Pods are deleted "
echo "-------------------------------"

cleanupHelmTestPods end-to-end
cleanupE2EErr=$?

if [ ${cleanupE2EErr} -ne 0 ]
then
    exit 1
fi

echo "----------------------------"
echo "- E2E Testing Kyma..."
echo "----------------------------"

exitCode=0


testcase="${ROOT_PATH}"/../../tests/end-to-end/backup-restore-test/deploy/chart/backup-test
release=$(basename "$testcase")

cleanupHelmE2ERelease "${release}"

ADMIN_EMAIL=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode)
ADMIN_PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)

helm install "$testcase" --name "${release}" --namespace end-to-end --set global.ingress.domainName="${DOMAIN}" --set-file global.adminEmail=<(echo -n "${ADMIN_EMAIL}") --set-file global.adminPassword=<(echo -n "${ADMIN_PASSWORD}") --tls
helm test "${release}" --timeout 10000 --tls
testResult=$?
if [ $testResult -eq 0 ]
then
    releasesToClean="$releasesToClean ${release}"
else
    exitCode=$testResult
fi

checkAndCleanupTest end-to-end
cleanupResult=$?
if [ $cleanupResult -ne 0 ]
then
   exitCode=$cleanupResult
fi


for release in $releasesToClean; do
    cleanupHelmE2ERelease "${release}"
    cleanupResult=$?
    if [ $cleanupResult -ne 0 ]
    then
        exitCode=$cleanupResult
    fi
done

if [ ${exitCode} -ne 0 ]
then
    log FAIL red
    exit 1
else
    log SUCCESS green
    exit 0
fi
